package models

type EventoGeracaoAgenda struct {
	EventoCobranca
	QtdeParcelas   int             `json:"QtdeParcelas"`
	VlrTotalDivida float64         `json:"VlrTotalDivida"`
	DtFinanciado   string          `json:"DtFinanciado"`
	Credito        Credito         `json:"Credito"`
	Conta          Conta           `json:"Conta"`
	Parcelas       []ParcelaAgenda `json:"Parcelas"`
}

type Credito struct {
	DtCredito *string `json:"DtCredito"`
}

type Conta struct {
	CodigoBanco *string `json:"CodigoBanco"`
	NomeBanco   *string `json:"NomeBanco"`
	Agencia     *string `json:"Agencia"`
	Conta       *string `json:"Conta"`
	ContaDigito *string `json:"ContaDigito"`
}

type ParcelaAgenda struct {
	NroParcela      int     `json:"NroParcela"`
	VlrParcela      float64 `json:"VlrParcela"`
	VlrPrincipal    float64 `json:"VlrPrincipal"`
	VlrDespesas     float64 `json:"VlrDespesas"`
	VlrSaldoDevedor float64 `json:"VlrSaldoDevedor"`
	DtVencimento    string  `json:"DtVencimento"`
}
