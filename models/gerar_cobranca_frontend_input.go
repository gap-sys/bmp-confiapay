package models

import (
	"cobranca-bmp/helpers"
)

type GerarCobrancaFrontendInput struct {
	NumeroAcompanhamento string  `json:"numero_acompanhamento" validate:"required"`
	NumeroCCB            int     `json:"numero_ccb" validate:"required"`
	IdProposta           int     `json:"id_proposta" validate:"required"`
	UrlWebhook           string  `json:"url_webhook" validate:"required"`
	DataVencimento       string  `json:"data_vencimento" validate:"required"`
	DataExpiracao        string  `json:"data_expiracao" validate:"required"`
	NumeroParcela        int     `json:"numero_parcela" validate:"required"`
	IdPropostaParcela    int     `json:"id_proposta_parcela" validate:"required"`
	ValorParcela         float64 `json:"valor_parcela"`
	TipoCobranca         int     `json:"id_forma_cobranca" validate:"required"`
	IdConvenio           int     `json:"id_convenio"`
	IdSecuritizadora     int     `json:"id_securitizadora"`
}

func (g GerarCobrancaFrontendInput) Validate() error {
	var errAPI APIError
	errAPI.ID = g.IdPropostaParcela
	errAPI.HasError = true

	messages := make([]APIMessage, 0)
	if err := helpers.StructValidate(g); err != nil {
		errValidation, ok := err.(*helpers.ErrValidation)
		if !ok {
			messages := append(messages, APIMessage{Description: err.Error()})
			errAPI.Messages = messages
			errAPI.Msg = messages[0].Description
			return errAPI
		}
		for _, message := range errValidation.GetAllMessages() {
			var msg = APIMessage{
				Description: message,
				Context:     "CONFIAPAY",
			}
			messages = append(messages, msg)
			errAPI.Msg += message + "\n"
		}
		errAPI.Messages = messages
		return errAPI
	}
	return nil
}
