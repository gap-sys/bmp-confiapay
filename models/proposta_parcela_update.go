package models

type PropostaParcelaUpdate struct {
	IdProposta       int    `json:"idProposta"`
	DataVencimento   string `json:"dataExpiracao"`
	Parcela          int    `json:"parcela"`
	Parcelas         []int  `json:"parcelas"`
	Boleto           any    `json:"boleto"`
	Pix              any    `json:"pix"`
	Response         any    `json:"response"`
	IdTipoCobranca   int    `json:"idTipoCobranca"`
	CodigoLiquidacao string `json:"codigoLiquidacao"`
	Status           int    `json:"status"`
	SearchMode       string `json:"searchMode"`
}
