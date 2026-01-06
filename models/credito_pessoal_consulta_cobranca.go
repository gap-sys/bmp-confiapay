package models

import "cobranca-bmp/helpers"

type ConsultaCobrancaResponse struct {
	Msg                     string                    `json:"msg"`
	HasError                bool                      `json:"hasError"`
	Messages                []ConsultaCobrancaMessage `json:"messages"`
	NumeroProposta          int                       `json:"numeroProposta"`
	DtInclusao              string                    `json:"dtInclusao"`
	DtCancelamento          string                    `json:"dtCancelamento"`
	DtLiquidacao            string                    `json:"dtLiquidacao"`
	SituacaoAgenda          int                       `json:"situacaoAgenda"`
	DescricaoSituacaoAgenda string                    `json:"descricaoSituacaoAgenda"`
	DocumentoFederalCliente string                    `json:"documentoFederalCliente"`
	NomeCliente             string                    `json:"nomeCliente"`
	NomeEmpresa             string                    `json:"nomeEmpresa"`
	EcCliente               string                    `json:"ecCliente"`
	VlrFinanciado           float64                   `json:"vlrFinanciado"`
	VlrRetencaoDiarioAprox  float64                   `json:"vlrRetencaoDiarioAprox"`
	VlrTotalPago            float64                   `json:"vlrTotalPago"`
	QTdeParcelas            int                       `json:"qtdeParcelas"`
	SituacaoProposta        int                       `json:"situacaoProposta"`
	DescricaoTipoContrato   string                    `json:"descricaoTipoContrato"`
	Parcelas                []ConsultaCobrancaParcela `json:"parcelas"`
	Result                  any                       `json:"result"`
}

type ConsultaCobrancaMessage struct {
	MessageType int    `json:"messageType"`
	Code        string `json:"code"`
	Context     string `json:"context"`
	Description string `json:"description"`
	Field       string `json:"field"`
}

type ConsultaCobrancaParcela struct {
	CodigoParcela     string                       `json:"codigoParcela"`
	NroParcela        int                          `json:"nroParcela"`
	Situacao          int                          `json:"situacao"`
	DtVencimento      string                       `json:"dtVencimento"`
	VlrParcela        float64                      `json:"vlrParcela"`
	NroOrdemPagamento string                       `json:"nroOrdemPagamento"`
	VlrSaldoAtual     float64                      `json:"vlrSaldoAtual"`
	VlrTotalPago      float64                      `json:"vlrTotalPago"`
	VlrDesconto       float64                      `json:"vlrDesconto"`
	DtVencimentoAtual string                       `json:"dtVencimentoAtual"`
	Boletos           []ConsultaBoleto             `json:"boletos"`
	Pix               []ConsultaPix                `json:"pix"`
	Lancamentos       []ConsultaCobrancaLancamento `json:"lancamentos"`
}

type ConsultaBoleto struct {
	NumeroBoleto            int                  `json:"numeroBoleto"`
	Liquidacao              bool                 `json:"liquidacao"`
	DtCredito               string               `json:"dtCredito"`
	DtExpiracao             string               `json:"dtExpiracao"`
	LinhaDigitavel          string               `json:"linhaDigitavel"`
	OperacaoBoletoRegistro  int                  `json:"operacaoBoletoRegistro"`
	SituacaoBoletoRegistro  int                  `json:"situacaoBoletoRegistro"`
	DescricaoBoletoRegistro string               `json:"descricaoBoletoRegistro"`
	UrlImpressao            string               `json:"urlImpressao"`
	QrCode                  ConsultaBoletoQrCode `json:"qrCode"`
}

type ConsultaBoletoQrCode struct {
	Emv    string `json:"emv"`
	Imagem string `json:"imagem"`
}

type ConsultaPix struct {
	Liquidacao   bool   `json:"liquidacao"`
	DtCredito    string `json:"dtCredito"`
	DtVencimento string `json:"dtVencimento"`
	DtExpiracao  string `json:"dtExpiracao"`
	Emv          string `json:"emv"`
	Imagem       string `json:"imagem"`
}

type ConsultaCobrancaLancamento struct {
	DtVencimento         string  `json:"dtVencimento"`
	DtLancamento         string  `json:"dtLancamento"`
	VlrParcela           float64 `json:"vlrParcela"`
	VlrMulta             float64 `json:"vlrMulta"`
	VlrJuros             float64 `json:"vlrJuros"`
	VlrMora              float64 `json:"vlrMora"`
	VlrIOF               float64 `json:"vlrIOF"`
	VlrTarifas           float64 `json:"vlrTarifas"`
	VlrAbatimento        float64 `json:"vlrAbatimento"`
	VlrParcelaAtualizado float64 `json:"vlrParcelaAtualizado"`
	VlrPagamento         float64 `json:"vlrPagamento"`
	VlrSaldoParcela      float64 `json:"vlrSaldoParcela"`
	VlrDesconto          float64 `json:"vlrDesconto"`
	InformarFundo        bool    `json:"informarFundo"`
	DescLancamento       string  `json:"descLancamento"`
	ViaBoleto            bool    `json:"viaBoleto"`
}

type ConsultaCobrancaFrontEndInput struct {
	IdProposta           int    `json:"id_proposta" validate:"required"`
	NumeroAcompanhamento string `json:"numero_acompanhamento" validate:"required"`
	IdConvenio           int    `json:"id_convenio" validate:"required"`
	IdSecuritizadora     int    `json:"id_securitizadora" validate:"required"`
	Parcela              int    `json:"numero_parcela"`
}

func (c *ConsultaCobrancaFrontEndInput) Validate() error {
	var APIError APIError
	var msgs string
	var messageErrors = make([]APIMessage, 0)
	err := helpers.StructValidate(c)
	if err != nil {
		errs, okErr := err.(*helpers.ErrValidation)
		if okErr {
			for _, message := range errs.GetAllMessages() {
				messageErrors = append(messageErrors, APIMessage{
					Code:        "",
					Description: message,
				})
				msgs += message + "\n"
			}

			APIError.HasError = true
			APIError.Messages = messageErrors
			APIError.Msg = msgs

		}
		return APIError
	}
	return nil

}
