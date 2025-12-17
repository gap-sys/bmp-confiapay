package models

type BuscarProposta struct {
	CodigoProposta string `json:"CodigoProposta,omitempty" validate:"required_without=CodigoOperacao"`
	CodigoOperacao string `json:"CodigoOperacao,omitempty" validate:"required_without=CodigoProposta"`
}

func (b BuscarProposta) Validate() error {
	if b.CodigoOperacao == "" && b.CodigoProposta == "" {
		return NewAPIError("", "Insira CodigoProposta ou CodigoOperacao", "")

	}
	return nil
}

type BuscarPropostaFrontend struct {
	IdProposta           int    `json:"IdProposta" validate:"required_without=NumeroAcompanhamento"`
	NumeroAcompanhamento string `json:"NumeroAcompanhamento" validate:"required_without=IdProposta"`
}

func (b BuscarPropostaFrontend) Validate() error {
	if b.IdProposta <= 0 && b.NumeroAcompanhamento == "" {
		return NewAPIError("", "Insira IdProposta ou NumeroAcompanhamento", "")

	}
	return nil
}
