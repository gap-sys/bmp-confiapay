package queue

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ é a struct que contém as configurações e canais do RabbitMQ.

type RabbitMQ struct {
	ctx                        context.Context
	conn                       *amqp.Connection
	exchange                   string
	dlqExchange                string
	cobrancaCh                 *amqp.Channel
	webhookCh                  *amqp.Channel
	dbCh                       *amqp.Channel
	DlqCh                      *amqp.Channel
	webhookMsgs                <-chan amqp.Delivery
	dbMsgs                     <-chan amqp.Delivery
	cobrancaCreditoPessoalMsgs <-chan amqp.Delivery
	cobrancaQueue              string
	webhookQueue               string
	dbqueue                    string
	dlqQueue                   string
	dbService                  dbService
	cobrancaService            CobrancaCreditoPessoalService
	webhookService             WebhookService
	QOS                        int
	globalQOS                  bool
	url                        string
	errchan                    chan *amqp.Error
	cache                      RedisPublisher
	logger                     *slog.Logger
	location                   *time.Location
	baseProducer               *producer
	cobrancaProducer           *CobrancaProducer
	mu                         *sync.Mutex
}

// NewRMQ cria uma nova instância de RabbitMQ.
// ctx é o contexto de cancelamento.
// loc é o fuso horário a ser utilizado.
// cache é o cache a ser utilizado.
// logger é o logger a ser utilizado.
// simulacaoFGTSService é o serviço de simulação FGTS.
// digitacaoFGTSService é o serviço de digitação FGTS.
// cobrancaFGTSService é o serviço de cobrança FGTS.
// simulacaoINSSCPService é o serviço de simulação INSSCP.
// digitacaoINSSCPService é o serviço de digitação INSSCP.
// dbService é o serviço de banco de dados.
// webhookService é o serviço de webhook.
// cobrancaINSSCPService é o serviço de cobrança INSSCP.
// digitacaoQOS é o QOS de digitação.
// simulacaoQOS é o QOS de simulação.
// globalQOS é se o QOS é global.
func NewRMQ(ctx context.Context, loc *time.Location, cache RedisPublisher,
	logger *slog.Logger,
	webhookService WebhookService,
	cobrancaService CobrancaCreditoPessoalService, dbService dbService,
	QOS int) (*RabbitMQ, error) {

	var rmq = &RabbitMQ{
		ctx:             ctx,
		cache:           cache,
		location:        loc,
		errchan:         make(chan *amqp.Error),
		cobrancaQueue:   config.COBRANCA_QUEUE,
		dlqQueue:        config.DLQ_QUEUE,
		dbqueue:         config.DB_QUEUE,
		webhookQueue:    config.WEBHOOK_QUEUE,
		url:             config.RABBITMQ_URL,
		exchange:        config.BMP_EXCHANGE,
		dlqExchange:     config.DLQ_EXCHANGE,
		webhookService:  webhookService,
		cobrancaService: cobrancaService,
		dbService:       dbService,
		QOS:             QOS,
		logger:          logger,
		mu:              &sync.Mutex{},
	}

	//Se conectando ao RabbitMQ
	conn, err := amqp.Dial(rmq.url)
	if err != nil {
		return nil, err
	}

	//Configurando o canal que notificará caso a conexão seja fechada por algum motivo.
	conn.NotifyClose(rmq.errchan)
	rmq.conn = conn

	//Configurando um canal para cada fila
	if err := rmq.setUpChannels(); err != nil {
		return nil, err
	}

	//Declarando as filas
	if err := rmq.declareQueues(); err != nil {
		return nil, err
	}

	//Configurando os consumers
	if err := rmq.setUpConsumers(); err != nil {
		return nil, err
	}

	//Configurando os producers
	setUpProducers(rmq)
	if err := rmq.injectProducer(); err != nil {
		return nil, err
	}

	return rmq, nil
}

// Injetando o Producer nos services, para que eles possam gravar mensagens nas filas.
func (r *RabbitMQ) injectProducer() error {

	if err := r.dbService.SetProducer(r.baseProducer); err != nil {
		return err
	}

	if err := r.webhookService.SetProducer(r.baseProducer); err != nil {
		return err
	}

	if err := r.cobrancaService.SetProducer(r.cobrancaProducer); err != nil {
		return err
	}
	return nil
}

// Configurando os canais
func (r *RabbitMQ) setUpChannels() error {

	err := r.setUpCobrancaChannels()
	if err != nil {
		return err
	}

	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	r.dbCh = ch

	ch, err = r.conn.Channel()
	if err != nil {
		return err
	}
	r.DlqCh = ch

	ch, err = r.conn.Channel()
	if err != nil {
		return err
	}
	r.webhookCh = ch

	return nil
}

