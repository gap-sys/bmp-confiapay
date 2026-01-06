package service

import "cobranca-bmp/models"

// Representa os repositórios que realizarão ações no banco de dados.

type ParcelaRepository interface {
	UpdateCodLiquidacao(IdPropostaParcela int, codigoLiquidacao string) (bool, error)
	UpdateGeracaoCobranca(data models.GerarCobrancaFrontendInput) (bool, error)
	UpdateCancelamentoCobranca(data models.CancelarCobrancaFrontendInput, codigoLiquidacao string) (bool, error)
	UpdateLancamentoParcela(data models.LancamentoParcelaFrontendInput) (bool, error)
	FindByCodLiquidacao(codigoLiquidacao string, numeroCCB int) (models.CobrancaBMP, error)
	FindByNumParcela(numParcela int, numeroCCB int) (models.CobrancaBMP, error)
	FindByDataVencimento(dataExpiracao string, numeroCCB int) (models.CobrancaBMP, error)
}
