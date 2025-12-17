package models

import "cobranca-bmp/helpers"

type CancelarCobrancaFrontendInput struct {
	IdProposta           int    `json:"idProposta" validate:"required"`
	NumeroAcompanhamento string `json:"numeroAcompanhamento"`
	CodigoLiquidacao     string `json:"codigoLiquidacao"`
	UrlWebhook           string `json:"urlWebhook"`
	Parcelas             []int  `json:"parcelas"`
}

func (c CancelarCobrancaFrontendInput) Validate() error {
	var APIError APIError
	var msgs string
	var messageErrors = make([]APIMessage, 0)
	err := helpers.StructValidate(c)
	if err != nil {
		errs, okErr := err.(*helpers.ErrValidation)
		if okErr {
			for _, message := range errs.GetAllMessages() {
				messageErrors = append(messageErrors, APIMessage{
					Code:        "",
					Description: message,
				})
				msgs += message + "\n"
			}

			APIError.HasError = true
			APIError.Messages = messageErrors
			APIError.Msg = msgs

		}
		return APIError
	}
	return nil
}
