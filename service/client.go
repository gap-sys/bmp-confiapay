package service

import (
	"cobranca-bmp/models"
)

type CobrancaClient interface {
	Auth(data models.AuthPayload) (string, error)
	GerarCobrancaParcela(payload models.CobrancaUnicaInput, token string, idempotencyKey string) (models.GerarCobrancaResponse, int, string, error)
	GerarCobrancaParcelasMultiplas(payload models.MultiplasCobrancasInput, token string, idempotencyKey string) (models.GerarCobrancaResponse, int, string, error)
	CancelarCobranca(payload models.CancelarCobrancaInput, token string, idempotencyKey string) (any, int, string, error)
	LancamentoParcela(payload models.LancamentoParcelaAPIInput, token string, idempotencyKey string) (any, int, string, error)
	ConsultarCobranca(payload models.ConsultarDetalhesInput, token string, idempotencyKey string) (models.ConsultaCobrancaResponse, int, string, error)
	ConsultarBoleto(payload models.ConsultaBoletoInput, token string, idempotencyKey string) (models.BoletoConsultado, int, string, error)
}
