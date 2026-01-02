package service

import (
	"cobranca-bmp/config"
	"cobranca-bmp/models"
	"errors"
	"time"
)

type WebhookClient interface {
	RequestToWebhook(data any, url string) error
}

type WebhookService struct {
	client WebhookClient

	queue QueueProducer
	loc   *time.Location
}

func NewWebhookService(client WebhookClient, loc *time.Location) *WebhookService {
	return &WebhookService{
		client: client,
		loc:    loc,
	}
}

// Configurando o producer das filas
func (w *WebhookService) SetProducer(q any) error {
	if q == nil {
		return errors.New("producer não pode ser nulo")
	}
	queue, ok := q.(Queue)

	if !ok {
		return errors.New("producer deve implementar queue")
	}

	w.queue = queue
	return nil
}

func (w *WebhookService) RequestToWebhook(data models.WebhookTaskData) error {

	err := w.client.RequestToWebhook(data.Data, data.Url)
	if err != nil {
		data.SetTry()
		w.queue.Produce(config.WEBHOOK_QUEUE, data, data.Delay)
		return err
	}
	return nil

}

func (w *WebhookService) SendToDLQ(data any) {
	w.queue.Produce(config.DLQ_QUEUE, data, 0)

}
