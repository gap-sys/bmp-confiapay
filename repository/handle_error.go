package repository

import (
	"cobranca-bmp/config"
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
)

// Verifica se um erro no banco de dados é por conta de conexão fechada.
func isConnError(db *sql.DB, err error) bool {
	// Se erro é nil, não é erro de conexão
	if err == nil {
		return false
	}

	// Verifica se é um erro de operação de rede
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}

	// Se db for nil, não dá para testar conexão
	if db == nil {
		return true
	}

	if err == sql.ErrConnDone {
		return true
	}

	if err == sql.ErrTxDone {
		return true
	}

	if errors.Is(err, context.Canceled) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Verifica mensagens comuns de erro de conexão
	errMsg := err.Error()

	connErrMessages := []string{
		"connect: connection refused",
		"no such host",
		"connection reset by peer",
		"connection reset",
		"server closed the connection unexpectedly",
		"connection terminated",
		"SSL is not enabled",
		"i/o timeout",
		"sql: database is closed",
		"EOF",
	}

	for _, msg := range connErrMessages {
		if strings.Contains(errMsg, msg) {
			return true
		}
	}

	// Tenta pingar com timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.DB_QUERY_TIMEOUT)
	defer cancel()
	if pingErr := db.PingContext(ctx); pingErr != nil {
		return true
	}

	return false
}
