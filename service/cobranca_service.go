package service

import (
	"cobranca-bmp/cache"
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// Representa um service para as operações de formalização/inclusão e geração de assinaturas para propostas.
type CobrancalService struct {
	ctx               context.Context
	logger            *slog.Logger
	loc               *time.Location
	client            CobrancaClient
	queue             QueueProducer
	cache             *cache.RedisCache
	parcelaRepository ParcelaRepository
	webhookService    *WebhookService
	updateService     *UpdateService
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
func NewCobrancaService(ctx context.Context, logger *slog.Logger, loc *time.Location,
	client CobrancaClient, webhookService *WebhookService,
	cache *cache.RedisCache, parcelaRepository ParcelaRepository, updateService *UpdateService) *CobrancalService {
	return &CobrancalService{
		ctx:               ctx,
		logger:            logger,
		loc:               loc,
		client:            client,
		webhookService:    webhookService,
		cache:             cache,
		updateService:     updateService,
		parcelaRepository: parcelaRepository,
	}
}

func (c CobrancalService) Auth(data models.AuthPayload) (string, error) {
	return c.client.Auth(data)
}

func (a *CobrancalService) Cobranca(payload models.CobrancaTaskData) (any, string, int, error) {
	if payload.Token == "" {
		token, err := a.Auth(payload.AuthPayload)
		if err != nil {
			return nil, "", 401, err
		}
		payload.Token = token

	}

	switch payload.Status {
	case config.STATUS_CANCELAR_COBRANCA:
		return a.CancelarCobranca(payload)
	case config.STATUS_CONSULTAR_COBRANCA:
		return a.ConsultarCobranca(payload)
	case config.STATUS_GERAR_COBRANCA:
		return a.GerarCobranca(&payload)

	}
	return nil, "", 500, models.NewAPIError("", "Status inválido", payload.AuthPayload.Id)

}

func (a *CobrancalService) GerarCobranca(payload *models.CobrancaTaskData) (any, string, int, error) {
	var status string
	var errCobrancas error
	var statusCode int
	var data models.GerarCobrancaResponse

	if payload.MultiplasCobrancas {

		data, statusCode, status, errCobrancas = a.GerarCobrancaParcelasMultiplas(payload)
	} else {
		data, statusCode, status, errCobrancas = a.GerarCobrancaParcela(payload)
	}

	if errCobrancas != nil {
		return a.HandleErrorCobranca(status, statusCode, *payload, errCobrancas.(models.APIError))
	} else if len(data.Cobrancas) < 1 {
		return a.HandleErrorCobranca(status, statusCode, *payload, models.NewAPIError("", "Cobrancas não geradas", strconv.Itoa(payload.GerarCobrancaInput.IdProposta)))

	}

	a.updateService.UpdateGeracaoParcela(models.UpdateDbData{
		GeracaoParcela:   &payload.GerarCobrancaInput,
		CodigoLiquidacao: data.Cobrancas[0].CodigoLiquidacao,
		Action:           "update_geracao",
	}, false)

	if payload.CalledAssync {
		a.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, data, "cobranca service"))
	}

	helpers.LogInfo(a.ctx, a.logger, a.loc, "geracao cobrancas", "200", fmt.Sprintf("Cobrança gerada, código liquidação:%s", data.Cobrancas[0].CodigoLiquidacao), nil)

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

	if payload.CalledAssync {
		switch payload.Status {

		case config.STATUS_GERAR_COBRANCA, config.STATUS_CANCELAR_COBRANCA:
			var op = map[string]string{config.STATUS_GERAR_COBRANCA: "R", config.STATUS_CANCELAR_COBRANCA: "C"}
			var whData = make(map[string]any)
			whData["has_error"] = true
			whData["msg"] = errAPI.Msg
			whData["id_proposta_parcela"] = payload.IdPropostaParcela
			whData["operacao"] = op[payload.Status]

			a.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "cobranca service"))

		case config.STATUS_CONSULTAR_COBRANCA:
			var dlqData = models.DLQData{
				Payload:  payload.ConsultarCobrancaInput,
				Mensagem: "Erro ao buscar cobranças geradas",
				Erro:     "Erro ao buscar cobranças geradas",
				Contexto: "webhook cobranças",
				Time:     time.Now().In(a.loc),
			}
			a.SendToDLQ(dlqData)

		}
	}

	return nil, status, statusCode, errAPI

}
func (a CobrancalService) SendToDLQ(data any) error {
	return a.queue.Produce(config.DLQ_QUEUE, data, 0)

}

