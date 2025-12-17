package models

type PropostaAPIResponse struct {
	Msg                   string               `json:"msg"`
	HasError              bool                 `json:"hasError"`
	Messages              []Message            `json:"messages"`
	Codigo                string               `json:"codigo"`
	DtInclusao            string               `json:"dtInclusao"`
	Situacao              int                  `json:"situacao"`
	VlrFinanciado         float64              `json:"vlrFinanciado"`
	QtdeParcelas          int                  `json:"qtdeParcelas"`
	VlrTAC                float64              `json:"vlrTAC"`
	VlrSeguro             float64              `json:"vlrSeguro"`
	VlrIOF                float64              `json:"vlrIOF"`
	VlrOutrasDespesas     float64              `json:"vlrOutrasDespesas"`
	VlrOutrosServicos     float64              `json:"vlrOutrosServicos"`
	VlrTotalCredito       float64              `json:"vlrTotalCredito"`
	VlrTotalDivida        float64              `json:"vlrTotalDivida"`
	PercCETMensal         float64              `json:"percCETMensal"`
	PercCETAnual          float64              `json:"percCETAnual"`
	PercJurosMensal       float64              `json:"percJurosMensal"`
	PercJurosAnual        float64              `json:"percJurosAnual"`
	CodigoProposta        string               `json:"codigoProposta"`
	NumeroCCB             string               `json:"numeroCCB"`
	MotivoRejeicao        string               `json:"motivoRejeicao"`
	MotivoPendencia       string               `json:"motivoPendencia"`
	MotivoCancelamento    string               `json:"motivoCancelamento"`
	TextoRetornoPagamento string               `json:"textoRetornoPagamento"`
	AutenticacaoBancaria  string               `json:"autenticacaoBancaria"`
	ControleBancario      string               `json:"controleBancario"`
	StatusProtocoloCEF    any                  `json:"statusProtocoloCEF"`
	PropostaLancamentos   []PropostaLancamento `json:"propostaLancamentos"`
	CancelamentoFGTS      []CancelamentoFGTS   `json:"cancelamentoFGTS"`
	Result                any                  `json:"result"`
}

type PropostaLancamento struct {
	CampoID               string  `json:"campoID,"`
	DescricaoCampo        string  `json:"descricaoCampo,"`
	VlrTransacao          float64 `json:"vlrTransacao,"`
	DtPrevPagto           string  `json:"dtPrevPagto,"`
	DtPagamento           string  `json:"dtPagamento,"`
	Situacao              int     `json:"situacao,"`
	LinhaDigitavel        string  `json:"linhaDigitavel,"`
	DtVenctoBoleto        string  `json:"dtVenctoBoleto,"`
	VlrBoleto             float64 `json:"vlrBoleto,"`
	CodigoBanco           int     `json:"codigoBanco,"`
	NumeroBanco           string  `json:"numeroBanco,"`
	TipoConta             int     `json:"tipoConta,"`
	Agencia               string  `json:"agencia,"`
	AgenciaDig            string  `json:"agenciaDig,"`
	Conta                 string  `json:"conta,"`
	ContaDig              string  `json:"contaDig,"`
	DocumentoFederal      string  `json:"documentoFederal,"`
	NomePagamento         string  `json:"nomePagamento,"`
	TextoRetornoPagamento string  `json:"textoRetornoPagamento,"`
	AutenticacaoBancaria  string  `json:"autenticacaoBancaria,"`
	ControleBancario      string  `json:"controleBancario,"`
	DescricaoOcorrencia   string  `json:"descricaoOcorrencia,"`
	NomeCedente           string  `json:"nomeCedente,"`
}

type CancelamentoFGTS struct {
	Detalhes           string  `json:"detalhes"`
	ContextoEvento     string  `json:"contextoEvento"`
	TipoEvento         int     `json:"tipoEvento"`
	NomeEvento         string  `json:"nomeEvento"`
	DtEvento           string  `json:"dtEvento"`
	NumCodigoBarras    string  `json:"numCodigoBarras"`
	NumLinhaDigitavel  string  `json:"numLinhaDigitavel"`
	PixCopiaCola       string  `json:"pixCopiaCola"`
	VlrBoleto          float64 `json:"vlrBoleto"`
	DtVencimento       string  `json:"dtVencimento"`
	NumeroBoleto       int     `json:"numeroBoleto"`
	UrlImpressaoBoleto string  `json:"urlImpressaoBoleto"`
}
