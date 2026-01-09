package models

import (
	"cobranca-bmp/helpers"
)

type GerarCobrancaFrontendInput struct {
	NumeroAcompanhamento string  `json:"numero_acompanhamento" validate:"required"`
	NumeroCCB            int     `json:"numero_ccb" validate:"required"`
	IdProposta           int     `json:"id_proposta" validate:"required"`
	UrlWebhook           string  `json:"url_webhook" validate:"required,url,ping"`
	DataVencimento       string  `json:"data_vencimento" validate:"required,datetime=2006-01-02"`
	DataExpiracao        string  `json:"data_expiracao" validate:"required,datetime=2006-01-02"`
	NumeroParcela        int     `json:"parcela" validate:"required"`
	IdPropostaParcela    int     `json:"id_proposta_parcela" validate:"required"`
	ValorParcela         float64 `json:"valor_parcela"`
	TipoCobranca         int     `json:"id_forma_cobranca"`
	IdConvenio           int     `json:"id_convenio"`
	IdSecuritizadora     int     `json:"id_securitizadora"`
	VlrDesconto          float64 `json:"valor_desconto"`
}

func (g GerarCobrancaFrontendInput) Validate() error {
	var errAPI APIError
	errAPI.ID = g.IdPropostaParcela
	errAPI.HasError = true

	messages := make([]APIMessage, 0)
	if err := helpers.StructValidate(g, "ping"); err != nil {
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
