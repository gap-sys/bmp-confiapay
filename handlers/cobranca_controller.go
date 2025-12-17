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
	return "/cobranca"
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
//	@Param			body	body	models.BuscarPropostaDb	true  "Id da proposta que se deseja realizar a cobrança."
//	@Success		200		{object}	models.APIResponse
//	@Failure		422		{object}	models.APIError
//	@Router			/cobranca/geracao [post]
func (a *CobrancaCreditoPessoalController) GerarCobranca() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input models.CobrancaUnicaFrontendInput

		err := c.BodyParser(&input)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(models.NewAPIError(config.ERR_PARSE_STR, helpers.ParseJsonError(err.Error()), ""))
		}

		if err := input.Validate(); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err)
		}

		IdProposta, _ := strconv.Atoi(input.Dto.CodigoOperacao)

		var payload = models.NewCobrancaTastkData(a.cache.GenID(), 1, IdProposta, input.Dto.CodigoProposta)
		payload.WebhookUrl = input.UrlWebhook
		payload.Convenio = input.IdCOnvenio
		payload.CobrancaUnicaFrontendInput = input

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
//	@Param			body	body	models.BuscarPropostaDb	true  "Identificador da proposta para ser buscado os dados bancários."
//	@Success		200		{object}	models.APIResponse
//	@Failure		422		{object}	models.APIError
//	@Router			/cobranca/cancelamento [post]
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
				NroParcelas:    input.Parcelas,
			},
			DTOCancelarCobrancas: models.DTOCancelarCobrancas{
				CodigosLiquidacoes: []string{input.CodigoLiquidacao},
			},
		}

		var cancelamentoTaskData = models.NewCobrancaTastkData(a.cache.GenID(), 1, input.IdProposta, input.NumeroAcompanhamento)
		cancelamentoTaskData.CancelamentoCobranca = bmpPayload

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
//	@Router			/cobrancas [post]
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

		data, statusCode, _, err := a.cobrancaService.ConsultarCobranca(models.ConsultarDetalhesInput{
			DTO: models.DtoCobranca{CodigoProposta: input.NumeroAcompanhamento,
				CodigoOperacao: strconv.Itoa(input.IdProposta),
				NroParcelas:    input.Parcelas,
			},
			DTOConsultaDetalhes: models.DTOConsultaDetalhes{
				TrazerBoleto:          true,
				TrazerAgendaDetalhada: true,
			},
		}, input.Convenio)

		if err != nil {
			return c.Status(statusCode).JSON(err)
		}

		return c.Status(fiber.StatusOK).JSON(data.Result)
	}

}