func (a *CobrancalService) GerarCobrancaParcela(payload *models.CobrancaTaskData) (models.GerarCobrancaResponse, int, string, error) {
	var payloadBMP = models.CobrancaUnicaInput{
		Dto: models.DtoCobranca{
			CodigoProposta: payload.GerarCobrancaInput.NumeroAcompanhamento,
			CodigoOperacao: strconv.Itoa(payload.GerarCobrancaInput.IdProposta),
			NroParcelas:    []int{payload.GerarCobrancaInput.NumeroParcela},
		},
		DtoGerarUnicaLiquidacao: models.DtoGerarUnicaLiquidacao{
			DtVencimento:            payload.GerarCobrancaInput.DataVencimento,
			DtExpiracao:             payload.GerarCobrancaInput.DataExpiracao,
			Liquidacao:              true,
			PagamentoViaBoleto:      payload.TipoCobranca == config.TIPO_COBRANCA_BOLETOPIX || payload.TipoCobranca == config.TIPO_COBRANCA_BOLETO,
			PagamentoViaPIX:         payload.TipoCobranca == config.TIPO_COBRANCA_PIX,
			DescricaoLiquidacao:     fmt.Sprintf("Parcela %d da proposta %d", payload.GerarCobrancaInput.NumeroParcela, payload.GerarCobrancaInput.IdProposta),
			VlrDesconto:             payload.GerarCobrancaInput.VlrDesconto,
			PermiteDescapitalizacao: false,
			TipoRegistro:            1,
		},
	}

	data, statusCode, status, err := a.client.GerarCobrancaParcela(payloadBMP, payload.Token, payload.IdempotencyKey)

	if err != nil {
		errAPI, ok := err.(models.APIError)
		if !ok {
			errAPI = models.NewAPIError("", err.Error(), payload.Token, payload.IdempotencyKey)
		}
		return data, statusCode, status, errAPI
	}

	payload.Reset()
	//a.updateService.UpdatePropostaParcelasCodLiquidacao(payload.IdProposta, &data, false)

	return data, 200, "", nil
}
func (a *CobrancalService) GerarCobrancaParcelasMultiplas(payload *models.CobrancaTaskData) (models.GerarCobrancaResponse, int, string, error) {
	var payloadBMP models.MultiplasCobrancasInput
	payloadBMP.Dto = models.DtoCobranca{
		CodigoProposta: payload.NumeroAcompanhamento,
		CodigoOperacao: strconv.Itoa(payload.IdProposta),
	}

	payloadBMP.DtoGerarMultiplasLiquidacoes = models.DtoGerarMultiplasLiquidacoes{
		DescricaoLiquidacao: fmt.Sprintf("Parcela %d da proposta %d", payload.GerarCobrancaInput.NumeroParcela, payload.GerarCobrancaInput.IdProposta),
		NotificarCliente:    false,
		TipoRegistro:        1,
		PagamentoViaBoleto:  payload.TipoCobranca == config.TIPO_COBRANCA_BOLETOPIX || payload.TipoCobranca == config.TIPO_COBRANCA_BOLETO,
		PagamentoViaPIX:     payload.TipoCobranca == config.TIPO_COBRANCA_PIX,
		Parcelas: []models.ParcelaCobranca{
			{
				NroParcela:              payload.GerarCobrancaInput.NumeroParcela,
				DtVencimento:            payload.GerarCobrancaInput.DataVencimento,
				DtExpiracao:             payload.GerarCobrancaInput.DataExpiracao,
				PermiteDescapitalizacao: false,
				VlrDesconto:             payload.GerarCobrancaInput.VlrDesconto,
			},
		},
	}

	data, statusCode, status, err := a.client.GerarCobrancaParcelasMultiplas(payloadBMP, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := a.Auth(payload.AuthPayload)
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

	data, statusCode, status, err := a.client.CancelarCobranca(payload.CancelamentoCobranca, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := a.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, status, statusCode, err
		}

		payload.Token = token
		data, statusCode, status, err = a.client.CancelarCobranca(payload.CancelamentoCobranca, payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		errApi, ok := err.(models.APIError)
		if !ok {
			errApi = models.NewAPIError("", "Não foi possível buscar token", strconv.Itoa(payload.IdProposta))
		}
		return nil, status, statusCode, errApi
	}

	a.updateService.UpdateCancelamentoParcela(models.UpdateDbData{
		CancelamentoCobranca: &payload.CancelamentoData,
		CodigoLiquidacao:     payload.CancelamentoCobranca.DTOCancelarCobrancas.CodigosLiquidacoes[0],
		Action:               "update_cancelamento",
	}, false)

	return data, status, statusCode, nil
}

func (a *CobrancalService) ConsultarCobranca(payload models.CobrancaTaskData) (models.ConsultaCobrancaResponse, string, int, error) {

	data, statusCode, status, err := a.client.ConsultarCobranca(payload.ConsultarCobrancaInput, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := a.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, status, statusCode, err
		}

		payload.Token = token
		data, statusCode, status, err = a.client.ConsultarCobranca(payload.ConsultarCobrancaInput, payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		return data, status, statusCode, err
	}

	if payload.WebhookUrl != "" {
		var whData = make(map[string]any)
		whData["id_proposta_parcela"] = payload.CobrancaDBInfo.IdPropostaParcela
		whData["codigo_liquidacao"] = payload.CobrancaDBInfo.CodigoLiquidacao
		whData["operacao"] = "R"

		switch payload.CobrancaDBInfo.IdFormaCobranca {
		case config.TIPO_COBRANCA_BOLETOPIX:
			whData["boleto"] = data.Parcelas[0].Boletos

		case config.TIPO_COBRANCA_PIX:
			whData["pix"] = data.Parcelas[0].Pix
		}

		a.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "consulta-cobranca"))

	}

	return data, status, statusCode, nil
}

func (c *CobrancalService) FindByCodLiquidacao(codigoLiquidacao string, numeroCCB int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByCodLiquidacao(codigoLiquidacao, numeroCCB)
}

func (c *CobrancalService) FindByNumParcela(numParcela int, numeroCCB int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByNumParcela(numParcela, numeroCCB)
}

func (c *CobrancalService) FindByDataVencimento(dataExpiracao string, numeroCCB int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByDataVencimento(dataExpiracao, numeroCCB)
}
