package models

type WebhookResponse struct {
	IdParcela        int             `json:"idParcela"`
	CodigoLiquidacao string          `json:"codigoLiquidacao"`
	Boleto           *ConsultaBoleto `json:"boleto,omitempty"`
	Pix              *ConsultaPix    `json:"pix,omitempty"`
	TipoEvento       string          `json:"tipoEvento,omitempty"`
}
