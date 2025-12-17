package models

type EventoCobranca struct {
	TipoEvento  int    `json:"TipoEvento"`
	NomeEvento  string `json:"NomeEvento"`
	NroProposta int    `json:"NroProposta"`
}

type EventoCancelamentoAgenda struct {
	EventoCobranca
	DtEvento              string             `json:"DtEvento"`
	AcrescimoAbatimento   *any               `json:"AcrescimoAbatimento"`
	LancamentoParcela     *any               `json:"LancamentoParcela"`
	ProrrogacaoVencimento *any               `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *any               `json:"GeracaoBoleto"`
	CancelamentoBoleto    CancelamentoBoleto `json:"CancelamentoBoleto"`
}

type AcrescimoAbatimento struct {
	NroParcela    int `json:"NroParcela"`
	VlrSaldo      int `json:"VlrSaldo"`
	VlrAcrescimo  int `json:"VlrAcrescimo"`
	VlrAbatimento int `json:"VlrAbatimento"`
	VlrSaldoAtual int `json:"VlrSaldoAtual"`
}

type EventoAcrescimoAbatimento struct {
	EventoCobranca
	DtEvento              string              `json:"DtEvento"`
	AcrescimoAbatimento   AcrescimoAbatimento `json:"AcrescimoAbatimento"`
	LancamentoParcela     *any                `json:"LancamentoParcela"`
	ProrrogacaoVencimento *any                `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *any                `json:"GeracaoBoleto"`
	CancelamentoBoleto    *any                `json:"CancelamentoBoleto"`
}

type BoletoLancamento struct {
	NroBoleto  *int `json:"NroBoleto"`
	Liquidacao *any `json:"Liquidacao"`
}

type LancamentoParcela struct {
	NroParcela         int              `json:"NroParcela"`
	VlrOriginalParcela float64          `json:"VlrOriginalParcela"`
	VlrSaldo           float64          `json:"VlrSaldo"`
	ParcelaLiquidada   bool             `json:"ParcelaLiquidada"`
	VlrEncargos        float64          `json:"VlrEncargos"`
	VlrAbatimento      float64          `json:"VlrAbatimento"`
	VlrPagamento       float64          `json:"VlrPagamento"`
	VlrDesconto        float64          `json:"VlrDesconto"`
	VlrSaldoAtual      float64          `json:"VlrSaldoAtual"`
	VlrExcedente       float64          `json:"VlrExcedente"`
	DtVencimento       string           `json:"DtVencimento"`
	DtVencimentoAtual  string           `json:"DtVencimentoAtual"`
	Boleto             BoletoLancamento `json:"Boleto"`
}

type EventoLancamentoParcela struct {
	EventoCobranca
	DtEvento              string            `json:"DtEvento"`
	AcrescimoAbatimento   *any              `json:"AcrescimoAbatimento"`
	LancamentoParcela     LancamentoParcela `json:"LancamentoParcela"`
	ProrrogacaoVencimento *any              `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *any              `json:"GeracaoBoleto"`
	CancelamentoBoleto    *any              `json:"CancelamentoBoleto"`
}

type ProrrogacaoVencimento struct {
	NroParcela           int    `json:"NroParcela"`
	DtVencimentoAnterior string `json:"DtVencimentoAnterior"`
	DtVencimentoAtual    string `json:"DtVencimentoAtual"`
}

type EventoProrrogacaoVencimento struct {
	EventoCobranca
	DtEvento              string                 `json:"DtEvento"`
	AcrescimoAbatimento   *any                   `json:"AcrescimoAbatimento"`
	LancamentoParcela     *any                   `json:"LancamentoParcela"`
	ProrrogacaoVencimento *ProrrogacaoVencimento `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *any                   `json:"GeracaoBoleto"`
	CancelamentoBoleto    *any                   `json:"CancelamentoBoleto"`
}

type GeracaoBoleto struct {
	CodigoBoleto string  `json:"CodigoBoleto"`
	NroBoleto    int     `json:"NroBoleto"`
	VlrBoleto    float64 `json:"VlrBoleto"`
	DtVencimento string  `json:"DtVencimento"`
	Liquidacao   bool    `json:"Liquidacao"`
}

