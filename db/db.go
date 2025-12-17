package db

import (
	"cobranca-bmp/config"
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

// Abre uma conexão com postgres utilizando uma string de conexão.
func Open(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
