package models

type Documento struct {
	Arquivo     string `json:"arquivo" validate:"required"`
	Codigo      string `json:"codigo,omitempty"`
	DtValidade  string `json:"dtValidade"` //verificar consistência na API
	Extensao    string `json:"extensao" validate:"required,max=20"`
	NomeArquivo string `json:"nomeArquivo" validate:"required,max=255"`
	/*
		1	Fotoidentificação Pessoa
		2	Fotoidentificação Veículo
		3	Fotoidentificação Maquinário
		4	RG
		5	CPF
		6	Comprovante Endereço
		7	Comprovante Renda
		8	CNH
		9	Carteira Reservista
		10	Título Eleitor
		11	Passaporte
		12	Carteira Trabalho
		13	Contrato Social
		14	CNPJ
		15	Extrato Bancário
		16	Certidão Casamento
		17	Certidão Divórcio
		18	Certidão Óbito
		19	Ata de Eleição
		20	Procurações Públicas
		21	Balanço Patrimonial com DRE
		22	CCB
		23	Protocolo
		24	Balancete com DRE
		25	Faturamento
		26	Declaração IR
		27	Nota Fiscal Garantia
		28	Quitação de Garantia
		29	Laudo Técnico
		30	Certidão de Nascimento
		31	Averbação de Divórcio
		40	Documento Veículo (DUT)
		41	Nota Fiscal Garantia Adicional
		42	Orçamento
		50	Nota Fiscal
		51	Apólice de Seguro
		52	Nota Promissória
		53	CRV/CRLV
		54	Foto Veículo/Máquina
		55	Ficha Cadastral
		56	Consulta Bureau
		57	Frota
		58	Foto do Estabelecimento
		59	Formulário de Cadastro
		60	Contrato Prestação de Serviço
		61	Documento Sócio
		62	Dossiê de Análise
		63	Contrato
		64	Comprovante Despachante
		65	Comprovante de Débito
		66	Comprovante Crédito Cliente
	*/
	TipoDocumento int `json:"tipoDocumento,omitempty" validate:"required,oneof= 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 40 41 42 50 51 52 53 54 55 56 57 58 59 60 61 62 63 64 65 66"`
}
