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
	operations        map[string]string
}

// Configurando o producer das filas
func (c *CobrancalService) SetProducer(q any) error {
	if q == nil {
		return errors.New("producer não pode ser nulo")
	}
	queue, ok := q.(Queue)

	if !ok {
		return errors.New("producer deve implementar queue")
	}
	c.queue = queue
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
		operations: map[string]string{
			config.STATUS_GERAR_COBRANCA:     "Geração de cobrança",
			config.STATUS_CANCELAR_COBRANCA:  "Cancelamento de cobrança",
			config.STATUS_LANCAMENTO_PARCELA: "Lançamento na parcela",
			config.STATUS_CONSULTAR_COBRANCA: "Consulta de cobrança",
		},
	}
}

func (c CobrancalService) Auth(data models.AuthPayload) (string, error) {
	return c.client.Auth(data)
}

func (c *CobrancalService) Cobranca(payload *models.CobrancaTaskData) (any, string, int, error) {
	if payload.Token == "" {
		token, err := c.Auth(payload.AuthPayload)
		if err != nil {
			return nil, "", 401, err
		}
		payload.Token = token

	}

	switch payload.Status {
	case config.STATUS_CANCELAR_COBRANCA:
		data, status, statusCode, err := c.CancelarCobranca(payload)

		if err != nil {
			if payload.CalledAssync {
				err, _ := err.(models.APIError)
				return c.HandleErrorCobranca(status, statusCode, payload, err)
			}

		}
		return data, status, statusCode, err

	case config.STATUS_CONSULTAR_COBRANCA:
		switch payload.ModoConsulta {
		case config.CONSULTA_DETALHADA:
			data, status, statusCode, err := c.ConsultarCobranca(payload, true)

			if err != nil {
				if payload.CalledAssync {
					err, _ := err.(models.APIError)
					return c.HandleErrorCobranca(status, statusCode, payload, err)
				}

			}
			return data, status, statusCode, err

		case config.CONSULTA_BOLETO:
			data, status, statusCode, err := c.ConsultarBoleto(payload, true)

			if err != nil {
				if payload.CalledAssync {
					err, _ := err.(models.APIError)
					return c.HandleErrorCobranca(status, statusCode, payload, err)
				}

			}
			return data, status, statusCode, err
		}

	case config.STATUS_GERAR_COBRANCA:
		return c.GerarCobranca(payload)

	case config.STATUS_LANCAMENTO_PARCELA:
		data, status, statusCode, err := c.LancamentoParcela(payload)
		if err != nil {
			if payload.CalledAssync {
				err, _ := err.(models.APIError)
				return c.HandleErrorCobranca(status, statusCode, payload, err)
			}

		}

		return data, status, statusCode, err

	default:
		return nil, "", 500, models.NewAPIError("", "Status inválido", payload.AuthPayload.Id)

	}
	return nil, "", 200, nil

}

