package models

import (
	"encoding/json"
	"os"
)

type UpdateDBData struct {
	IdProposta        int                    `json:"idProposta,omitempty"`
	IdSimulacao       int                    `json:"idSimulacao"`
	CodigoProposta    string                 `json:"codigoProposta,omitempty"`
	IdTipoData        int                    `json:"idTipoData"`
	IdFormaCobranca   int                    `xml:"idFormaCobranca"`
	TipoDataToExclude []int                  `json:"tipoDataToExclude"`
	StatusId          int                    `json:"statusBmp,omitempty"`
	ErrorCode         string                 `json:"errorCode,omitempty"`
	Context           string                 `json:"context,omitempty"`
	FromWebhook       bool                   `json:"fromWebhook,omitempty"`
	UrlCCB            string                 `json:"urlCCB,omitempty"`
	NumeroCCB         int                    `json:"numeroCCB"`
	UpdateFluxo       bool                   `json:"updateFluxo,omitempty"`
	Reprocessamento   bool                   `json:"reprocessamento,omitempty"`
	Json              any                    `json:"json,omitempty"`
	StatusErroPadrao  string                 `json:"statusErrPadrao,omitempty"`
	Action            string                 `json:"action,omitempty"`
	Msg               string                 `json:"msg,omitempty"`
	Convenio          int                    `json:"convenio"`
	NumeroProposta    string                 `json:"numeroProposta"`
	Cobrancas         *GerarCobrancaResponse `json:"parcelasResponse"`
	ParcelaUpdate     *PropostaParcelaUpdate `json:"parcelaUpdate"`
}

func (d UpdateDBData) MarshalBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *UpdateDBData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

type DbLogfileData struct {
	FilePath string
	File     *os.File
}
