package repository

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"database/sql"
	"log/slog"
	"strconv"
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
	/*
			_, err := s.db.ExecContext(ctx, `
			UPDATE
			      bmp_cobrancas
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
		        parcela=$10,
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
			) */

	_, err := s.db.ExecContext(ctx, `
	INSERT INTO bmp_cobrancas (
    id_proposta,
    id_securitizadora,
	id_convenio,
    numero_acompanhamento,
    codigo_liquidacao,
    numero_ccb,
    url_webhook,
    id_proposta_parcela,
    data_vencimento,
    data_expiracao,
    parcela,
    id_forma_cobranca,
    updated_at
) VALUES (
    $1,  
    $2,  
    $3,  
    $4,  
    $5,  
    $6, 
    $7, 
    $8,  
    $9,  
    $10, 
    $11, 
    $12,
	$13

)
ON CONFLICT (id_proposta_parcela)
DO UPDATE SET
    id_proposta           = EXCLUDED.id_proposta,
    id_securitizadora     = EXCLUDED.id_securitizadora,
	id_convenio            = EXCLUDED.id_convenio,
    numero_acompanhamento = EXCLUDED.numero_acompanhamento,
    codigo_liquidacao     = EXCLUDED.codigo_liquidacao,
    numero_ccb            = EXCLUDED.numero_ccb,
    url_webhook           = EXCLUDED.url_webhook,
	id_proposta_parcela   = EXCLUDED.id_proposta_parcela,
    data_vencimento       = EXCLUDED.data_vencimento,
    data_expiracao        = EXCLUDED.data_expiracao,
    parcela        = EXCLUDED.parcela,
    id_forma_cobranca     = EXCLUDED.id_forma_cobranca,
    updated_at            = EXCLUDED.updated_at;
	`,
		data.IdProposta,
		data.IdSecuritizadora,
		data.IdConvenio,
		data.NumeroAcompanhamento,
		codigoLiquidacao,
		data.NumeroCCB,
		data.UrlWebhook,
		data.IdPropostaParcela,
		data.DataVencimento,
		data.DataExpiracao,
		data.NumeroParcela,
		data.TipoCobranca,
		now,
	)

	if err != nil {
		//	var logData = map[string]any{s}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao realizar update na tabela simulacao_status", err.Error(), data)
		return isConnError(s.db, err), err
	}

	return false, nil
}

