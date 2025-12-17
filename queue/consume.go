package queue

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"encoding/json"
	"time"
)

// ConsumeQueues consome as mensagens das filas.
func (r *RabbitMQ) ConsumeQueues() {
	//	go r.ConsumeDB()
	go r.ConsumeWebhook()
	go r.MonitoreServer()
	go r.ConsumeCobrancaQueues()
}

// setUpConsumers configura os consumers para as filas digitacaoDlq, simulacaoDlq, dados bancarios e webhook.
func (r *RabbitMQ) setUpConsumers() error {
	msgs, err := r.dbCh.Consume(r.dbqueue, "bmp-fgts-update", false, false, false, false, nil)
	if err != nil {
		return err
	}
	r.dbMsgs = msgs

	msgs, err = r.webhookCh.Consume(r.webhookQueue, "bmp-fgts-webhook", false, false, false, false, nil)
	if err != nil {
		return err
	}
	r.webhookMsgs = msgs

	if err := r.setUpCobrancaConsumers(); err != nil {
		return err
	}

	return nil
}

// ConsumeWebhook ficará em execução para ouvir as mensagens da fila de envio para o webhook Confiapay.
func (r *RabbitMQ) ConsumeWebhook() {
	for msg := range r.webhookMsgs {
		var payload models.WebhookTaskData
		err := json.Unmarshal(msg.Body, &payload)
		if err != nil {
			var dlqData = models.DLQData{
				Payload:  payload,
				Contexto: "consumo de webhook",
				Mensagem: "Erro ao deserializar json para requisitar para webhook na fila",
				Erro:     err.Error(),
				Time:     time.Now().In(r.location),
			}
			if err := msg.Nack(false, false); err != nil {
				dlqData.Mensagem += "\n " + "Erro ao realizar Nack de mensagem"
				dlqData.Erro += "\n " + err.Error()
				helpers.LogError(r.ctx, r.logger, r.location, "Consumo de webhook", "", "Erro ao realizar Nack de mensagem", err.Error(), nil)
			}
			helpers.LogError(r.ctx, r.logger, r.location, "Consumo de webhook", "", "Erro ao deserializar json para requisitar para webhook na fila", err.Error(), nil)
			r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
			continue
		}

		helpers.LogInfo(r.ctx, r.logger, r.location, "Consumo de webhook", "", "Entrou na fila de webhook", payload)

		if err := r.webhookService.RequestToWebhook(payload); err != nil {
			if err := msg.Nack(false, false); err != nil {
				var dlqData = models.DLQData{
					Payload:  payload,
					Contexto: "consumo de webhook",
					Mensagem: "Erro ao realizar Nack de mensagem",
					Erro:     err.Error(),
					Time:     time.Now().In(r.location),
				}
				helpers.LogError(r.ctx, r.logger, r.location, "Consumo de webhook", "", "Erro ao realizar Nack de mensagem", err.Error(), nil)
				r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
			}
			continue
		}

		if err := msg.Ack(false); err != nil {
			var dlqData = models.DLQData{
				Payload:  payload,
				Contexto: "consumo de webhook",
				Mensagem: "Erro ao realizar Ack de mensagem",
				Erro:     err.Error(),
				Time:     time.Now().In(r.location),
			}

			helpers.LogError(r.ctx, r.logger, r.location, "Consumo de webhook", "", "Erro ao realizar Ack de mensagem", err.Error(), nil)
			r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
		}

	}
}

// ConsumeDB ficará em execução para ouvir as mensagens da fila de db.
func (r *RabbitMQ) ConsumeDB() {
	for msg := range r.dbMsgs {
		var payload models.UpdateDBData
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			var dlqData = models.DLQData{
				Payload:  payload,
				Contexto: "consumo de db",
				Mensagem: "Erro ao deserializar json para atualização na fila",
				Erro:     err.Error(),
				Time:     time.Now().In(r.location),
			}
			if err := msg.Nack(false, false); err != nil {
				dlqData.Mensagem += "\n " + "Erro ao realizar Nack de mensagem"
				dlqData.Erro += "\n " + err.Error()
				helpers.LogError(r.ctx, r.logger, r.location, "Consumo de db", "", "Erro ao realizar Nack de mensagem", err.Error(), nil)
			}
			r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
			helpers.LogError(r.ctx, r.logger, r.location, "Consumo de db", "", "Erro ao deserializar json para atualização no db", err.Error(), nil)
			continue

		}

		helpers.LogInfo(r.ctx, r.logger, r.location, "Consumo de db", "", "Entrou na fila de db", payload)
		var err error
		var noConn bool

		noConn, err = r.dbService.UpdateAssync(payload)

		if err != nil {
			if err := msg.Nack(false, false); err != nil {
				var dlqData = models.DLQData{
					Payload:  payload,
					Contexto: "consumo de db",
					Mensagem: "Erro ao realizar Nack de mensagem",
					Erro:     err.Error(),
					Time:     time.Now().In(r.location),
				}
				r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
				helpers.LogError(r.ctx, r.logger, r.location, "Consumo de db", "", "Erro ao realizar Nack de mensagem", err.Error(), nil)
			}
			if noConn {
				r.baseProducer.Produce(r.dbqueue, payload, config.DB_QUEUE_DELAY)
			} else {
				var dlqData = models.DLQData{
					Payload:  payload,
					Contexto: "consumo de db",
					Mensagem: "Erro ao realizar inserção em proposta_status",
					Erro:     err.Error(),
					Time:     time.Now().In(r.location),
				}
				r.baseProducer.Produce(r.dlqQueue, dlqData, 0)

			}
			continue

		}

		if err := msg.Ack(false); err != nil {
			var dlqData = models.DLQData{
				Payload:  payload,
				Contexto: "consumo de db",
				Mensagem: "Erro ao realizar Ack de mensagem",
				Erro:     err.Error(),
				Time:     time.Now().In(r.location),
			}
			r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
			helpers.LogError(r.ctx, r.logger, r.location, "Consumo de db", "", "Erro ao realizar Ack de mensagem", err.Error(), nil)
		}

	}

}
