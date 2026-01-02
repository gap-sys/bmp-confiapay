package queue

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RedisPublisher interface {
	Publish(payload models.UpdateDbData) error
}

// producer é a struct que contém as configurações e canais do producer.
type producer struct {
	ctx              context.Context
	conn             *amqp.Connection
	dlqCh            *amqp.Channel
	simulacaoDlqCh   *amqp.Channel
	dbCh             *amqp.Channel
	dadosBancariosCh *amqp.Channel
	webhookCh        *amqp.Channel
	dbQueue          string
	dlqExchange      string
	exchange         string
	dlqQueue         string
	webhookQueue     string
	logger           *slog.Logger
	location         *time.Location
	redisPublisher   RedisPublisher
}

// setUpProducers configura os producers para as filas digitacaoDlq, simulacaoDlq, dados bancarios e webhook.
func setUpProducers(rmq *RabbitMQ) {
	var baseProducer = producer{

		ctx:            rmq.ctx,
		conn:           rmq.conn,
		dlqCh:          rmq.DlqCh,
		dlqQueue:       rmq.dlqQueue,
		dbCh:           rmq.dbCh,
		webhookCh:      rmq.webhookCh,
		dbQueue:        rmq.dbqueue,
		dlqExchange:    rmq.dlqExchange,
		webhookQueue:   rmq.webhookQueue,
		logger:         rmq.logger,
		location:       rmq.location,
		redisPublisher: rmq.cache,
		exchange:       rmq.exchange,
	}

	rmq.baseProducer = &baseProducer

	setUpCobrancaProducer(rmq, baseProducer)

}

// Produce produz uma mensagem na fila especificada.
// queue é a fila onde a mensagem será produzida.
// data é o dado a ser produzido.
// delay é o delay em milissegundos.
func (p *producer) Produce(queue string, data any, delay time.Duration) error {
	body, err := json.Marshal(data)
	if err != nil {
		var dlqData = models.DLQData{Payload: data, Mensagem: "Erro ao serializar json no producer", Erro: err.Error(), Time: time.Now().In(p.location)}
		helpers.LogError(p.ctx, p.logger, p.location, "producer", "", "erro ao serializar json", err.Error(), nil)
		p.Produce(p.dlqQueue, dlqData, 0)

		return err
	}

	headers := amqp.Table{
		"x-delay": int32(delay.Milliseconds()),
	}

	msg := amqp.Publishing{
		Body:         body,
		DeliveryMode: 2,
		Headers:      headers,
	}
	switch queue {

	case p.webhookQueue:
		err := p.webhookCh.Publish(p.exchange, queue, false, false, msg)
		if err != nil {
			helpers.LogError(p.ctx, p.logger, p.location, "producer webhook", "", "erro ao gravar mensagem na fila de webhook", err.Error(), nil)
			var dlqData = models.DLQData{Payload: data, Contexto: "producer webhook", Erro: err.Error(), Mensagem: "erro ao gravar mensagem na fila de webhook", Time: time.Now().In(p.location)}
			p.Produce(p.dlqQueue, dlqData, 0)
			return err
		}
		return nil

	case p.dbQueue:
		for try := range config.DB_UPDATE_MAX_RETRIES {
			if err := p.dbCh.Publish(p.exchange, queue, false, false, msg); err == nil {
				return nil
			} else {
				errMsg := fmt.Sprintf("erro ao gravar mensagem na fila de atualização, tentativa %d", try+1)
				helpers.LogError(p.ctx, p.logger, p.location, "producer atualização", "", errMsg, err.Error(), nil)
			}

		}
		redisPayload := data.(models.UpdateDbData)
		p.redisPublisher.Publish(redisPayload)
		return err

	case p.dlqQueue:
		msg.Headers = amqp.Table{}
		err := p.dlqCh.Publish(p.dlqExchange, queue, false, false, msg)
		if err != nil {
			helpers.LogError(p.ctx, p.logger, p.location, "producer dlq", "", "erro ao gravar mensagem na fila de dlq de digitacao", err.Error(), nil)
			return err
		}

	default:
		err := errors.New("invalid queue name")
		helpers.LogError(p.ctx, p.logger, p.location, "producer", "", "fila inválida", err.Error(), map[string]string{"fila": queue})
		return err

	}

	return nil

}