func (s *ParcelaRepo) UpdateCancelamentoCobranca(data models.CancelarCobrancaFrontendInput, codigoLiquidacao string) (bool, error) {

	now := time.Now().In(s.location)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()
	/*
			_, err := s.db.ExecContext(ctx, `
			UPDATE
			      bmp_cobrancas
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
		        parcela=$10,
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
			) */

	_, err := s.db.ExecContext(ctx, `
	INSERT INTO bmp_cobrancas (
    id_proposta,
    id_securitizadora,
	id_convenio,
    numero_acompanhamento,
    codigo_liquidacao,
    numero_ccb,
    url_webhook,
    id_proposta_parcela,
	parcela,
    updated_at
) VALUES (
    $1,  
    $2,  
    $3,  
    $4,  
    $5,  
    $6,  
    $7, 
    $8,
	$9,
	$10
	
	
)
ON CONFLICT (id_proposta_parcela)
DO UPDATE SET
    id_proposta           = EXCLUDED.id_proposta,
    id_securitizadora     = EXCLUDED.id_securitizadora,
	id_convenio            = EXCLUDED.id_convenio,
    numero_acompanhamento = EXCLUDED.numero_acompanhamento,
    codigo_liquidacao     = EXCLUDED.codigo_liquidacao,
    numero_ccb            = EXCLUDED.numero_ccb,
    url_webhook           = EXCLUDED.url_webhook,
	id_proposta_parcela   = EXCLUDED.id_proposta_parcela,
	parcela        = EXCLUDED.parcela,
    updated_at            = EXCLUDED.updated_at
	`,
		data.IdProposta,
		data.IdSecuritizadora,
		data.IdConvenio,
		data.NumeroAcompanhamento,
		codigoLiquidacao,
		data.NumeroCCB,
		data.UrlWebhook,
		data.IdPropostaParcela,
		data.NumeroParcela,
		now,
	)

	if err != nil {
		//	var logData = map[string]any{s}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao realizar update na tabela bmp_cobrancas", err.Error(), data)
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
	      bmp_cobrancas 
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

func (s *ParcelaRepo) FindByCodLiquidacao(codigoLiquidacao string, numeroCCB int) (models.CobrancaBMP, error) {
	var cobranca models.CobrancaBMP
	cobranca.CodigoLiquidacao = codigoLiquidacao
	numeroCCBString := strconv.Itoa(numeroCCB)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()

	err := s.db.QueryRowContext(ctx, `
			SELECT
			    id_proposta,
				numero_acompanhamento,
		        id_securitizadora,
				id_convenio, 
		        numero_ccb,
		        url_webhook,
		        id_proposta_parcela,
		        parcela,
		        id_forma_cobranca
			
			FROM 
			    bmp_cobrancas

	        WHERE
				codigo_liquidacao=$1 AND numero_ccb=$2`,
		codigoLiquidacao, numeroCCBString,
	).Scan(&cobranca.IdProposta,
		&cobranca.NumeroAcompanhamento,
		&cobranca.IdSecuritizadora,
		&cobranca.IdConvenio,
		&cobranca.NumeroCCB,
		&cobranca.UrlWebhook,
		&cobranca.IdPropostaParcela,
		&cobranca.NumeroParcela,
		&cobranca.IdFormaCobranca)
	if err != nil {
		var logData = map[string]any{"codigo_liquidacao": codigoLiquidacao, "numero_ccb": numeroCCB}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao buscar em cobranca_bmp por código de liquidação", err.Error(), logData)
		return models.CobrancaBMP{}, err
	}

	return cobranca, nil

}

func (s *ParcelaRepo) FindByNumParcela(numParcela int, numeroCCB int) (models.CobrancaBMP, error) {
	var cobranca models.CobrancaBMP
	cobranca.NumeroParcela = numParcela
	numeroCCBString := strconv.Itoa(numeroCCB)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()

	err := s.db.QueryRowContext(ctx, `
			SELECT
			    id_proposta,
				numero_acompanhamento,
		        id_securitizadora,
				id_convenio, 
		        numero_ccb,
		        url_webhook,
		        id_proposta_parcela,
		      codigo_liquidacao,
		        id_forma_cobranca
			
			FROM 
			    bmp_cobrancas

			WHERE
				parcela=$1 AND numero_ccb=$2`,
		numParcela, numeroCCBString,
	).Scan(&cobranca.IdProposta,
		&cobranca.NumeroAcompanhamento,
		&cobranca.IdSecuritizadora,
		&cobranca.IdConvenio,
		&cobranca.NumeroCCB,
		&cobranca.UrlWebhook,
		&cobranca.IdPropostaParcela,
		&cobranca.CodigoLiquidacao,
		&cobranca.IdFormaCobranca)
	if err != nil {
		var logData = map[string]any{"parcela": numParcela, "numero_ccb": numeroCCB}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao buscar em cobranca_bmp por número de parcela", err.Error(), logData)
		return models.CobrancaBMP{}, err
	}

	return cobranca, nil

}

func (s *ParcelaRepo) FindByDataVencimento(dataVencimento string, numeroCCB int) (models.CobrancaBMP, error) {
	var cobranca models.CobrancaBMP
	numeroCCBString := strconv.Itoa(numeroCCB)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()

	err := s.db.QueryRowContext(ctx, `
			SELECT
			    id_proposta,
				numero_acompanhamento,
		        id_securitizadora,
				id_convenio, 
		        numero_ccb,
		        url_webhook,
		        id_proposta_parcela,
		        codigo_liquidacao,
			    parcela,
		        id_forma_cobranca
			
			FROM 
			    bmp_cobrancas

			WHERE
				data_vencimento=$1 AND numero_ccb=$2`,
		dataVencimento, numeroCCBString,
	).Scan(&cobranca.IdProposta,
		&cobranca.NumeroAcompanhamento,
		&cobranca.IdSecuritizadora,
		&cobranca.IdConvenio,
		&cobranca.NumeroCCB,
		&cobranca.UrlWebhook,
		&cobranca.IdPropostaParcela,
		&cobranca.CodigoLiquidacao,
		&cobranca.NumeroParcela,
		&cobranca.IdFormaCobranca)
	if err != nil {
		var logData = map[string]any{"data_vencimento": dataVencimento, "numero_ccb": numeroCCB}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao buscar em cobranca_bmp por data de vencimento", err.Error(), logData)
		return models.CobrancaBMP{}, err
	}

	return cobranca, nil

}