func (c *CobrancalService) GerarCobranca(payload *models.CobrancaTaskData) (any, string, int, error) {
	var status string
	var errCobrancas error
	var statusCode int
	var data models.GerarCobrancaResponse

	if !payload.Updated {

		_, err := c.updateService.UpdateGeracaoParcela(models.UpdateDbData{
			GeracaoParcela: &payload.GerarCobrancaInput,
			Action:         "update_geracao",
		}, true)
		if err != nil {
			return c.HandleErrorCobranca(config.API_STATUS_ERR, 500, payload, models.NewAPIError("", "Não foi possível gravar cobrancas na base de dados: "+err.Error(), strconv.Itoa(payload.GerarCobrancaInput.IdProposta)))
		}

		payload.Updated = true

	}

	payload.CobrancaDBInfo.IdFormaCobranca = payload.GerarCobrancaInput.TipoCobranca

	cobrancaInfo, _ := c.parcelaRepository.FindByIdPropostaParcela(payload.GerarCobrancaInput.IdPropostaParcela)
	if cobrancaInfo.CodigoLiquidacao != "" {
		payload.ConsultarCobrancaInput = models.ConsultarDetalhesInput{
			DTO: models.DtoCobranca{
				CodigoProposta: payload.GerarCobrancaInput.NumeroAcompanhamento,
				CodigoOperacao: strconv.Itoa(payload.GerarCobrancaInput.IdProposta),
				NroParcelas:    []int{payload.GerarCobrancaInput.NumeroParcela},
			},
			DTOConsultaDetalhes: models.DTOConsultaDetalhes{
				TrazerBoleto: true,
			},
		}
		payload.Status = config.STATUS_CONSULTAR_COBRANCA
		payload.NumeroBoleto = cobrancaInfo.NumeroBoleto
		payload.CobrancaDBInfo.CodigoLiquidacao = cobrancaInfo.CodigoLiquidacao
		payload.CobrancaDBInfo.IdPropostaParcela = payload.GerarCobrancaInput.IdPropostaParcela
		payload.IdPropostaParcela = payload.GerarCobrancaInput.IdPropostaParcela
		payload.SwitchCobrancaMode()

		return c.Cobranca(payload)

	}

	if payload.MultiplasCobrancas {

		data, statusCode, status, errCobrancas = c.GerarCobrancaParcelasMultiplas(payload)
	} else {
		data, statusCode, status, errCobrancas = c.GerarCobrancaParcela(payload)
	}

	if errCobrancas != nil {
		return c.HandleErrorCobranca(status, statusCode, payload, errCobrancas.(models.APIError))
	} else if len(data.Cobrancas) < 1 {
		return c.HandleErrorCobranca(status, statusCode, payload, models.NewAPIError("", "Cobrancas não geradas", strconv.Itoa(payload.GerarCobrancaInput.IdProposta)))

	}

	c.updateService.UpdateCodLiquidacao(models.UpdateDbData{
		IdPropostaParcela: payload.IdPropostaParcela,
		CodigoLiquidacao:  data.Cobrancas[0].CodigoLiquidacao,
		Action:            "update_codigo_liquidacao",
	}, false)

	var cobrancaGeradaInfo = map[string]any{
		"id_proposta":         payload.GerarCobrancaInput.IdProposta,
		"id_proposta_parcela": payload.IdPropostaParcela,
		"codigo_liquidacao":   data.Cobrancas[0].CodigoLiquidacao,
	}

	helpers.LogInfo(c.ctx, c.logger, c.loc, "cobranca service", "200", "Cobrança gerada", cobrancaGeradaInfo)

	return data, "", 200, nil
}

func (c CobrancalService) SendToDLQ(data any) error {
	return c.queue.Produce(config.DLQ_QUEUE, data, 0)

}

