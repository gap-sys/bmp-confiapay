package models

import "os"

type UpdateDbData struct {
	IdPropostaParcela    int                            `json:"idPropostaParcela"`
	CodigoLiquidacao     string                         `json:"codigoLiquidacao"`
	Action               string                         `json:"action"`
	GeracaoParcela       *GerarCobrancaFrontendInput    `json:"geracaoParcela"`
	CancelamentoCobranca *CancelarCobrancaFrontendInput `json:"cancelamentoCobranca"`
}

type DbLogfileData struct {
	FilePath string   `json:"filePath"`
	File     *os.File `json:"file"`
}
