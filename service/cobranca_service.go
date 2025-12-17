package service

import (
	"cobranca-bmp/cache"
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"
)

// Representa um service para as operações de formalização/inclusão e geração de assinaturas para propostas.
type CobrancalService struct {
	ctx            context.Context
	logger         *slog.Logger
	loc            *time.Location
	client         CreditoPessoalClient
	queue          QueueProducer
	cache          *cache.RedisCache
	webhookService *WebhookService
}

// Configurando o producer das filas
func (a *CobrancalService) SetProducer(q any) error {
	if q == nil {
		return errors.New("producer não pode ser nulo")
	}
	queue, ok := q.(Queue)

	if !ok {
		return errors.New("producer deve implementar queue")
	}
	a.queue = queue
	return nil
}

// Instancia um novo service.
func NewCobrancaService(ctx context.Context, logger *slog.Logger, loc *time.Location, client CreditoPessoalClient, webhookService *WebhookService, cache *cache.RedisCache) *CobrancalService {
	return &CobrancalService{
		ctx:            ctx,
		logger:         logger,
		loc:            loc,
		client:         client,
		webhookService: webhookService,
	}
}

func (a *CobrancalService) Auth(key, client string) (string, error) {

	return a.client.Auth(key, client, false, config.INSSCP_JWT_KEY)
}

func (a *CobrancalService) Cobranca(payload models.CobrancaTaskData) (any, string, int, error) {
	return a.GerarCobranca(&payload)

}

func (a *CobrancalService) GerarCobranca(payload *models.CobrancaTaskData) (any, string, int, error) {
	var status string
	var errCobrancas error
	var statusCode int
	var data any

	if payload.Token == "" {
		token, err := a.client.Auth(payload.IdempotencyKey, config.INSSCP_CLIENT_ID, false, config.INSSCP_JWT_KEY)
		if err != nil {
			helpers.LogError(a.ctx, a.logger, a.loc, "cobrança crédito pessoal", "401", "Erro ao buscar token", err.Error(), err)
			return a.HandleErrorCobranca(config.API_STATUS_UNAUTHORIZED, 401, *payload, models.NewAPIError("Erro ao buscar token", strconv.Itoa(payload.IdProposta), "CONFIAPAY"))
		}
		payload.Token = token
	}

	//a.updateService.UpdatePropostaParcelaFormaCobranca(payload.IdProposta, payload.TipoCobranca, payload.NumeroCCB, false)

	if payload.MultiplasCobrancas {

		data, statusCode, status, errCobrancas = a.GerarCobrancaParcelasMultiplas(payload)
	} else {
		data, statusCode, status, errCobrancas = a.GerarCobrancaParcela(payload)
	}

	if errCobrancas != nil {
		return a.HandleErrorCobranca(status, statusCode, *payload, errCobrancas.(models.APIError))
	}

	if payload.CalledAssync {
		a.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, data, "cobranca service"))
	}

	return data, "", 200, nil
}

// Realiza o tratamento de erros para a liberação de propostas.
func (a *CobrancalService) HandleErrorCobranca(status string, statusCode int, payload models.CobrancaTaskData, errAPI models.APIError) (any, string, int, error) {

	if errAPI.ID <= 0 {
		errAPI.ID = payload.IdProposta
	}

	if len(errAPI.Messages) < 1 {
		errAPI.Messages = []models.APIMessage{
			{
				Description: errAPI.Error(),
			},
		}
	}

	if (status == config.API_STATUS_TIMEOUT && payload.TimeoutRetries > 1) || (status == config.API_STATUS_RATE_LIMIT && payload.RateLimitRetries > 1) {
		payload.SetTry(config.TIMEOUT_DELAY, status)
		if status == config.API_STATUS_TIMEOUT {
			payload.IdempotencyKey = strconv.Itoa(a.cache.GenID())
		}
		payload.CalledAssync = true

		err := a.queue.Produce(config.COBRANCA_QUEUE, payload, payload.CurrentDelay)
		if err != nil {

			return nil, config.API_STATUS_ERR, 500, models.NewAPIError("", "Erro ao enviar dados para fila: "+err.Error(), strconv.Itoa(payload.IdProposta))
		}

		var resp = models.NewAPIError("", "A geração de boleto/pix entrou na fila de processamento. Aguarde!", strconv.Itoa(payload.IdProposta))
		resp.HasError = false
		return resp, "", 200, nil

	}

	//	if payload.Origem != 1 && !reprocessamento {
	//	a.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, errAPI, "cobrança crédito pessoal"))

	//	}
	//helpers.LogError(a.ctx, a.logger, a.loc, "", strconv.Itoa(statusCode), "Erro ao realizar cobrança", errAPI.Error(), errAPI.Result)
	if payload.CalledAssync {
		a.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, errAPI, "cobranca service"))
	}

	return nil, status, statusCode, errAPI

}
func (a CobrancalService) SendToDLQ(data any) error {
	return a.queue.Produce(config.DIGTACAO_QUEUE_DLQ, data, 0)

}

