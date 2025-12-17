package handlers

import (
	"cobranca-bmp/models"
)

type PropostaService interface {
	Update(data models.UpdateDBData)
	FindFinalStatus(idProposta int) (int, error)
	FindTipoData(idProposta int, idTipoData ...int) (int, error)
	DeleteFluxo(idProposta, idTipoData int) (bool, error)
	UpdateDigitacao(data models.UpdateDBData, calledAssync bool) (bool, error)
	FindPropostaInfo(id int) (models.PropostaInfo, error)
	UpdateEstorno(idProposta int, data any, calledAssync bool) (bool, error)
	UpdateParcelas(payload models.PropostaParcelaUpdate, calledAssync bool) (bool, error)
}

type WebhookService interface {
	RequestToWebhook(data models.WebhookTaskData) error
	SendToDLQ(data any)
}

type TokenService interface {
	GenerateToken(clientId, privateKey string) (string, string, error)
}

type CobrancaCreditoPessoalService interface {
	Auth(key string, client string) (string, error)
	ConsultarCobranca(payload models.ConsultarDetalhesInput) (models.ConsultaCobrancaResponse, int, string, error)
	Cobranca(payload models.CobrancaTaskData) (any, string, int, error)
	CancelarCobranca(payload models.CobrancaTaskData) (any, string, int, error)
	SendToDLQ(data any) error
}
