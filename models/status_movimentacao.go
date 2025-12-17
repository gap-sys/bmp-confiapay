package models

type StatusMovimentacao struct {
	IDEsteira         int    `json:"idEsteira"`
	IDStatus          int    `json:"idStatus"`
	IDVinculo         int    `json:"idVinculo"`
	Status            int    `json:"status"`
	MensagemHistorico string `json:"mensagemHistorico"`
	Exclusivo         string `json:"exclusivo"`
	Operacao          string `json:"operacao"`
	Sequencia         int    `json:"sequencia"`
}
