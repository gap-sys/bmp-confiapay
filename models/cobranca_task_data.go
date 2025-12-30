package models

import (
	"cobranca-bmp/config"
	"fmt"
	"time"
)

type CobrancaTaskData struct {
	IdProposta             int                        `json:"idProposta"`
	WebhookUrl             string                     `json:"urlWebhook"`
	IdempotencyKey         string                     `json:"idempotencyKey,omitempty"`
	TimeoutRetries         int64                      `json:"timeoutRetries"`
	MultiplasCobrancas     bool                       `json:"multiplasCobrancas"`
	TipoCobranca           int                        `json:"tipoCobranca"`
	NumeroCCB              int                        `json:"numeroCCB"`
	RateLimitRetries       int64                      `json:"rateLimitRetries"`
	TimeoutDelay           time.Duration              `json:"timeoutDelay"`
	CurrentDelay           time.Duration              `json:"currentDelay"`
	Token                  string                     `json:"token,omitempty"`
	Status                 string                     `json:"status"`
	NumeroAcompanhamento   string                     `json:"numeroAcompanhamento"`
	GerarCobrancaInput     GerarCobrancaFrontendInput `json:"gerarCobrancaInput"`
	CancelamentoCobranca   CancelarCobrancaInput      `json:"cancelamentoCobranca"`
	ConsultarCobrancaInput ConsultarDetalhesInput     `json:"consultarCobrancaInput"`
	CalledAssync           bool                       `json:"calledAssync"`
}

func NewCobrancaTastkData(idempotencyKey int, status string, IdProposta int, numeroAcompanhamento string) CobrancaTaskData {
	var c = CobrancaTaskData{
		IdProposta:         IdProposta,
		RateLimitRetries:   config.RATE_LIMIT_MAX_RETRIES,
		TimeoutRetries:     config.TIMEOUT_MAX_RETRIES,
		CurrentDelay:       0,
		MultiplasCobrancas: false,
		CalledAssync:       false,
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
	}
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
	c.IdempotencyKey = fmt.Sprintf("%d:%d:%d", base, c.IdProposta)
}
