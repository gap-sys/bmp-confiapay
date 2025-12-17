package models

import (
	"cobranca-bmp/helpers"
	"fmt"
)

type PropostaParcelasDb struct {
	Parcela          int     `json:"parcela" validate:"required"`
	DataVencimento   string  `json:"dataVencimento" validate:"required,datetime=2006-01-02"`
	DataExpiracao    string  `json:"dataExpiracao" validate:"required,datetime=2006-01-02"`
	ValorParcela     float64 `json:"valorParcela" validate:"required"`
	IdFormaCobranca  int64   `json:"idFormaCobranca"`
	CodigoLiquidacao string  `json:"codigoLiquidacao"`
	Status           int     `json:"status"`
}

func (p PropostaParcelasDb) Validate(idProposta int) error {
	var errAPI APIError
	errAPI.ID = idProposta
	errAPI.HasError = true

	messages := make([]APIMessage, 0)
	if err := helpers.StructValidate(p, "validate_cpf"); err != nil {
		errValidation, ok := err.(*helpers.ErrValidation)
		if !ok {
			messages := append(messages, APIMessage{Description: err.Error()})
			errAPI.Messages = messages
			errAPI.Msg = messages[0].Description
			return errAPI
		}
		for _, message := range errValidation.GetAllMessages() {
			var msg = APIMessage{
				Description: message + fmt.Sprintf("(parcela %d)", p.Parcela),
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
