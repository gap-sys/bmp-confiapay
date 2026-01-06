package models

type LancamentoParcelaAPIInput struct {
	DTO                  DTOPropostaParcela   `json:"dto"`
	DTOLancamentoParcela DTOLancamentoParcela `json:"dtoLancamentoParcela"`
}

type DTOLancamentoParcela struct {
	DtPagamento             string       `json:"dtPagamento"`
	VlrPagamento            float64      `json:"vlrPagamento"`
	VlrDesconto             float64      `json:"vlrDesconto"`
	PermiteDescapitalizacao bool         `json:"permiteDescapitalizacao"`
	DescricaoLancamento     string       `json:"descricaoLancamento"`
	Parametros              []Parametros `json:"parametros,omitempty"`
}

type Parametros struct {
	Nome  string `json:"nome"`
	Valor string `json:"valor"`
}

type DTOPropostaParcela struct {
	CodigoOperacao string `json:"codigoOperacao"`
	CodigoProposta string `json:"codigoProposta"`
	NroParcela     int    `json:"nroParcela"`
}