type EventoGeracaoBoleto struct {
	EventoCobranca
	DtEvento              string        `json:"DtEvento"`
	AcrescimoAbatimento   *any          `json:"AcrescimoAbatimento"`
	LancamentoParcela     *any          `json:"LancamentoParcela"`
	ProrrogacaoVencimento *any          `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         GeracaoBoleto `json:"GeracaoBoleto"`
	CancelamentoBoleto    *any          `json:"CancelamentoBoleto"`
}

type CancelamentoBoleto struct {
	CodigoBoleto string `json:"CodigoBoleto"`
	NroBoleto    int    `json:"NroBoleto"`
	Parcelas     []int  `json:"Parcelas"`
}

type GeracaoCobranca struct {
	CodigoLiquidacao string  `json:"CodigoLiquidacao"`
	CodigoBoleto     *string `json:"CodigoBoleto"`
	NroBoleto        *int    `json:"NroBoleto"`
	Emv              *string `json:"Emv"`    // código do pix copia e cola
	Imagem           *string `json:"Imagem"` // imagem do qr code
	Parcelas         []int   `json:"Parcelas"`
}

type EventoRegistroCobranca struct {
	EventoCobranca
	DtEvento              string                 `json:"DtEvento"`
	AcrescimoAbatimento   *AcrescimoAbatimento   `json:"AcrescimoAbatimento"`
	LancamentoParcela     *LancamentoParcela     `json:"LancamentoParcela"`
	ProrrogacaoVencimento *ProrrogacaoVencimento `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *GeracaoBoleto         `json:"GeracaoBoleto"`
	CancelamentoBoleto    *CancelamentoBoleto    `json:"CancelamentoBoleto"`
	GeracaoCobranca       *GeracaoCobranca       `json:"GeracaoCobranca"`
	CancelamentoCobranca  any                    `json:"CancelamentoCobranca"`
}

type EventoLiquidacaoAgenda struct {
	EventoCobranca
	DtEvento              string                 `json:"DtEvento"`
	AcrescimoAbatimento   *AcrescimoAbatimento   `json:"AcrescimoAbatimento"`
	LancamentoParcela     *LancamentoParcela     `json:"LancamentoParcela"`
	ProrrogacaoVencimento *ProrrogacaoVencimento `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *GeracaoBoleto         `json:"GeracaoBoleto"`
	CancelamentoBoleto    *CancelamentoBoleto    `json:"CancelamentoBoleto"`
}

type EventoCancelamentoCobranca struct {
	EventoCobranca
	DtEvento              string                 `json:"DtEvento"`
	AcrescimoAbatimento   *AcrescimoAbatimento   `json:"AcrescimoAbatimento"`
	LancamentoParcela     *LancamentoParcela     `json:"LancamentoParcela"`
	ProrrogacaoVencimento *ProrrogacaoVencimento `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *GeracaoBoleto         `json:"GeracaoBoleto"`
	CancelamentoCobranca  CancelamentoCobranca   `json:"CancelamentoCobranca"`
}

type CancelamentoCobranca struct {
	CodigoLiquidacao  string `json:"CodigoLiquidacao"`
	CodigoBoleto      any    `json:"CodigoBoleto"`
	NroBoleto         any    `json:"NroBoleto"`
	Emv               string `json:"Emv"`
	Parcelas          []int  `json:"Parcelas"`
	EstornoLancamento any    `json:"EstornoLancamento"`
}

type EventoCancelamentoBoleto struct {
	EventoCobranca
	DtEvento              string                 `json:"DtEvento"`
	AcrescimoAbatimento   *AcrescimoAbatimento   `json:"AcrescimoAbatimento"`
	LancamentoParcela     *LancamentoParcela     `json:"LancamentoParcela"`
	ProrrogacaoVencimento *ProrrogacaoVencimento `json:"ProrrogacaoVencimento"`
	GeracaoBoleto         *GeracaoBoleto         `json:"GeracaoBoleto"`
	CancelamentoBoleto    *CancelamentoBoleto    `json:"CancelamentoBoleto"`
}

type EventoEstorno struct {
	EventoCobranca
	DtEvento              string            `json:"DtEvento"` // mantive como string (ex: "2025-11-25")
	AcrescimoAbatimento   float64           `json:"AcrescimoAbatimento,omitempty"`
	LancamentoParcela     any               `json:"LancamentoParcela,omitempty"`     // estrutura desconhecida -> RawMessage
	ProrrogacaoVencimento any               `json:"ProrrogacaoVencimento,omitempty"` // estrutura desconhecida -> RawMessage
	GeracaoBoleto         any               `json:"GeracaoBoleto,omitempty"`
	CancelamentoBoleto    any               `json:"CancelamentoBoleto"`
	GeracaoCobranca       any               `json:"GeracaoCobranca,omitempty"`
	CancelamentoCobranca  any               `json:"CancelamentoCobranca,omitempty"`
	EstornoLancamento     EstornoLancamento `json:"EstornoLancamento"`
}

type EstornoLancamento struct {
	NroParcela            int                   `json:"NroParcela"`
	LancamentosEstornados []LancamentoEstornado `json:"LancamentosEstornados"`
}

type LancamentoEstornado struct {
	Codigo     string  `json:"Codigo"`
	DtInclusao string  `json:"DtInclusao"`
	DtEstorno  string  `json:"DtEstorno"`
	Tipo       string  `json:"Tipo"`
	Valor      float64 `json:"Valor"`
	Motivo     string  `json:"Motivo"`
}
