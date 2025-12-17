package queue

import (
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// CobrancaProducer é a struct que contém as configurações e canais do producer CreditoPessoal.
type CobrancaProducer struct {
	producer
	digitacaoCreditoPessoalCh       *amqp.Channel
	simulacaoCreditoPessoalCh       *amqp.Channel
	webhookCh                       *amqp.Channel
	simulacaoCreditoPessoalQueue    string
	digitacaoCreditoPessoalQueue    string
	dbQueue                         string
	cobrancaCh                      *amqp.Channel
	cancelamentoCreditoPessoalCh    *amqp.Channel
	liberacaoCreditoPessoalCh       *amqp.Channel
	cobrancaQueue                   string
	liberacaoCreditoPessoalQueue    string
	cancelamentoCreditoPessoalQueue string
}

// setUpCobrancaProducer configura os producers para as filas digitacaoCreditoPessoal, simulacaoCreditoPessoal, cobrancaCreditoPessoal, cancelamentoCreditoPessoal, liberacaoCreditoPessoal, db e webhook.
func setUpCobrancaProducer(rmq *RabbitMQ, p producer) {

	rmq.cobrancaProducer = &CobrancaProducer{
		producer: p,

		dbQueue:       rmq.dbqueue,
		webhookCh:     rmq.webhookCh,
		cobrancaCh:    rmq.cobrancaCh,
		cobrancaQueue: rmq.cobrancaQueue,
	}

}

// Produce produz uma mensagem na fila especificada.
// queue é a fila onde a mensagem será produzida.
// data é o dado a ser produzido.
// delay é o delay em milissegundos.
func (c *CobrancaProducer) Produce(queue string, data any, delay time.Duration) error {
	body, err := json.Marshal(data)
	// Se houver um erro ao serializar o json, envia para a DLQ
	if err != nil {
		var dlqData = models.DLQData{Payload: data, Mensagem: "Erro ao serializar json no producer", Erro: err.Error(), Time: time.Now().In(c.location)}
		helpers.LogError(c.ctx, c.logger, c.location, "producer", "", "erro ao serializar json", err.Error(), nil)
		switch queue {
		case c.cobrancaQueue:
			dlqData.Contexto = "producer cobrança"
			c.Produce(c.dlqQueue, dlqData, 0)

		case c.webhookQueue:
			dlqData.Contexto = "producer webhook"
			c.Produce(c.dlqQueue, dlqData, 0)

		}

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

	case c.dlqQueue:
		msg.Headers = amqp.Table{}
		err := c.dlqCh.PublishWithContext(c.ctx, c.dlqExchange, queue, false, false, msg)
		if err != nil {
			helpers.LogError(c.ctx, c.logger, c.location, "producer dlq", "", "erro ao gravar mensagem na fila de dlq", err.Error(), nil)
			return err
		}
		return nil

	case c.cobrancaQueue:
		err := c.cobrancaCh.PublishWithContext(c.ctx, c.exchange, queue, false, false, msg)
		if err != nil {
			helpers.LogError(c.ctx, c.logger, c.location, "producer cobrança", "", "erro ao gravar mensagem na fila de cobrança", err.Error(), nil)
			var dlqData = models.DLQData{Payload: data, Contexto: "producer cobrança", Erro: err.Error(), Mensagem: "erro ao gravar mensagem na fila de cobrança", Time: time.Now().In(c.location)}
			c.Produce(c.dlqQueue, dlqData, 0)
			return err
		}
		return nil

	case c.webhookQueue:
		err := c.webhookCh.PublishWithContext(c.ctx, c.exchange, queue, false, false, msg)
		if err != nil {
			helpers.LogError(c.ctx, c.logger, c.location, "producer webhook", "", "erro ao gravar mensagem na fila de webhook", err.Error(), nil)
			var dlqData = models.DLQData{Payload: data, Contexto: "producer webhook", Erro: err.Error(), Mensagem: "erro ao gravar mensagem na fila de webhook", Time: time.Now().In(c.location)}
			c.Produce(c.dlqQueue, dlqData, 0)
			return err
		}
		return nil
		/*
			case c.dbQueue:
				for try := range config.DB_UPDATE_MAX_RETRIES {
					if err := c.dbCh.PublishWithContext(c.ctx, c.exchange, queue, false, false, msg); err == nil {
						return nil
					} else {
						errMsg := fmt.Sprintf("erro ao gravar mensagem na fila de atualização, tentativa %d", try+1)
						helpers.LogError(c.ctx, c.logger, c.location, "producer atualização", "", errMsg, err.Error(), nil)
					}

				}
				redisPayload := data.(models.UpdateDBData)
				c.redisPublisher.Publish(redisPayload)
				return err
		*/
	default:
		err := errors.New("invalid queue name")
		helpers.LogError(c.ctx, c.logger, c.location, "producer", "", "fila inválida", err.Error(), map[string]string{"fila": queue})
		return err

	}

}
