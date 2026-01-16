package models

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type CobrancaTaskData struct {
	IdProposta             int                            `json:"idProposta"`
	WebhookUrl             string                         `json:"urlWebhook"`
	IdempotencyKey         string                         `json:"idempotencyKey,omitempty"`
	IdPropostaParcela      int                            `json:"idPropostaParcela"`
	TimeoutRetries         int64                          `json:"timeoutRetries"`
	MultiplasCobrancas     bool                           `json:"multiplasCobrancas"`
	TipoCobranca           int                            `json:"tipoCobranca"`
	NumeroCCB              int                            `json:"numeroCCB"`
	NumeroBoleto           int                            `json:"numeroBoleto"`
	RateLimitRetries       int64                          `json:"rateLimitRetries"`
	ConsultaRetries        int                            `json:"consultaRetries"`
	TimeoutDelay           time.Duration                  `json:"timeoutDelay"`
	CurrentDelay           time.Duration                  `json:"currentDelay"`
	Token                  string                         `json:"token,omitempty"`
	Status                 string                         `json:"status"`
	NumeroAcompanhamento   string                         `json:"numeroAcompanhamento"`
	AuthPayload            AuthPayload                    `json:"authPayload"`
	GerarCobrancaInput     GerarCobrancaFrontendInput     `json:"gerarCobrancaInput"`
	CancelamentoData       CancelarCobrancaFrontendInput  `json:"cancelamentoData"`
	LancamentoParcela      LancamentoParcelaFrontendInput `json:"lancamentoParcela"`
	LancamentoParcelaInput LancamentoParcelaAPIInput      `json:"lancamentoParcelaInput"`
	CancelamentoCobranca   CancelarCobrancaInput          `json:"cancelamentoCobranca"`
	ConsultarCobrancaInput ConsultarDetalhesInput         `json:"consultarCobrancaInput"`
	CalledAssync           bool                           `json:"calledAssync"`
	ModoConsulta           string                         `json:"modoConsulta"`
	Updated                bool                           `json:"updated"`
	WhData                 map[string]any
	CobrancaDBInfo         CobrancaBMP `json:"cobrancaDBInfo"`
}

func NewCobrancaTastkData(idempotencyKey int, status string, IdProposta int, numeroAcompanhamento string, authPayload AuthPayload) *CobrancaTaskData {
	var c = &CobrancaTaskData{
		IdProposta:           IdProposta,
		NumeroAcompanhamento: numeroAcompanhamento,
		RateLimitRetries:     config.RATE_LIMIT_MAX_RETRIES,
		TimeoutRetries:       config.TIMEOUT_MAX_RETRIES,
		CurrentDelay:         0,
		MultiplasCobrancas:   config.GERAR_MULTIPLAS_COBRANCAS,
		CalledAssync:         false,
		AuthPayload:          authPayload,
		Status:               status,
		WhData:               nil,
		ModoConsulta:         config.CONSULTA_DETALHADA,
		ConsultaRetries:      10,
		Updated:              false,
	}
	c.GenIdempotencyKey(idempotencyKey)

	return c
}

// Seta as tentativas
func (a *CobrancaTaskData) SetTry(interval time.Duration, status string) {

	switch status {
	case config.API_STATUS_RATE_LIMIT:
		a.RateLimitRetries--
		a.CurrentDelay = config.RATE_LIMIT_DELAY
	case config.API_STATUS_TIMEOUT:
		a.TimeoutDelay += interval
		a.CurrentDelay = a.TimeoutDelay
		a.TimeoutRetries--

	default:
		a.CurrentDelay += interval

	}

	if a.Status == config.STATUS_CONSULTAR_COBRANCA {
		a.ConsultaRetries--
		return
	}
}

func (a *CobrancaTaskData) SwitchCobrancaMode() error {
	if a.ModoConsulta == config.CONSULTA_BOLETO {
		a.ModoConsulta = config.CONSULTA_DETALHADA
	} else if a.ModoConsulta == config.CONSULTA_DETALHADA && a.NumeroBoleto > 0 {
		a.ModoConsulta = config.CONSULTA_BOLETO
	}
	return nil
}

// Reseta todas as tentativas e o delay.
func (a *CobrancaTaskData) Reset() {
	a.TimeoutDelay = config.INITIAL_DELAY
	a.RateLimitRetries = config.RATE_LIMIT_MAX_RETRIES
	a.TimeoutRetries = config.TIMEOUT_MAX_RETRIES
	a.CurrentDelay = a.TimeoutDelay
}

func (a *CobrancaTaskData) Validate() error {
	return nil
}

func (c *CobrancaTaskData) GenIdempotencyKey(base int) {
	c.IdempotencyKey = helpers.GenerateIdempotencyKey(fmt.Sprintf("%d:%d:%d", base, c.IdProposta, rand.NewSource(10).Int63()))
}

func (c *CobrancaTaskData) FormatLancamento() LancamentoParcelaAPIInput {

	var bmpPayload = LancamentoParcelaAPIInput{
		DTO: DTOPropostaParcela{
			CodigoOperacao: strconv.Itoa(c.IdProposta),
			CodigoProposta: c.LancamentoParcela.NumeroAcompanhamento,
			NroParcela:     c.LancamentoParcela.NumeroParcela,
		},
		DTOLancamentoParcela: DTOLancamentoParcela{
			DtPagamento:             c.LancamentoParcela.DataLancamento,
			PermiteDescapitalizacao: false,
		},
	}

	switch c.LancamentoParcela.Operacao {
	case "S":

		bmpPayload.DTOLancamentoParcela.VlrPagamento = 0.01
		if c.LancamentoParcela.ValorDesconto > 0 {
			bmpPayload.DTOLancamentoParcela.VlrDesconto = helpers.Round(-1*(c.LancamentoParcela.ValorDesconto+0.01), 2)
		} else {
			bmpPayload.DTOLancamentoParcela.VlrDesconto = helpers.Round(c.LancamentoParcela.ValorDesconto-0.01, 2)
		}

		bmpPayload.DTOLancamentoParcela.DescricaoLancamento = "S"

	case "P":
		bmpPayload.DTOLancamentoParcela.VlrPagamento = c.LancamentoParcela.ValorLancamento
		bmpPayload.DTOLancamentoParcela.VlrDesconto = c.LancamentoParcela.ValorDesconto
		bmpPayload.DTOLancamentoParcela.DescricaoLancamento = "P"

	}

	return bmpPayload
}
