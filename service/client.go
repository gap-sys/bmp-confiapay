package service

import (
	"cobranca-bmp/models"
)

type CreditoPessoalClient interface {
	DigitacaoCreditoPessoalClient
	CobrancaCreditoPessoalClient
}

type DigitacaoCreditoPessoalClient interface {
	BuscarProposta(param models.BuscarProposta, token, idempotencyKey string) (models.PropostaAPIResponse, string, int, error)
	Auth(key, client string, expire bool, privateKey string) (string, error)
}

type CobrancaCreditoPessoalClient interface {
	Auth(key, client string, expire bool, privateKey string) (string, error)
	GerarCobrancaParcela(payload models.CobrancaUnicaInput, token string, idempotencyKey string) (models.GerarCobrancaResponse, int, string, error)
	GerarCobrancaParcelasMultiplas(payload models.MultiplasCobrancasInput, token string, idempotencyKey string) (models.GerarCobrancaResponse, int, string, error)
	CancelarCobranca(payload models.CancelarCobrancaInput, token string, idempotencyKey string) (any, int, string, error)
	ConsultarCobranca(payload models.ConsultarDetalhesInput, token string, idempotencyKey string) (models.ConsultaCobrancaResponse, int, string, error)
}
