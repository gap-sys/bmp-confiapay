package models

import (
	"cobranca-bmp/helpers"
)

type BuscarPropostaDb struct {
	IdProposta int `json:"IdProposta" validate:"required"`
}

type DigitacaoFGTSFrontendInput struct {
	BuscarPropostaDb
	Assync          bool   `json:"Assync,omitempty"`
	IdCliente       int    `json:"idCliente" validate:"required_with=Webhook"`
	Webhook         string `json:"webhook" validate:"omitempty,url,ping"`
	UrlCancelamento string `json:"urlCancelamento"`
}

func (b BuscarPropostaDb) Validate() error {
	if b.IdProposta <= 0 {
		return NewAPIError("", "Insira um IdProposta", "")

	}
	return nil
}

func (d DigitacaoFGTSFrontendInput) Validate() error {
	var APIError APIError
	var messages = make([]APIMessage, 0)
	err := helpers.StructValidate(d, "ping")
	if err != nil {
		errValidation, ok := err.(*helpers.ErrValidation)
		if ok {
			for _, msg := range errValidation.GetAllMessages() {
				messages = append(messages, APIMessage{
					Description: msg,
				})
				APIError.Msg += msg + " \n"
			}

			APIError.Messages = messages
			APIError.HasError = true
			err = APIError
		}
		return err
	}

	return nil
}

type DigitacaoINSSCPFrontendInput struct {
	BuscarPropostaDb
	Assync          bool   `json:"Assync,omitempty"`
	Webhook         string `json:"webhook" validate:"omitempty,url,ping"`
	UrlCancelamento string `json:"urlCancelamento"`
}

func (d DigitacaoINSSCPFrontendInput) Validate() error {
	var APIError APIError
	var messages = make([]APIMessage, 0)
	err := helpers.StructValidate(d, "ping")
	if err != nil {
		errValidation, ok := err.(*helpers.ErrValidation)
		if ok {
			for _, msg := range errValidation.GetAllMessages() {
				messages = append(messages, APIMessage{
					Description: msg,
				})
				APIError.Msg += msg + " \n"
			}

			APIError.Messages = messages
			APIError.HasError = true
			err = APIError
		}
		return err
	}

	return nil
}

type DigitacaoCreditoTrabalhadorFrontendInput struct {
	BuscarPropostaDb
	Assync          bool   `json:"Assync,omitempty"`
	Webhook         string `json:"webhook" validate:"omitempty,url,ping"`
	UrlCancelamento string `json:"urlCancelamento"`
}

func (d DigitacaoCreditoTrabalhadorFrontendInput) Validate() error {
	var APIError APIError
	var messages = make([]APIMessage, 0)
	err := helpers.StructValidate(d, "ping")
	if err != nil {
		errValidation, ok := err.(*helpers.ErrValidation)
		if ok {
			for _, msg := range errValidation.GetAllMessages() {
				messages = append(messages, APIMessage{
					Description: msg,
				})
				APIError.Msg += msg + " \n"
			}

			APIError.Messages = messages
			APIError.HasError = true
			err = APIError
		}
		return err
	}

	return nil
}
