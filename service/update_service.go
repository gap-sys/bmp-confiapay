package service

import (
	"cobranca-bmp/config"
	"cobranca-bmp/models"
	"errors"
	"log/slog"
	"time"
)

type QueueProducer interface {
	Produce(queue string, data any, delay time.Duration) error
}

type UpdateService struct {
	queue             QueueProducer
	logger            *slog.Logger
	loc               *time.Location
	parcelaRepository ParcelaRepository
}

func NewUpdateService(logger *slog.Logger, loc *time.Location, parcelaRepository ParcelaRepository) *UpdateService {
	return &UpdateService{
		logger:            logger,
		loc:               loc,
		parcelaRepository: parcelaRepository,
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

func (u *UpdateService) UpdateAssync(data models.UpdateDbData) (bool, error) {
	var err error
	var noConn bool

	switch data.Action {
	case "update_cod_liquidacao":
		noConn, err = u.UpdateCodLiquidacao(data, true)

	case "update_geracao":
		noConn, err = u.UpdateGeracaoParcela(data, true)

	case "update_cancelamento":
		noConn, err = u.UpdateGeracaoParcela(data, true)
	default:
		return false, errors.New("ação de update inválida")
	}
	return noConn, err
}

func (u *UpdateService) UpdateCodLiquidacao(data models.UpdateDbData, calledAssync bool) (bool, error) {

	noConn, err := u.parcelaRepository.UpdateCodLiquidacao(data.IdPropostaParcela, data.CodigoLiquidacao)
	if err != nil {
		if !noConn {
			var dlqData = models.DLQData{
				Contexto: "db",
				Payload:  nil,
				Mensagem: "falha ao realizar update de código liquidação",
				Erro:     err.Error(),
				Time:     time.Now().In(u.loc),
			}
			u.queue.Produce(config.DLQ_QUEUE, dlqData, 0)
		} else {
			if !calledAssync {
				u.queue.Produce(config.DB_QUEUE, data, config.DB_QUEUE_DELAY)
			}
		}
		return noConn, err
	}
	return false, nil
}

func (u *UpdateService) UpdateGeracaoParcela(data models.UpdateDbData, calledAssync bool) (bool, error) {

	noConn, err := u.parcelaRepository.UpdateGeracaoCobranca(*data.GeracaoParcela, data.CodigoLiquidacao)
	if err != nil {
		if !noConn {
			var dlqData = models.DLQData{
				Contexto: "db",
				Payload:  nil,
				Mensagem: "falha ao realizar update de código liquidação",
				Erro:     err.Error(),
				Time:     time.Now().In(u.loc),
			}
			u.queue.Produce(config.DLQ_QUEUE, dlqData, 0)
		} else {
			if !calledAssync {
				u.queue.Produce(config.DB_QUEUE, data, config.DB_QUEUE_DELAY)
			}
		}
		return noConn, err
	}
	return false, nil
}

func (u *UpdateService) UpdateCancelamentoParcela(data models.UpdateDbData, calledAssync bool) (bool, error) {

	noConn, err := u.parcelaRepository.UpdateCancelamentoCobranca(*data.CancelamentoCobranca, data.CodigoLiquidacao)
	if err != nil {
		if !noConn {
			var dlqData = models.DLQData{
				Contexto: "db",
				Payload:  nil,
				Mensagem: "falha ao realizar update de código liquidação",
				Erro:     err.Error(),
				Time:     time.Now().In(u.loc),
			}
			u.queue.Produce(config.DLQ_QUEUE, dlqData, 0)
		} else {
			if !calledAssync {
				u.queue.Produce(config.DB_QUEUE, data, config.DB_QUEUE_DELAY)
			}
		}
		return noConn, err
	}
	return false, nil
}
