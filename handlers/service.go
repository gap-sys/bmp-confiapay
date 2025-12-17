package handlers

import (
	"cobranca-bmp/models"
)

type WebhookService interface {
	RequestToWebhook(data models.WebhookTaskData) error
	SendToDLQ(data any)
}

type TokenService interface {
	GenerateToken(clientId, privateKey string) (string, string, error)
}

type CobrancaCreditoPessoalService interface {
	//	Auth(key string, client string) (string, error)
	ConsultarCobranca(payload models.ConsultarDetalhesInput, convenio int) (models.ConsultaCobrancaResponse, int, string, error)
	Cobranca(payload models.CobrancaTaskData) (any, string, int, error)
	CancelarCobranca(payload models.CobrancaTaskData) (any, string, int, error)
	SendToDLQ(data any) error
}
