package repository

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// Representa as operações realizadas na tabela de simulação
type ParcelaRepo struct {
	ctx      context.Context
	db       *sql.DB
	logger   *slog.Logger
	location *time.Location
}

func NewParcelaRepo(ctx context.Context, db *sql.DB, logger *slog.Logger, location *time.Location) *ParcelaRepo {
	return &ParcelaRepo{db: db,
		logger:   logger,
		location: location,
		ctx:      ctx,
	}
}

func (s *ParcelaRepo) SetDB(db *sql.DB) {
	s.db = db

}

func (s *ParcelaRepo) UpdateGeracaoCobranca(data models.GerarCobrancaFrontendInput, codigoLiquidacao string) (bool, error) {

	now := time.Now().In(s.location)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()
	_, err := s.db.ExecContext(ctx, `
	UPDATE
	      bmp_cobranca 
	SET  
	    id_proposta=$1,
        id_securitizadora=$2,
        numero_acompanhamento=$3,
        codigo_liquidacao"=$4,
        numero_ccb=$5,
        url_webhook=$6,
        id_proposta_parcela=$7,
        data_vencimento=$8,
        data_expiracao=$9,
        numero_parcela=$10,
        id_forma_cobranca=$11,
		updated_at=$12



		 where id_proposta_parcela=$13`,
		data.IdProposta,
		data.IdSecuritizadora,
		data.NumeroAcompanhamento,
		codigoLiquidacao,
		data.NumeroCCB,
		data.UrlWebhook, data.IdPropostaParcela,
		data.DataVencimento,
		data.DataExpiracao,
		data.NumeroParcela,
		data.TipoCobranca,
		now,
		data.IdPropostaParcela,
	)

	if err != nil {
		//	var logData = map[string]any{s}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao realizar update na tabela simulacao_status", err.Error(), data)
		return isConnError(s.db, err), err
	}

	return false, nil
}

func (s *ParcelaRepo) UpdateCodLiquidacao(IdPropostaParcela int, codigoLiquidacao string) (bool, error) {

	now := time.Now().In(s.location)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()
	_, err := s.db.ExecContext(ctx, `
	UPDATE
	      bmp_cobranca 
	SET  
        codigo_liquidacao"=$1, 
		updated_at=$2     
	WHERE
	     id_proposta_parcela=$3`,

		codigoLiquidacao,
		now,
		IdPropostaParcela,
	)

	if err != nil {
		var logData = map[string]any{"id_proposta_parcela": IdPropostaParcela, "codigo_liquidacao": codigoLiquidacao}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao realizar update na tabela simulacao_status", err.Error(), logData)
		return isConnError(s.db, err), err
	}

	return false, nil
}