func (c *CobrancalService) GerarCobrancaParcela(payload *models.CobrancaTaskData) (models.GerarCobrancaResponse, int, string, error) {
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

	data, statusCode, status, err := c.client.GerarCobrancaParcela(payloadBMP, payload.Token, payload.IdempotencyKey)

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
func (c *CobrancalService) GerarCobrancaParcelasMultiplas(payload *models.CobrancaTaskData) (models.GerarCobrancaResponse, int, string, error) {
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

	data, statusCode, status, err := c.client.GerarCobrancaParcelasMultiplas(payloadBMP, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := c.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, statusCode, status, err
		}
		payload.Token = token

		data, statusCode, status, err = c.client.GerarCobrancaParcelasMultiplas(payloadBMP, payload.Token, payload.IdempotencyKey)
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

func (c *CobrancalService) CancelarCobranca(payload *models.CobrancaTaskData) (any, string, int, error) {

	if !payload.Updated {
		_, err := c.updateService.UpdateCancelamentoParcela(models.UpdateDbData{
			CancelamentoCobranca: &payload.CancelamentoData,
			CodigoLiquidacao:     payload.CancelamentoCobranca.DTOCancelarCobrancas.CodigosLiquidacoes[0],
			Action:               "update_cancelamento",
		}, true)

		if err != nil {
			return c.HandleErrorCobranca(config.API_STATUS_ERR, 500, payload, models.NewAPIError("", "Não foi possível gravar cobrancas na base de dados: "+err.Error(), strconv.Itoa(payload.GerarCobrancaInput.IdProposta)))
		}
		payload.Updated = true
	}

	data, statusCode, status, err := c.client.CancelarCobranca(payload.CancelamentoCobranca, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := c.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, status, statusCode, err
		}

		payload.Token = token
		data, statusCode, status, err = c.client.CancelarCobranca(payload.CancelamentoCobranca, payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		errApi, ok := err.(models.APIError)
		if !ok {
			errApi = models.NewAPIError("", "Não foi possível buscar token", strconv.Itoa(payload.IdProposta))
		}

		if status == config.API_STATUS_RATE_LIMIT || status == config.API_STATUS_TIMEOUT {
			payload.CalledAssync = true
		}

		return nil, status, statusCode, errApi
	}

	return data, status, statusCode, nil
}

func (c *CobrancalService) LancamentoParcela(payload *models.CobrancaTaskData) (any, string, int, error) {

	if !payload.Updated {
		_, err := c.updateService.UpdateLancamentoParcela(models.UpdateDbData{
			LancamentoParcela: &payload.LancamentoParcela,
			Action:            "update_lancamento",
		}, true)

		if err != nil {
			return c.HandleErrorCobranca(config.API_STATUS_ERR, 500, payload, models.NewAPIError("", "Não foi possível gravar cobrancas na base de dados: "+err.Error(), strconv.Itoa(payload.GerarCobrancaInput.IdProposta)))
		}

		payload.Updated = true
	}

	data, statusCode, status, err := c.client.LancamentoParcela(payload.FormatLancamento(), payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := c.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, status, statusCode, err
		}

		payload.Token = token
		data, statusCode, status, err = c.client.LancamentoParcela(payload.FormatLancamento(), payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		errApi, ok := err.(models.APIError)
		if !ok {
			errApi = models.NewAPIError("", "Não foi possível buscar token", strconv.Itoa(payload.IdProposta))
		}

		if status == config.API_STATUS_RATE_LIMIT || status == config.API_STATUS_TIMEOUT {
			payload.CalledAssync = true
		}

		return nil, status, statusCode, errApi
	}

	return data, status, statusCode, nil
}

func (c *CobrancalService) ConsultarCobranca(payload *models.CobrancaTaskData, callBack bool) (models.ConsultaCobrancaResponse, string, int, error) {
	data, statusCode, status, err := c.client.ConsultarCobranca(payload.ConsultarCobrancaInput, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := c.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, status, statusCode, err
		}

		payload.Token = token
		data, statusCode, status, err = c.client.ConsultarCobranca(payload.ConsultarCobrancaInput, payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		return data, status, statusCode, err
	}

	if len(data.Parcelas) < 1 {
		return models.ConsultaCobrancaResponse{}, "", 404, models.NewAPIError("", "Parcelas não encontradas", strconv.Itoa(payload.IdPropostaParcela))
	}

	var whData = make(map[string]any)
	whData["id_proposta_parcela"] = payload.CobrancaDBInfo.IdPropostaParcela
	whData["codigo_liquidacao"] = payload.CobrancaDBInfo.CodigoLiquidacao
	whData["operacao"] = "R"

	switch payload.CobrancaDBInfo.IdFormaCobranca {
	case config.TIPO_COBRANCA_BOLETO, config.TIPO_COBRANCA_BOLETOPIX:
		if len(data.Parcelas[0].Boletos) < 1 {
			return models.ConsultaCobrancaResponse{}, "", 404, models.NewAPIError("", "Boleto não encontrado", strconv.Itoa(payload.IdPropostaParcela))
		} else if data.Parcelas[0].Boletos[0].UrlImpressao == "" {
			return models.ConsultaCobrancaResponse{}, "", 404, models.NewAPIError("", "Url de boleto não retornada", strconv.Itoa(payload.IdPropostaParcela))

		}

		whData["boleto"] = data.Parcelas[0].Boletos

	case config.TIPO_COBRANCA_PIX:
		if len(data.Parcelas[0].Pix) < 1 {
			return models.ConsultaCobrancaResponse{}, "", 404, models.NewAPIError("", "Pix não encontrado", strconv.Itoa(payload.IdPropostaParcela))
		} else if data.Parcelas[0].Pix[0].Imagem != "" || data.Parcelas[0].Pix[0].Emv != "" {
			return models.ConsultaCobrancaResponse{}, "", 404, models.NewAPIError("", "QR code de pix não retornado", strconv.Itoa(payload.IdPropostaParcela))

		}

		whData["pix"] = data.Parcelas[0].Pix
	}

	payload.WhData = whData

	if payload.WebhookUrl != "" && callBack {
		c.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "consulta-cobranca"))
	}

	return data, status, statusCode, nil
}

