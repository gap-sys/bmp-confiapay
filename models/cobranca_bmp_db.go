package models

type CobrancaBMP struct {
	NumeroAcompanhamento string  `json:"numero_acompanhamento" validate:"required"`
	NumeroCCB            int     `json:"numero_ccb" validate:"required"`
	IdProposta           int     `json:"id_proposta" validate:"required"`
	UrlWebhook           string  `json:"url_webhook" validate:"required"`
	DataVencimento       string  `json:"data_vencimento" validate:"required"`
	DataExpiracao        string  `json:"data_expiracao" validate:"required"`
	NumeroParcela        int     `json:"numero_parcela" validate:"required"`
	CodigoLiquidacao     string  `json:"codigoLiquidacao"`
	IdPropostaParcela    int     `json:"id_proposta_parcela" validate:"required"`
	ValorParcela         float64 `json:"valor_parcela"`
	TipoCobranca         int     `json:"id_forma_cobranca" validate:"required"`
	IdConvenio           int     `json:"id_convenio"`
	IdSecuritizadora     int     `json:"id_securitizadora"`
	IdFormaCobranca      int     `json:"idFormaCobranca"`
}
