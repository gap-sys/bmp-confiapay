package handlers

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (w *WebhookController) WebhookCobranca() fiber.Handler {

	return func(c *fiber.Ctx) error {
		var rawBody = append([]byte{}, c.Body()...)
		var headers = c.Request().Header.String()

		var resp any

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

					/*
						var geracaoAgenda models.EventoGeracaoAgenda
						json.Unmarshal(rawBody, &geracaoAgenda)
						bodyProcessed = true

						for _, p := range geracaoAgenda.Parcelas {
							parcelas = append(parcelas, p.NroParcela)

						}*/

				case 1:
					/*
						var cancelamento models.EventoCancelamentoAgenda
						json.Unmarshal(rawBody, &cancelamento)
						SearchMode = "liquidacao"
						bodyProcessed = true
					*/

					return

				case 2:
					/*
						var abatimento models.EventoAcrescimoAbatimento
						json.Unmarshal(rawBody, &abatimento)
						SearchMode = "parcela"
						bodyProcessed = true
					*/

					return

				case 3: //Parcela paga
					var lancamento models.EventoLancamentoParcela
					json.Unmarshal(rawBody, &lancamento)
					bodyProcessed = true
					cobrancaInfo, err := w.cobrancaService.FindByNumParcela(lancamento.LancamentoParcela.NroParcela, input.NroProposta)
					if err != nil {
						var dlqData = models.DLQData{
							Payload:  resp,
							Mensagem: "Erro ao buscar cobranca por número de parcela",
							Erro:     "Erro ao buscar cobranca por número de parcela",
							Contexto: "webhook cobranças",
							Time:     time.Now().In(w.location),
						}

						w.webhookService.SendToDLQ(dlqData)
						return

					}

					var whData = make(map[string]any)
					whData["id_proposta_parcela"] = cobrancaInfo.IdPropostaParcela
					whData["codigo_liquidacao"] = cobrancaInfo.CodigoLiquidacao
					whData["data_pagamento"] = lancamento.DtEvento
					whData["valor_pago"] = lancamento.LancamentoParcela.VlrPagamento
					whData["encargos"] = lancamento.LancamentoParcela.VlrEncargos
					whData["valor_desconto"] = lancamento.LancamentoParcela.VlrDesconto
					whData["operacao"] = "P"

					w.webhookService.RequestToWebhook(models.NewWebhookTaskData(cobrancaInfo.UrlWebhook, whData, "cancelamento-cobranca"))

					return

				case 4:
					/*
						var prorrogacao models.EventoProrrogacaoVencimento
						json.Unmarshal(rawBody, &prorrogacao)
						SearchMode = "parcela"
						bodyProcessed = true
					*/

					return

				case 5:

					var geracaoBoleto models.EventoGeracaoBoleto
					json.Unmarshal(rawBody, &geracaoBoleto)
					bodyProcessed = true
					cobrancaInfo, err := w.cobrancaService.FindByDataVencimento(geracaoBoleto.GeracaoBoleto.DtVencimento, input.NroProposta)
					if err != nil {
						var dlqData = models.DLQData{
							Payload:  resp,
							Mensagem: "Erro ao buscar cobranca por data de vencimento",
							Erro:     "Erro ao buscar cobranca por data de vencimento",
							Contexto: "webhook cobranças",
							Time:     time.Now().In(w.location),
						}

						w.webhookService.SendToDLQ(dlqData)
						return

					}

					var authPAyload = models.AuthPayload{
						IdConvenio:       cobrancaInfo.IdConvenio,
						IdSecuritizadora: cobrancaInfo.IdSecuritizadora,
					}

					cobrancaPayload := models.NewCobrancaTastkData(w.cache.GenID(), config.STATUS_CONSULTAR_COBRANCA, cobrancaInfo.IdProposta, cobrancaInfo.NumeroAcompanhamento, authPAyload)
					cobrancaPayload.CobrancaDBInfo = cobrancaInfo
					cobrancaPayload.ConsultarCobrancaInput = models.ConsultarDetalhesInput{
						DTO: models.DtoCobranca{
							CodigoProposta: cobrancaInfo.NumeroAcompanhamento,
							CodigoOperacao: strconv.Itoa(cobrancaInfo.IdProposta),
							NroParcelas:    []int{cobrancaInfo.NumeroParcela},
						},
						DTOConsultaDetalhes: models.DTOConsultaDetalhes{
							TrazerBoleto: true,
						},
					}

					cobrancaPayload.WebhookUrl = cobrancaInfo.UrlWebhook
					cobrancaPayload.CalledAssync = true

					w.cobrancaService.Cobranca(cobrancaPayload)

					return

				case 6: //Cancelamento de boleto
					var cancelamento models.EventoCancelamentoBoleto
					json.Unmarshal(rawBody, &cancelamento)
					bodyProcessed = true
					cobrancaInfo, err := w.cobrancaService.FindByNumParcela(cancelamento.CancelamentoBoleto.Parcelas[0], input.NroProposta)
					if err != nil {
						var dlqData = models.DLQData{
							Payload:  resp,
							Mensagem: "Erro ao buscar cobranca por número de parcela",
							Erro:     "Erro ao buscar cobranca por número de parcela",
							Contexto: "webhook cobranças",
							Time:     time.Now().In(w.location),
						}

						w.webhookService.SendToDLQ(dlqData)
						return

					}

					var whData = make(map[string]any)
					whData["id_proposta_parcela"] = cobrancaInfo.IdPropostaParcela
					whData["codigo_liquidacao"] = cobrancaInfo.CodigoLiquidacao
					whData["operacao"] = "C"

					w.webhookService.RequestToWebhook(models.NewWebhookTaskData(cobrancaInfo.UrlWebhook, whData, "cancelamento-cobranca"))

					return

				case 7:
					/*
						var liquidacao models.EventoLiquidacaoAgenda
						json.Unmarshal(rawBody, &liquidacao)
						SearchMode = "liquidacao"
						bodyProcessed = true
					*/

					return

				case 8: //Registro de cobrança
					var registroCobranca models.EventoRegistroCobranca
					json.Unmarshal(rawBody, &registroCobranca)
					//helpers.LogError(c.Context(), w.logger, w.location, "webhook cobranças", "", "Não foi possível encontrar parcela", nil, input
					bodyProcessed = true
					cobrancaInfo, err := w.cobrancaService.FindByCodLiquidacao(registroCobranca.GeracaoCobranca.CodigoLiquidacao, input.NroProposta)
					if err != nil {
						var dlqData = models.DLQData{
							Payload:  resp,
							Mensagem: "Erro ao buscar cobranca por código de liquidação",
							Erro:     "Erro ao buscar cobranca por código de liquidação",
							Contexto: "webhook cobranças",
							Time:     time.Now().In(w.location),
						}

						w.webhookService.SendToDLQ(dlqData)
						return

					}

					var authPAyload = models.AuthPayload{
						IdConvenio:       cobrancaInfo.IdConvenio,
						IdSecuritizadora: cobrancaInfo.IdSecuritizadora,
					}

					cobrancaPayload := models.NewCobrancaTastkData(w.cache.GenID(), config.STATUS_CONSULTAR_COBRANCA, cobrancaInfo.IdProposta, cobrancaInfo.NumeroAcompanhamento, authPAyload)
					cobrancaPayload.CobrancaDBInfo = cobrancaInfo
					cobrancaPayload.ConsultarCobrancaInput = models.ConsultarDetalhesInput{
						DTO: models.DtoCobranca{
							CodigoProposta: cobrancaInfo.NumeroAcompanhamento,
							CodigoOperacao: strconv.Itoa(cobrancaInfo.IdProposta),
							NroParcelas:    []int{cobrancaInfo.NumeroParcela},
						},
						DTOConsultaDetalhes: models.DTOConsultaDetalhes{
							TrazerBoleto: true,
						},
					}

					cobrancaPayload.WebhookUrl = cobrancaInfo.UrlWebhook
					cobrancaPayload.CalledAssync = true

					w.cobrancaService.Cobranca(cobrancaPayload)

				case 9: //Cancelamento de pix
					var cancelamento models.EventoCancelamentoCobranca
					json.Unmarshal(rawBody, &cancelamento)
					SearchMode = "parcelas"
					bodyProcessed = true
					cobrancaInfo, err := w.cobrancaService.FindByNumParcela(cancelamento.CancelamentoCobranca.Parcelas[0], input.NroProposta)
					if err != nil {
						var dlqData = models.DLQData{
							Payload:  resp,
							Mensagem: "Erro ao buscar cobranca por número de parcela",
							Erro:     "Erro ao buscar cobranca por número de parcela",
							Contexto: "webhook cobranças",
							Time:     time.Now().In(w.location),
						}

						w.webhookService.SendToDLQ(dlqData)
						return

					}

					var whData = make(map[string]any)
					whData["id_proposta_parcela"] = cobrancaInfo.IdPropostaParcela
					whData["codigo_liquidacao"] = cobrancaInfo.CodigoLiquidacao
					whData["operacao"] = "C"

					w.webhookService.RequestToWebhook(models.NewWebhookTaskData(cobrancaInfo.UrlWebhook, whData, "cancelamento-cobranca"))

					return

				case 10:
					/*
						var estorno models.EventoEstorno
						json.Unmarshal(rawBody, &estorno)
						helpers.LogError(c.Context(), w.logger, w.location, "webhook cobranças", "", "Não foi possível encontrar parcela", nil, input)

						bodyProcessed = true*/
					return

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
					helpers.LogError(context.Background(), w.logger, w.location, "webhook cobrancas", "", "Payload não identificado recebido no webhook", "Payload não identificado recebido no webhook", payload)
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