func (c *CobrancalService) ConsultarBoleto(payload *models.CobrancaTaskData, callBack bool) (models.BoletoConsultado, string, int, error) {
	var consultaInput = models.ConsultaBoletoInput{
		DTO: models.ConsultaBoletoDTO{
			CodigoProposta: payload.NumeroAcompanhamento,
			CodigoOperacao: strconv.Itoa(payload.IdProposta),
		},
		DTOConsultaBoletos: models.ConsultaBoletosFiltroDTO{
			NumerosBoletos: []int{payload.NumeroBoleto},
		},
	}

	data, statusCode, status, err := c.client.ConsultarBoleto(consultaInput, payload.Token, payload.IdempotencyKey)

	if status == config.API_STATUS_UNAUTHORIZED {
		token, authErr := c.Auth(payload.AuthPayload)
		if authErr != nil {
			return data, status, statusCode, err
		}

		payload.Token = token
		data, statusCode, status, err = c.client.ConsultarBoleto(consultaInput, payload.Token, payload.IdempotencyKey)
	}

	if err != nil {
		return data, status, statusCode, err
	}

	if len(data.Boletos) < 1 {
		return models.BoletoConsultado{}, "", 404, models.NewAPIError("", "Boleto não encontrado", strconv.Itoa(payload.IdPropostaParcela))
	}

	var whData = make(map[string]any)
	whData["id_proposta_parcela"] = payload.CobrancaDBInfo.IdPropostaParcela
	whData["codigo_liquidacao"] = payload.CobrancaDBInfo.CodigoLiquidacao
	whData["operacao"] = "R"
	whData["boleto"] = data.Boletos[0].FormatToWebhook()

	payload.WhData = whData

	if payload.WebhookUrl != "" && callBack {
		c.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "consulta-boleto"))
	}

	return data, status, statusCode, nil
}

func (c *CobrancalService) FindByCodLiquidacao(codigoLiquidacao string, numeroCCB int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByCodLiquidacao(codigoLiquidacao, numeroCCB)
}

func (c *CobrancalService) UpdateCodLiquidacao(idPropostaParcela int, codigoLiquidacao string) (bool, error) {
	return c.updateService.UpdateCodLiquidacao(models.UpdateDbData{
		CodigoLiquidacao:  codigoLiquidacao,
		IdPropostaParcela: idPropostaParcela,
		Action:            "update_codigo_liquidacao",
	}, false)
}

func (c *CobrancalService) UpdateNumeroBoleto(idPropostaParcela, numeroBoleto int) (bool, error) {
	return c.updateService.UpdateNumeroBoleto(models.UpdateDbData{
		IdPropostaParcela: idPropostaParcela,
		NumeroBoleto:      numeroBoleto,
		Action:            "update_numero_boleto",
	}, false)
}

func (c *CobrancalService) FindByNumParcela(numParcela int, numeroCCB int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByNumParcela(numParcela, numeroCCB)
}

func (c *CobrancalService) FindByDataVencimento(dataExpiracao string, numeroCCB int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByDataVencimento(dataExpiracao, numeroCCB)
}

