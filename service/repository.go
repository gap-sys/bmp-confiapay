package service

import "cobranca-bmp/models"

// Representa os repositórios que realizarão ações no banco de dados.

type ParcelaRepository interface {
	UpdateCodLiquidacao(IdPropostaParcela int, codigoLiquidacao string) (bool, error)
	UpdateGeracaoCobranca(data models.GerarCobrancaFrontendInput, codigoLiquidacao string) (bool, error)
}
