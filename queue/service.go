package queue

import "cobranca-bmp/models"

// Service é a interface que define os métodos que um serviço de fila deve implementar.
type Service interface {
	SetProducer(any) error
}

// dbService é a interface que define os métodos que um serviço de banco de dados deve implementar.
type dbService interface {
	Service
	UpdateAssync(data models.UpdateDbData) (bool, error)
}

// WebhookService é a interface que define os métodos que um serviço de webhook deve implementar.
type WebhookService interface {
	Service
	RequestToWebhook(data models.WebhookTaskData) error
}

type CobrancaCreditoPessoalService interface {
	Service
	Cobranca(payload *models.CobrancaTaskData) (any, string, int, error)
}
