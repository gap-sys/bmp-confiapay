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

type CobrancaService interface {
	//	Auth(key string, client string) (string, error)
	//	ConsultarCobranca(payload *models.CobrancaTaskData) (models.ConsultaCobrancaResponse, string, int, error)
	Cobranca(payload *models.CobrancaTaskData) (any, string, int, error)
	//CancelarCobranca(payload *models.CobrancaTaskData) (any, string, int, error)
	FindByCodLiquidacao(codigoLiquidacao string, numeroCCB int) (models.CobrancaBMP, error)
	FindByNumParcela(numParcela int, numeroCCB int) (models.CobrancaBMP, error)
	FindByDataVencimento(dataExpiracao string, numeroCCB int) (models.CobrancaBMP, error)
	UpdateCodLiquidacao(idPropostaParcela int, codigoLiquidacao string) (bool, error)
	FindByIdPropostaParcela(idPropostaParcela int) (models.CobrancaBMP, error)
	SendToDLQ(data any) error
}
