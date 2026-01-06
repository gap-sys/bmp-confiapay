package models

import "cobranca-bmp/helpers"

type LancamentoParcelaFrontendInput struct {
	IdPropostaParcela    int     `json:"id_proposta_parcela" validate:"required"`
	IdProposta           int     `json:"id_proposta" validate:"required"`
	NumeroAcompanhamento string  `json:"numero_acompanhamento" validate:"required"`
	Operacao             string  `json:"operacao" validate:"required,oneof=P S" `
	DataLancamento       string  `json:"data_lancamento" validate:"required,datetime=2006-01-02"`
	ValorLancamento      float64 `json:"valor_lancamento" validate:"required_if=Operacao P"`
	ValorDesconto        float64 `json:"valor_desconto" validate:"required_if=Operacao S"`
	NumeroCCB            int     `json:"numero_ccb" validate:"required"`
	UrlWebhook           string  `json:"url_webhook" validate:"required,url,ping"`
	NumeroParcela        int     `json:"parcela" validate:"required"`
	IdConvenio           int     `json:"id_convenio"`
	IdSecuritizadora     int     `json:"id_securitizadora"`
}

func (l LancamentoParcelaFrontendInput) Validate() error {
	var APIError APIError
	var msgs string
	var messageErrors = make([]APIMessage, 0)
	err := helpers.StructValidate(l, "ping")
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
