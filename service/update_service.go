package service

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"errors"
	"log/slog"
	"time"
)

type QueueProducer interface {
	Produce(queue string, data any, delay time.Duration) error
}

type UpdateService struct {
	queue    QueueProducer
	logger   *slog.Logger
	loc      *time.Location
	convenio int
}

func NewUpdateService(logger *slog.Logger, loc *time.Location, convenio int) *UpdateService {
	return &UpdateService{
		logger:   logger,
		convenio: convenio,
		loc:      loc,
	}
}

func (u *UpdateService) String() string {
	switch u.convenio {
	case config.FGTS_CONVENIO:
		return "FGTS"
	case config.INSSCP_CONVENIO:
		return "INSSCP"
	case config.CREDITO_TRABALHADOR_CONVENIO:
		return "Crédito Trabalhador"
	case config.CREDITO_PESSOAL_CONVENIO:
		return "Crédito Pessoal"
	default:
		return ""
	}
}

func (u *UpdateService) SetProducer(producer any) error {
	if producer == nil {
		return errors.New("producer não pode ser nulo")
	}
	queue, ok := producer.(QueueProducer)
	if ok {
		u.queue = queue
		return nil
	}
	return errors.New("producer inválido")

}

func (u *UpdateService) UpdateAssync(data models.UpdateDBData) (bool, error) {
	helpers.LogInfo(context.Background(), u.logger, u.loc, "db", "", "Entrou em update assync, convenio="+u.String(), data)
	var err error
	var noConn bool

	switch data.Action {
	case "":
	default:
		return false, errors.New("ação de update inválida")
	}
	return noConn, err
}