func (a *CobrancalService) GerarCobrancaParcela(payload *models.CobrancaTaskData) (any, int, string, error) {

	data, statusCode, status, err := a.client.GerarCobrancaParcela(payload.CobrancaUnicaFrontendInput.CobrancaUnicaInput, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := a.client.Auth(strconv.Itoa(a.cache.GenID()), config.INSSCP_CLIENT_ID, true, config.INSSCP_JWT_KEY)
		if authErr != nil {
			return data, statusCode, status, err
		}
		payload.Token = token

		data, statusCode, status, err = a.client.GerarCobrancaParcela(models.CobrancaUnicaInput{}, payload.Token, payload.IdempotencyKey)
	}
	if err != nil {
		errAPI, ok := err.(models.APIError)
		if !ok {
			errAPI = models.NewAPIError("", err.Error(), payload.Token, payload.IdempotencyKey)
		}
		return data, statusCode, status, errAPI
	}

	payload.Reset()
	//a.updateService.UpdatePropostaParcelasCodLiquidacao(payload.IdProposta, &data, false)

	return nil, 200, "", nil
}
func (a *CobrancalService) GerarCobrancaParcelasMultiplas(payload *models.CobrancaTaskData) (any, int, string, error) {
	var payloadBMP models.MultiplasCobrancasInput
	payloadBMP.Dto = models.DtoCobranca{
		CodigoProposta: payload.NumeroAcompanhamento,
		CodigoOperacao: strconv.Itoa(payload.IdProposta),
	}

	payloadBMP.DtoGerarMultiplasLiquidacoes = models.DtoGerarMultiplasLiquidacoes{
		DescricaoLiquidacao: "Parcelamento da proposta" + strconv.Itoa(payload.IdProposta),
		//NotificarCliente:    payload.TipoCobranca == config.CREDITO_PESSOAL_COBRANCA_BOLETO || payload.TipoCobranca == config.CREDITO_PESSOAL_COBRANCA_BOLETOPIX,
		//PagamentoViaBoleto:  payload.TipoCobranca == config.CREDITO_PESSOAL_COBRANCA_BOLETO || payload.TipoCobranca == config.CREDITO_PESSOAL_COBRANCA_BOLETOPIX,
		//PagamentoViaPIX:     payload.TipoCobranca == config.CREDITO_PESSOAL_COBRANCA_PIX,
		Parcelas:     make([]models.ParcelaCobranca, 0),
		TipoRegistro: 2,
	}

	data, statusCode, status, err := a.client.GerarCobrancaParcelasMultiplas(payloadBMP, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := a.client.Auth(strconv.Itoa(a.cache.GenID()), config.INSSCP_CLIENT_ID, true, config.INSSCP_JWT_KEY)
		if authErr != nil {
			return data, statusCode, status, err
		}
		payload.Token = token

		data, statusCode, status, err = a.client.GerarCobrancaParcelasMultiplas(payloadBMP, payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		errAPI, ok := err.(models.APIError)
		if !ok {
			errAPI = models.NewAPIError("", err.Error(), strconv.Itoa(payload.IdProposta))
		}
		return data, statusCode, status, errAPI
	}

	//a.updateService.UpdatePropostaParcelasCodLiquidacao(payload.IdProposta, &data, false)

	return data, 200, status, nil
}

func (a *CobrancalService) CancelarCobranca(payload models.CobrancaTaskData) (any, string, int, error) {
	idempotencyKey := strconv.Itoa(a.cache.GenID())
	token, err := a.client.Auth(idempotencyKey, config.INSSCP_CLIENT_ID, false, config.INSSCP_JWT_KEY)
	if err != nil {
		errApi, ok := err.(models.APIError)
		if !ok {
			errApi = models.NewAPIError("", "Não foi possível buscar token", strconv.Itoa(payload.IdProposta))
		}
		helpers.LogError(a.ctx, a.logger, a.loc, "cobrança crédito pessoal", "401", "Erro ao buscar token", err.Error(), err)
		return nil, config.API_STATUS_UNAUTHORIZED, 401, errApi
	}
	payload.Token = token

	data, statusCode, status, err := a.client.CancelarCobranca(payload.CancelamentoCobranca, payload.Token, idempotencyKey)
	if err != nil {

		if status == config.API_STATUS_UNAUTHORIZED {
			token, authErr := a.client.Auth(strconv.Itoa(a.cache.GenID()), config.INSSCP_CLIENT_ID, true, config.INSSCP_JWT_KEY)

			if authErr != nil {
				errApi, ok := err.(models.APIError)
				if !ok {
					errApi = models.NewAPIError("", "Não foi possível buscar token", strconv.Itoa(payload.IdProposta))
				}

				return nil, status, statusCode, errApi
			}
			payload.Token = token
			idempotencyKey = strconv.Itoa(a.cache.GenID())
			data, statusCode, status, err = a.client.CancelarCobranca(payload.CancelamentoCobranca, payload.Token, idempotencyKey)
		}

		if err != nil {
			errApi, ok := err.(models.APIError)
			if !ok {
				errApi = models.NewAPIError("", "Não foi possível buscar token", strconv.Itoa(payload.IdProposta))
			}
			return nil, status, statusCode, errApi
		}

	}
	return data, status, statusCode, nil
}

func (a *CobrancalService) ConsultarCobranca(payload models.ConsultarDetalhesInput) (models.ConsultaCobrancaResponse, int, string, error) {
	idempotencyKey := strconv.Itoa(a.cache.GenID())
	token, err := a.client.Auth(idempotencyKey, config.INSSCP_CLIENT_ID, false, config.INSSCP_JWT_KEY)
	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "cobrança crédito pessoal", "401", "Erro ao buscar token", err.Error(), err)
		return models.ConsultaCobrancaResponse{}, 401, "", err
	}
	data, statusCode, status, err := a.client.ConsultarCobranca(payload, token, idempotencyKey)
	if err != nil {
		if status == config.API_STATUS_UNAUTHORIZED {
			token, authErr := a.client.Auth(strconv.Itoa(a.cache.GenID()), config.INSSCP_CLIENT_ID, true, config.INSSCP_JWT_KEY)
			if authErr != nil {
				return data, statusCode, status, err
			}
			idempotencyKey = strconv.Itoa(a.cache.GenID())
			data, statusCode, status, err = a.client.ConsultarCobranca(payload, token, idempotencyKey)

		}
		if err != nil {
			return data, statusCode, status, err
		}

	}
	return data, statusCode, status, nil
}
