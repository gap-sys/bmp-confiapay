package models

import "cobranca-bmp/helpers"

type MultiplasCobrancasInput struct {
	Dto                          DtoCobranca                  `json:"dto"`
	DtoGerarMultiplasLiquidacoes DtoGerarMultiplasLiquidacoes `json:"dtoGerarMultiplasLiquidacoes"`
}

type CobrancaUnicaInput struct {
	Dto                     DtoCobranca             `json:"dto"`
	DtoGerarUnicaLiquidacao DtoGerarUnicaLiquidacao `json:"dtoGerarUnicaLiquidacao"`
	NroParcelas             []int                   `json:"nroParcelas,omitempty"`
}

type CobrancaUnicaFrontendInput struct {
	CobrancaUnicaInput
	NumeroCCB  int    `json:"numeroCCB" validate:"required"`
	UrlWebhook string `json:"urlWebhook" validate:"required,url,ping"`
	IdCOnvenio int    `json:"idConvenio" validate:"required,oneof=2 3 5"`
}

func (c CobrancaUnicaFrontendInput) Validate() error {
	var APIError APIError
	var messages = make([]APIMessage, 0)
	err := helpers.StructValidate(c, "ping")
	if err != nil {
		errValidation, ok := err.(*helpers.ErrValidation)
		if ok {
			for _, msg := range errValidation.GetAllMessages() {
				messages = append(messages, APIMessage{
					Description: msg,
				})
				APIError.Msg += msg + " \n"
			}

			APIError.Messages = messages
			APIError.HasError = true
			err = APIError
		}
		return err
	}

	return nil

}

type DtoCobranca struct {
	CodigoProposta string `json:"codigoProposta"`
	CodigoOperacao string `json:"codigoOperacao"`
	NroParcelas    []int  `json:"nroParcelas,omitempty"`
}

type DtoGerarMultiplasLiquidacoes struct {
	DescricaoLiquidacao string            `json:"descricaoLiquidacao"`
	NotificarCliente    bool              `json:"notificarCliente"`
	TipoRegistro        int               `json:"tipoRegistro"` //1-online 2-offline
	PagamentoViaBoleto  bool              `json:"pagamentoViaBoleto"`
	PagamentoViaPIX     bool              `json:"pagamentoViaPIX,omitempty"`
	Parcelas            []ParcelaCobranca `json:"parcelas"`
}

type DtoGerarUnicaLiquidacao struct {
	DtVencimento            string  `json:"dtVencimento"`
	DtExpiracao             string  `json:"dtExpiracao"`
	Liquidacao              bool    `json:"liquidacao,omitempty"`
	PagamentoViaBoleto      bool    `json:"pagamentoViaBoleto"`
	PagamentoViaPIX         bool    `json:"pagamentoViaPIX"`
	DescricaoLiquidacao     string  `json:"descricaoLiquidacao"`
	VlrLiquidacao           float64 `json:"vlrLiquidacao,omitempty"`
	VlrDesconto             float64 `json:"vlrDesconto"`
	PermiteDescapitalizacao bool    `json:"permiteDescapitalizacao"`
	TipoRegistro            int     `json:"tipoRegistro"`
}

type ParcelaCobranca struct {
	NroParcela              int    `json:"nroParcela"`
	DtVencimento            string `json:"dtVencimento"`
	DtExpiracao             string `json:"dtExpiracao"`
	VlrDesconto             int    `json:"vlrDesconto"`
	PermiteDescapitalizacao bool   `json:"permiteDescapitalizacao"`
}

type Cobranca struct {
	CodigoLiquidacao string `json:"codigoLiquidacao"`
	Parcelas         []int  `json:"parcelas"`
}

type DTOCancelarCobrancas struct {
	CodigosLiquidacoes []string `json:"codigosLiquidacoes"`
}

type CancelarCobrancaInput struct {
	DTO                  DtoCobranca          `json:"dto"`
	DTOCancelarCobrancas DTOCancelarCobrancas `json:"dtoCancelarCobrancas"`
}

type DTOConsultaDetalhes struct {
	TrazerBoleto          bool `json:"trazerBoleto"`
	TrazerAgendaDetalhada bool `json:"trazerAgendaDetalhada"`
}

type ConsultarDetalhesInput struct {
	DTO                 DtoCobranca         `json:"dto"`
	DTOConsultaDetalhes DTOConsultaDetalhes `json:"dtoConsultaDetalhes"`
}
