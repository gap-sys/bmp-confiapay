package queue

import (
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"encoding/json"
	"time"
)

// Configura os consumidores para as filas
func (r *RabbitMQ) setUpCobrancaConsumers() error {

	msgs, err := r.cobrancaCh.Consume(r.cobrancaQueue, "bmp-cobranca", false, false, false, false, nil)
	if err != nil {
		return err
	}

	r.cobrancaCreditoPessoalMsgs = msgs

	return nil
}

func (r *RabbitMQ) ConsumeCobrancaQueues() {
	go r.ConsumeCobranca()

}

// ConsumeCobranca ficará em execução para ouvir as mensagens da fila de cobranças do CreditoPessoal.
func (r *RabbitMQ) ConsumeCobranca() {
	for msg := range r.cobrancaCreditoPessoalMsgs {
		var payload models.CobrancaTaskData
		err := json.Unmarshal(msg.Body, &payload)
		if err != nil {
			var dlqData = models.DLQData{
				Payload:  payload,
				Contexto: "Consumo de cobrança",
				Mensagem: "Erro ao deserializar json para requisitar para cobrança na fila",
				Erro:     err.Error(),
				Time:     time.Now().In(r.location),
			}

			if err := msg.Nack(false, false); err != nil {
				dlqData.Mensagem += "\n " + "Erro ao realizar Nack de mensagem"
				dlqData.Erro += "\n " + err.Error()
				helpers.LogError(r.ctx, r.logger, r.location, "Consumo de cobrança", "", "Erro ao realizar Nack de mensagem", err.Error(), nil)
			}
			helpers.LogError(r.ctx, r.logger, r.location, "Consumo de cobrança", "", "Erro ao deserializar json para requisitar para cobrança na fila", err.Error(), nil)
			r.cobrancaProducer.Produce(r.dlqQueue, dlqData, 0)
			continue
		}

		helpers.LogInfo(r.ctx, r.logger, r.location, "Consumo de cobrança", "", "Entrou na fila de cobrança", map[string]any{
			"IdempotencyKey":   payload.IdempotencyKey,
			"IdProposta":       payload.IdProposta,
			"timeoutRetries":   payload.TimeoutRetries,
			"rateLimitRetries": payload.RateLimitRetries,
			"delay":            payload.CurrentDelay,
		})

		_, _, _, err = r.cobrancaCreditoPessoalService.Cobranca(payload)
		if err != nil {
			if err := msg.Nack(false, false); err != nil {
				var dlqData = models.DLQData{
					Payload:  payload,
					Contexto: "Consumo de cobrança",
					Mensagem: "Erro ao realizar Nack de mensagem",
					Erro:     err.Error(),
					Time:     time.Now().In(r.location),
				}
				helpers.LogError(r.ctx, r.logger, r.location, "Consumo de cobrança", "", "Erro ao realizar Nack de mensagem", err.Error(), nil)
				r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
			}
			continue
		}

		if err := msg.Ack(false); err != nil {
			var dlqData = models.DLQData{
				Payload:  payload,
				Contexto: "Consumo de cobrança",
				Mensagem: "Erro ao realizar Ack de mensagem",
				Erro:     err.Error(),
				Time:     time.Now().In(r.location),
			}

			helpers.LogError(r.ctx, r.logger, r.location, "Consumo de cobrança", "", "Erro ao realizar Ack de mensagem", err.Error(), nil)
			r.baseProducer.Produce(r.dlqQueue, dlqData, 0)
		}

	}
}
