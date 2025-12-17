package handlers

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (w *WebhookController) WebhookCobranca() fiber.Handler {

	return func(c *fiber.Ctx) error {
		var rawBody = append([]byte{}, c.Body()...)
		var headers = c.Request().Header.String()

		var resp any
		var parcelas = make([]int, 0)

		apiKey := c.Get("Api-Key")
		if apiKey == "" {
			helpers.LogError(c.Context(), w.logger, w.location, "webhook insscp", "", "ApiKey vazia enviada", nil, map[string]string{"url": c.OriginalURL(),
				"headers": c.Request().Header.String(),
			})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ApiKey é obrigatória"})

		} else if apiKey != config.API_KEY {
			helpers.LogError(c.Context(), w.logger, w.location, "webhook insscp", "", "ApiKey inválida enviada", nil, map[string]string{"url": c.OriginalURL(),
				"headers": c.Request().Header.String(),
			})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ApiKey inválida"})

		}

		go func() {
			var input *models.EventoCobranca
			var bodyProcessed bool = false

			if err := json.Unmarshal(rawBody, &resp); err != nil {
				var dlqData = models.DLQData{
					Payload:  rawBody,
					Mensagem: "Não foi possível deserializar body de requisição",
					Erro:     "Não foi possível deserializar body de requisição",
					Contexto: "webhook cobranças",
					Time:     time.Now().In(w.location),
				}
				w.webhookService.SendToDLQ(dlqData)

			}

			if err := json.Unmarshal(rawBody, &input); err != nil {
				var dlqData = models.DLQData{
					Payload:  rawBody,
					Mensagem: "Não foi possível deserializar body de requisição",
					Erro:     "Não foi possível deserializar body de requisição",
					Contexto: "webhook cobranças",
					Time:     time.Now().In(w.location),
				}
				w.webhookService.SendToDLQ(dlqData)

			}

			if input != nil {
				var SearchMode string
				SearchMode += ""

				/*if propostaInfo.Convenio != config.CREDITO_PESSOAL_CONVENIO {
					helpers.LogDebug(c.Context(), w.logger, w.location, "webhook cobranças", "", "Convênio não possui geração de cobranças", resp)
					return

				}*/

				//helpers.LogDebug(context.Background(), w.logger, w.location, "webhook cobranças", "", "Body recebido", input)
				switch input.TipoEvento {

				case
					0:
					var geracaoAgenda models.EventoGeracaoAgenda
					json.Unmarshal(rawBody, &geracaoAgenda)
					bodyProcessed = true

					for _, p := range geracaoAgenda.Parcelas {
						parcelas = append(parcelas, p.NroParcela)

					}

				case 1:
					var cancelamento models.EventoCancelamentoAgenda
					json.Unmarshal(rawBody, &cancelamento)
					SearchMode = "liquidacao"
					bodyProcessed = true

					return

				case 2:
					var abatimento models.EventoAcrescimoAbatimento
					json.Unmarshal(rawBody, &abatimento)
					SearchMode = "parcela"
					bodyProcessed = true

					return

				case 3:
					var lancamento models.EventoLancamentoParcela
					json.Unmarshal(rawBody, &lancamento)
					SearchMode = "parcela"
					bodyProcessed = true

					return

				case 4:
					var prorrogacao models.EventoProrrogacaoVencimento
					json.Unmarshal(rawBody, &prorrogacao)
					SearchMode = "parcela"
					bodyProcessed = true

					return

				case 5:
					var geracaoBoleto models.EventoGeracaoBoleto
					json.Unmarshal(rawBody, &geracaoBoleto)
					SearchMode = "data_vencimento"
					bodyProcessed = true

				case 6:
					var cancelamento models.EventoCancelamentoBoleto
					json.Unmarshal(rawBody, &cancelamento)
					SearchMode = "parcelas"
					bodyProcessed = true

					return

				case 7:
					var liquidacao models.EventoLiquidacaoAgenda
					json.Unmarshal(rawBody, &liquidacao)
					SearchMode = "liquidacao"
					bodyProcessed = true

					return

				case 8:
					var registroCobranca models.EventoRegistroCobranca
					json.Unmarshal(rawBody, &registroCobranca)
					helpers.LogError(c.Context(), w.logger, w.location, "webhook cobranças", "", "Não foi possível encontrar parcela", nil, input)

					bodyProcessed = true

				case 9:
					var cancelamento models.EventoCancelamentoCobranca
					json.Unmarshal(rawBody, &cancelamento)
					SearchMode = "parcelas"
					bodyProcessed = true

					return

				case 10:
					var estorno models.EventoEstorno
					json.Unmarshal(rawBody, &estorno)
					helpers.LogError(c.Context(), w.logger, w.location, "webhook cobranças", "", "Não foi possível encontrar parcela", nil, input)

					bodyProcessed = true

				default:
					bodyProcessed = false

				}

				if !bodyProcessed {
					var payload = map[string]any{
						"Body": resp,

						"Header": headers,
					}
					var dlqData = models.DLQData{
						Payload:  payload,
						Mensagem: "Payload não identificado recebido no webhook",
						Erro:     "Payload não identificado recebido no webhook",
						Contexto: "webhook cobranças",
						Time:     time.Now().In(w.location),
					}
					helpers.LogError(context.Background(), w.logger, w.location, "webhook insscp", "", "Payload não identificado recebido no webhook", "Payload não identificado recebido no webhook", payload)
					w.webhookService.SendToDLQ(dlqData)
					return
				}

				//Verifica se a parcela já possui as cobranças geradas antes de ir buscar no BMP

				return

			} else {

				var payload = map[string]any{
					"Body": resp,

					"Header": headers,
				}
				var dlqData = models.DLQData{
					Payload:  payload,
					Mensagem: "Payload não identificado recebido no webhook",
					Erro:     "Payload não identificado recebido no webhook",
					Contexto: "webhook cobranças",
					Time:     time.Now().In(w.location),
				}
				helpers.LogError(context.Background(), w.logger, w.location, "webhook insscp", "", "Payload não identificado recebido no webhook", "Payload não identificado recebido no webhook", payload)
				w.webhookService.SendToDLQ(dlqData)
				return

			}

		}()

		return c.JSON(fiber.Map{"success": "payload recebido"})

	}
}
