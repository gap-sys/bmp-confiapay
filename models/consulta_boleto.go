package models

type Boleto struct {
	NumeroBoleto            int      `json:"numeroBoleto"`
	NumeroProposta          int      `json:"numeroProposta"`
	Liquidacao              bool     `json:"liquidacao"`
	Parcelas                string   `json:"parcelas"`
	Descricao               string   `json:"descricao"`
	DtInclusao              string   `json:"dtInclusao"`
	DtVencimento            string   `json:"dtVencimento"`
	DtCancelamento          *string  `json:"dtCancelamento"`
	DtExpiracao             string   `json:"dtExpiracao"`
	DtCredito               string   `json:"dtCredito"`
	ValorBoleto             float64  `json:"valorBoleto"`
	VlrPago                 *float64 `json:"vlrPago"`
	Impressao               string   `json:"impressao"`
	LinhaDigitavel          string   `json:"linhaDigitavel"`
	DocumentoFederalCliente string   `json:"documentoFederalCliente"`
	Cliente                 string   `json:"cliente"`
	Parceiro                string   `json:"parceiro"`
}

type BoletoConsultado struct {
	Boletos  []Boleto `json:"boletos"`
	Msg      string   `json:"msg"`
	HasError bool     `json:"hasError"`
	Messages any      `json:"messages"`
}

func (b Boleto) FormatToWebhook() []BoletoWebhook {
	return []BoletoWebhook{{
		Boleto:       b,
		UrlImpressao: b.Impressao,
	}}

}

type BoletoWebhook struct {
	Boleto
	UrlImpressao string `json:"urlImpressao"`
}

type ConsultaBoletoInput struct {
	DTO                ConsultaBoletoDTO        `json:"dto"`
	DTOConsultaBoletos ConsultaBoletosFiltroDTO `json:"dtoConsultaBoletos"`
}

type ConsultaBoletoDTO struct {
	CodigoProposta string `json:"codigoProposta"`
	CodigoOperacao string `json:"codigoOperacao"`
}

type ConsultaBoletosFiltroDTO struct {
	NumerosBoletos          []int `json:"numerosBoletos"`
	TrazerBoletosCancelados bool  `json:"trazerBoletosCancelados"`
}
