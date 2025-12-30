package models

import "cobranca-bmp/helpers"

type GerarCobrancaFrontendInput struct {
	NumeroAcompanhamento string  `json:"numeroAcompanhamento" validate:"required"`
	NumeroCCB            int     `json:"numeroCCB" validate:"required"`
	IdProposta           int     `json:"idProposta" validate:"required"`
	UrlWebhook           string  `json:"urlWebhook" validate:"required"`
	Token                string  `json:"token" validate:"required"`
	DataVencimento       string  `json:"dataVencimento" validate:"required"`
	DataExpiracao        string  `json:"dataExpiracao" validate:"required"`
	NumeroParcela        int     `json:"numeroParcela" validate:"required"`
	IdParcela            string  `json:"idParcela" validate:"required"`
	ValorParcela         float64 `json:"valorParcela" `
	TipoCobranca         int     `json:"tipoCobranca" validate:"required"`
}

func (g GerarCobrancaFrontendInput) Validate() error {
	return helpers.StructValidate(g)
}