// Realiza o tratamento de erros nas cobranças.
func (c *CobrancalService) HandleErrorCobranca(status string, statusCode int, payload *models.CobrancaTaskData, errAPI models.APIError) (any, string, int, error) {
	operation := c.operations[payload.Status]
	logMsg := fmt.Sprintf("Erro ao realizar %s", operation)
	helpers.LogError(c.ctx, c.logger, c.loc, "cobranca service", strconv.Itoa(statusCode), logMsg, errAPI, errAPI)

	if errAPI.ID <= 0 {
		errAPI.ID = payload.IdPropostaParcela
	}

	if len(errAPI.Messages) < 1 {
		errAPI.Messages = []models.APIMessage{
			{
				Description: errAPI.Error(),
			},
		}
	}

	if errAPI.Messages[0].Code != "" {
		errAPI.Msg = fmt.Sprintf("%s(%s)", errAPI.Messages[0].Description, errAPI.Messages[0].Code)
	}

	if statusCode == 500 || status == config.API_STATUS_TIMEOUT || status == config.API_STATUS_RATE_LIMIT {
		errAPI.Msg = fmt.Sprintf("%s indisponível(%s)", operation, errAPI.Messages[0].Code)

	}

	if payload.Status == config.STATUS_CONSULTAR_COBRANCA {
		payload.SetTry(config.TIMEOUT_DELAY, status)
		payload.SwitchCobrancaMode()
		payload.GenIdempotencyKey(c.cache.GenID())
		c.queue.Produce(config.CONSULTA_QUEUE, payload, payload.CurrentDelay)
		return nil, status, statusCode, errAPI
	}

	if (status == config.API_STATUS_TIMEOUT && payload.TimeoutRetries > 1) || (status == config.API_STATUS_RATE_LIMIT && payload.RateLimitRetries > 1) {
		payload.SetTry(config.TIMEOUT_DELAY, status)
		if status == config.API_STATUS_TIMEOUT {
			payload.GenIdempotencyKey(c.cache.GenID())
		}
		payload.CalledAssync = true

		err := c.queue.Produce(config.COBRANCA_QUEUE, payload, payload.CurrentDelay)
		if err != nil {

			return nil, config.API_STATUS_ERR, 500, models.NewAPIError("", "Erro ao enviar dados para fila: "+err.Error(), strconv.Itoa(payload.IdProposta))
		}

		var resp = models.NewAPIError("", fmt.Sprintf("%s entrou em fila de processamento. Aguarde!", c.operations[payload.Status]), strconv.Itoa(payload.IdProposta))
		resp.HasError = false
		return resp, "", 200, nil

	}

	if payload.CalledAssync {
		var whData = make(map[string]any)

		switch payload.Status {
		case config.STATUS_GERAR_COBRANCA:
			whData["has_error"] = true
			whData["msg"] = errAPI.Msg
			whData["id_proposta_parcela"] = payload.IdPropostaParcela
			whData["operacao"] = "R"

			c.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "cobranca service"))

		case config.STATUS_CANCELAR_COBRANCA:
			whData["has_error"] = true
			whData["msg"] = errAPI.Msg
			whData["id_proposta_parcela"] = payload.IdPropostaParcela
			whData["operacao"] = "C"

			c.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "cobranca service"))

		case config.STATUS_LANCAMENTO_PARCELA:
			var whData = make(map[string]any)
			whData["has_error"] = true
			whData["msg"] = errAPI.Msg
			whData["id_proposta_parcela"] = payload.IdPropostaParcela
			whData["operacao"] = payload.LancamentoParcela.Operacao
			c.webhookService.RequestToWebhook(models.NewWebhookTaskData(payload.WebhookUrl, whData, "cobranca service"))

		}

	}

	return nil, status, statusCode, errAPI

}

func (c *CobrancalService) FindByIdPropostaParcela(idPropostaParcela int) (models.CobrancaBMP, error) {
	return c.parcelaRepository.FindByIdPropostaParcela(idPropostaParcela)
}
