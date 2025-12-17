package models

type ConsultarCobrancaResponse struct {
	Msg      string            `json:"msg"`
	HasError bool              `json:"hasError"`
	Messages []CobrancaMessage `json:"messages"`
	Boletos  []BoletoResponse  `json:"boletos"`
}

type GerarCobrancaResponse struct {
	Msg       string            `json:"msg"`
	HasError  bool              `json:"hasError"`
	Messages  []CobrancaMessage `json:"messages"`
	Cobrancas []Cobranca        `json:"cobrancas"`
	Result    any               `json:"result"`
}

type CobrancaMessage struct {
	MessageType int    `json:"messageType"`
	Code        string `json:"code"`
	Context     string `json:"context"`
	Description string `json:"description"`
	Field       string `json:"field"`
}

type BoletoResponse struct {
	Liquidacao              bool   `json:"liquidacao"`
	Parcelas                string `json:"parcelas"`
	Descricao               string `json:"descricao"`
	DtInclusao              string `json:"dtInclusao"`
	DtVencimento            string `json:"dtVencimento"`
	DtCancelamento          string `json:"dtCancelamento"`
	DtExpiracao             string `json:"dtExpiracao"`
	DtCredito               string `json:"dtCredito"`
	ValorBoleto             int    `json:"valorBoleto"`
	VlrPago                 int    `json:"vlrPago"`
	Impressao               string `json:"impressao"`
	LinhaDigitavel          string `json:"linhaDigitavel"`
	DocumentoFederalCliente string `json:"documentoFederalCliente"`
	Cliente                 string `json:"cliente"`
	Parceiro                string `json:"parceiro"`
	QrCode                  QrCode `json:"qrCode"`
}

type QrCode struct {
	Emv    string `json:"emv"`
	Imagem string `json:"imagem"`
}