// Declara as filas
func (r *RabbitMQ) declareQueues() error {

	err := r.cobrancaCh.ExchangeDeclare(r.exchange, "x-delayed-message",
		true, false, false, false,
		amqp.Table{"x-delayed-type": "direct"})
	if err != nil {
		return err
	}

	err = r.DlqCh.ExchangeDeclare(r.dlqExchange, amqp.ExchangeDirect,
		true, false, false, false,
		nil)
	if err != nil {
		return err
	}

	err = r.declareCobrancaQueues()
	if err != nil {
		return err
	}

	_, err = r.webhookCh.QueueDeclare(r.webhookQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = r.webhookCh.QueueBind(r.webhookQueue, r.webhookQueue, r.exchange, false, nil)
	if err != nil {
		return err
	}

	_, err = r.dbCh.QueueDeclare(r.dbqueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = r.dbCh.QueueBind(r.dbqueue, r.dbqueue, r.exchange, false, nil)
	if err != nil {
		return err
	}

	_, err = r.DlqCh.QueueDeclare(r.dlqQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = r.DlqCh.QueueBind(r.dlqQueue, r.dlqQueue, r.dlqExchange, false, nil)
	if err != nil {
		return err
	}

	return nil

}

// Reconecta com o servidor, abre novos canais de conexão, inicia os consumidores e inicia novas goroutines para consumir as filas.
func (r *RabbitMQ) Reconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil || r.conn.IsClosed() {
		conn, err := amqp.Dial(r.url)
		if err != nil {
			return fmt.Errorf("falha ao se reconectar ao rabbitmq e abrir conexão: %s", err.Error())
		}
		r.conn = conn
		r.errchan = make(chan *amqp.Error)
		r.conn.NotifyClose(r.errchan)

		if err := r.setUpChannels(); err != nil {
			return fmt.Errorf("falha ao se reconectar no rabbitmq e obter canais: %s", err.Error())
		}

	}
	if err := r.declareQueues(); err != nil {
		return fmt.Errorf("falha ao se reconectar ao rabbitmq e gravar filas: %s", err.Error())
	}
	if err := r.setUpConsumers(); err != nil {
		return fmt.Errorf("falha ao se reconectar ao rabbitmq e consumir filas: %s", err.Error())
	}

	setUpProducers(r)
	if err := r.injectProducer(); err != nil {
		return err
	}

	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-reconnect", "", "reconexão bem-sucedida com o rabbitmq", nil)

	go r.ConsumeQueues()

	return nil

}

// SetQOS define o QOS (Quality of Service) para a fila de cobranca.
// O QOS define a quantidade de mensagens que podem ser consumidas por vez.
func (r *RabbitMQ) SetQOS(qos int) error {

	if qos > 0 {
		r.QOS = qos
	}

	return r.Reconnect()
}

func (r *RabbitMQ) GetQOS() map[string]any {
	return map[string]any{
		"QOS": r.QOS,
	}
}

// Monitora a conexão, e caso ela seja fechada por algum erro, tenta se reconectar.
func (r *RabbitMQ) MonitoreServer() {
	for errConn := range r.errchan {
		helpers.LogError(r.ctx, r.logger, r.location, "rabbitmq-monitore server", "", "rabbitmq  desconectado", errConn.Error(), nil)
		for {
			time.Sleep(time.Second * 20)
			err := r.Reconnect()
			if err == nil {
				return
			}
			helpers.LogError(r.ctx, r.logger, r.location, "rabbitmq-monitore server", "", "falha ao se reconectar ao rabbitmq", err.Error(), nil)
		}
	}
}

// Encerra a conexão com o rabbitmq
func (r *RabbitMQ) Close() error {

	if err := r.dbCh.Close(); err != nil {
		helpers.LogError(r.ctx, r.logger, r.location, "rabbitmq-close", "", "falha ao fechar canal de db", err.Error(), nil)
		return err
	}

	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-close", "", "canal de db fechado", nil)

	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-close", "", "canal de simulacao dlq fechado", nil)

	if err := r.DlqCh.Close(); err != nil {
		helpers.LogError(r.ctx, r.logger, r.location, "rabbitmq-close", "", "falha ao fechar canal de digitacao dlq", err.Error(), nil)
		return err
	}
	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-close", "", "canal de digitacao dlq fechado", nil)

	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-close", "", "canal de dados bancários fechado", nil)

	if err := r.closeCobranca(); err != nil {
		return err
	}

	if err := r.conn.Close(); err != nil {
		helpers.LogError(r.ctx, r.logger, r.location, "rabbitmq-close", "", "falha ao fechar conexão com o rabbitmq", err.Error(), nil)
		return err
	}
	helpers.LogInfo(r.ctx, r.logger, r.location, "rabbitmq-close", "", "conexão com o rabbitmq fechada", nil)
	return nil
}

func (r *RabbitMQ) Check() (map[string]any, error) {
	var info = make(map[string]any)
	var chInfo = make(map[string]any)
	var globalChInfo = make(map[string]any)
	var errors = make([]string, 0)
	if r.conn.IsClosed() {
		errors = append(errors, "conexão fechada com o Rabbitmq")
	}

	if r.dbCh.IsClosed() {
		errors = append(errors, "canal de db fechado")
		globalChInfo["DB"] = "fechado"
	} else {
		globalChInfo["DB"] = "ok"
	}

	if r.DlqCh.IsClosed() {
		globalChInfo["Digitacao DLQ"] = "fechado"
		errors = append(errors, "canal de digitação-dlq fechado")

	} else {
		globalChInfo["Digitacao DLQ"] = "ok"
	}

	chInfo["global"] = globalChInfo

	data, err := r.checkCreditoPessoal()
	chInfo["cobranca"] = data
	if err != nil {
		errors = append(errors, err.Error())
	}
	info["canais"] = chInfo

	if len(errors) > 0 {
		info["errors"] = errors
		info["status"] = "error"
	} else {
		info["status"] = "ok"

	}

	info["QOS"] = r.GetQOS()

	return info, nil

}

func (r *RabbitMQ) GetServiceName() string {
	return "rabbitmq"
}
