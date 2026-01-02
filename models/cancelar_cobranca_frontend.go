package models

import "cobranca-bmp/helpers"

type CancelarCobrancaFrontendInput struct {
	IdPropostaParcela    int    `json:"id_proposta_parcela" validate:"required"`
	IdProposta           int    `json:"id_proposta" validate:"required"`
	NumeroAcompanhamento string `json:"numero_acompanhamento" validate:"required"`
	CodigoLiquidacao     string `json:"codigo_liquidacao" validate:"required" `
	NumeroCCB            int    `json:"numero_ccb" validate:"required"`
	UrlWebhook           string `json:"url_webhook" validate:"required,url,ping"`
	NumeroParcela        int    `json:"numero_parcela" validate:"required"`
	IdConvenio           int    `json:"id_convenio"`
	IdSecuritizadora     int    `json:"id_securitizadora"`
}

func (c CancelarCobrancaFrontendInput) Validate() error {
	var APIError APIError
	var msgs string
	var messageErrors = make([]APIMessage, 0)
	err := helpers.StructValidate(c, "ping")
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
