package repository

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
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

func (s *ParcelaRepo) Update(id, idStatus int) error {

	now := time.Now().In(s.location)

	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()
	_, err := s.db.ExecContext(ctx, `
	UPDATE
	      bmp_cobranca 
	SET  
	    id_status=$1,
		updated_at=$2



		 where id=$3`, idStatus, now, id)

	if err != nil {
		var logData = map[string]any{"id_simulacao": id, "id_status": idStatus}
		helpers.LogError(s.ctx, s.logger, s.location, "db", "", "Erro ao realizar update na tabela simulacao_status", err.Error(), logData)
		return err
	}

	return nil
}
