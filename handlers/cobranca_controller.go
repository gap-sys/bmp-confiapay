package handlers

import (
	"cobranca-bmp/cache"
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CobrancaCreditoPessoalController struct {
	cache           *cache.RedisCache
	cobrancaService CobrancaService
	loc             *time.Location
	webhookService  WebhookService
}

func NewCobrancaCreditoPessoalController(cache *cache.RedisCache, service CobrancaService, webhookService WebhookService, loc *time.Location) *CobrancaCreditoPessoalController {
	return &CobrancaCreditoPessoalController{
		cache:           cache,
		cobrancaService: service,
		loc:             loc,
		webhookService:  webhookService,
	}
}

func (a *CobrancaCreditoPessoalController) GetPrefix() string {
	return "/"
}

// Serão expostos apenas os endpoints necessários para atualizar dados bancários e realizar o envio de assinatura via certificadora.
func (a *CobrancaCreditoPessoalController) Route(r fiber.Router) {
	r.Post("/consulta", a.ConsultarCobrancas())
	r.Post("/cancelamento", a.CancelarCobranca())
	r.Post("/geracao", a.GerarCobranca())

}

// Gerar Cobranca godoc
//
//	@Summary		GerarCobranca Proposta.
//	@Description	Realiza a geracao de cobranças na API do BMP.
//	@Tags			CreditoPessoal
//	@Accept			json
//	@Produce		json
//	@Param			body	body	models.GerarCobrancaFrontendInput	true  "Dados da cobrança."
//	@Success		200		{object}	models.APIResponse
//	@Failure		422		{object}	models.APIError
//	@Router			/geracao [post]
func (a *CobrancaCreditoPessoalController) GerarCobranca() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input models.GerarCobrancaFrontendInput

		err := c.BodyParser(&input)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(models.NewAPIError(config.ERR_PARSE_STR, helpers.ParseJsonError(err.Error()), ""))
		}

		if err := input.Validate(); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err)
		}

		authPayload := models.AuthPayload{
			IdConvenio:       input.IdConvenio,
			IdSecuritizadora: input.IdSecuritizadora,
			Id:               strconv.Itoa(input.IdPropostaParcela),
		}

		var payload = models.NewCobrancaTastkData(a.cache.GenID(), config.STATUS_GERAR_COBRANCA, input.IdProposta, input.NumeroAcompanhamento, authPayload)
		payload.WebhookUrl = input.UrlWebhook
		payload.TipoCobranca = input.TipoCobranca
		payload.GerarCobrancaInput = input
		payload.IdPropostaParcela = input.IdPropostaParcela

		if err := payload.Validate(); err != nil {
			return c.Status(422).JSON(err)
		}

		//var resp = models.NewAPIError("", "A geração de cobranças entrou em processamento. Aguarde!", input.Dto.CodigoOperacao)
		//resp.HasError = false

		data, _, statusCode, err := a.cobrancaService.Cobranca(payload)
		if err != nil {
			return c.Status(statusCode).JSON(err)
		}

		return c.Status(statusCode).JSON(data)
	}
}

// CancelarCobranca godoc
//
//	@Summary		Cancelar Boleto.
//	@Description	Realiza o cancelamento de boletos.
//	@Tags			CreditoPessoal
//	@Accept			json
//	@Produce		json
//	@Param			body	body	models.CancelarCobrancaFrontendInput	true  "Dados da cobrança."
//	@Success		200		{object}	models.APIResponse
//	@Failure		422		{object}	models.APIError
//	@Router			/cancelamento [post]
func (a *CobrancaCreditoPessoalController) CancelarCobranca() fiber.Handler {

	return func(c *fiber.Ctx) error {
		var input models.CancelarCobrancaFrontendInput

		err := c.BodyParser(&input)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(models.NewAPIError(config.ERR_PARSE_STR, helpers.ParseJsonError(err.Error()), ""))
		}
		if err := input.Validate(); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err)
		}

		var bmpPayload = models.CancelarCobrancaInput{
			DTO: models.DtoCobranca{
				CodigoProposta: input.NumeroAcompanhamento,
				CodigoOperacao: strconv.Itoa(input.IdProposta),
				NroParcelas:    []int{1},
			},
			DTOCancelarCobrancas: models.DTOCancelarCobrancas{
				CodigosLiquidacoes: []string{input.CodigoLiquidacao},
			},
		}

		var authPayload = models.AuthPayload{
			IdConvenio:       input.IdConvenio,
			IdSecuritizadora: input.IdSecuritizadora,
			Id:               strconv.Itoa(input.IdPropostaParcela),
		}

		var cancelamentoTaskData = models.NewCobrancaTastkData(a.cache.GenID(), config.STATUS_CANCELAR_COBRANCA, input.IdProposta, input.NumeroAcompanhamento, authPayload)
		cancelamentoTaskData.CancelamentoCobranca = bmpPayload
		cancelamentoTaskData.CancelamentoData = input
		cancelamentoTaskData.IdPropostaParcela = input.IdPropostaParcela

		data, _, statusCode, err := a.cobrancaService.CancelarCobranca(cancelamentoTaskData)

		//var resp = models.NewAPIError("", "O cancelamento de cobranças entrou em processamento. Aguarde!", strconv.Itoa(input.IdProposta))
		if err != nil {
			return c.Status(statusCode).JSON(err)
		}
		//resp.HasError = false
		return c.Status(fiber.StatusOK).JSON(data)
	}
}

// Consultar Cobrancas godoc
//
//	@Summary		Consultar Cobranças.
//	@Description	Realiza a consulta de cobranças geradas para uma proposta específica.
//	@Tags			CreditoPessoal
//	@Accept			json
//	@Produce		json
//	@Param			body	body	models.ConsultaCobrancaFrontEndInput	true  "Identificador da proposta."
//	@Success		200		{object}	models.ConsultaCobrancaResponse
//	@Failure		422		{object}	models.APIError
//	@Router			/consulta [post]
func (a *CobrancaCreditoPessoalController) ConsultarCobrancas() fiber.Handler {

	return func(c *fiber.Ctx) error {
		var input models.ConsultaCobrancaFrontEndInput

		err := c.BodyParser(&input)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(models.NewAPIError(config.ERR_PARSE_STR, helpers.ParseJsonError(err.Error()), ""))
		}

		if err := input.Validate(); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err)
		}

		var parcelas = make([]int, 0)
		if input.Parcela > 0 {
			parcelas = append(parcelas, input.Parcela)
		}

		var authPayload = models.AuthPayload{
			IdConvenio:       input.IdConvenio,
			IdSecuritizadora: input.IdSecuritizadora,
			Id:               strconv.Itoa(input.IdProposta),
		}

		consultaTaskData := models.NewCobrancaTastkData(a.cache.GenID(), config.STATUS_CONSULTAR_COBRANCA, input.IdProposta, input.NumeroAcompanhamento, authPayload)
		consultaTaskData.ConsultarCobrancaInput =
			models.ConsultarDetalhesInput{
				DTO: models.DtoCobranca{CodigoProposta: input.NumeroAcompanhamento,
					CodigoOperacao: strconv.Itoa(input.IdProposta),
					NroParcelas:    parcelas,
				},
				DTOConsultaDetalhes: models.DTOConsultaDetalhes{
					TrazerBoleto:          true,
					TrazerAgendaDetalhada: true,
				},
			}

		data, _, statusCode, err := a.cobrancaService.Cobranca(consultaTaskData)

		if err != nil {
			return c.Status(statusCode).JSON(err)
		}

		return c.Status(fiber.StatusOK).JSON(data)
	}

}
