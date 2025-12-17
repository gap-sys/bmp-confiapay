package db

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Representa um objeto que pode modificar internamente o db para o qual está apontando(será utilizado para repositórios)
type DBSetter interface {
	SetDB(db *sql.DB)
}

// Representa um gerenciador do banco de dados. Será utilizado para configurar conexões, retornar informações sobre seu estado e realizar shutdown.
type DbManager struct {
	ctx     context.Context
	mu      sync.Mutex
	db      *sql.DB
	loc     *time.Location
	logger  *slog.Logger
	name    string
	config  *config.DBPoolConfig
	setters []DBSetter
}

type dbStats struct {
	sql.DBStats
	MaxIddleConnections int    `json:"maxIddleConnections"`
	OpenTimeout         string `json:"openTimeout"`
	IddleTimeout        string `json:"iddleTimeout"`
}

func NewDBManager(ctx context.Context, db *sql.DB, name string, loc *time.Location, logger *slog.Logger, config *config.DBPoolConfig, setters ...DBSetter) *DbManager {

	var dbMng = &DbManager{
		ctx:     ctx,
		mu:      sync.Mutex{},
		db:      db,
		name:    name,
		loc:     loc,
		logger:  logger,
		setters: setters,
		config:  config,
	}

	dbMng.applyConfig()
	return dbMng
}

func (d *DbManager) applyConfig() {
	d.db.SetConnMaxIdleTime(d.config.ConnMaxIdleTime)
	d.db.SetConnMaxLifetime(d.config.ConnMaxLifetime)
	d.db.SetMaxIdleConns(d.config.MaxIdleConns)
	d.db.SetMaxOpenConns(d.config.MaxOpenConns)

}

// Set config altera as configurações do banco de dados.
// Será utilizado para fazer alterações via API
// Como apenas alguns valores podem ser passados, irá alterar apenas o que for enviado na request
// Mantendo os valores originais caso o campo seja omitido.
func (d *DbManager) setConfigs(config *config.DBPoolConfig) {
	if config == nil {
		return
	}

	if config.ConnMaxIdleTime <= 0 {
		config.ConnMaxIdleTime = d.config.ConnMaxIdleTime
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = d.config.ConnMaxLifetime
	}

	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = d.config.MaxIdleConns
	}

	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = d.config.MaxIdleConns
	}

	d.config = config

}

func (d *DbManager) Check() (map[string]any, error) {
	var info = make(map[string]any)
	if d.db == nil {
		err := errors.New("db nulo")
		info["erro"] = err.Error()
		return info, err
	}

	if err := d.PingWithTimeout(time.Second * 2); err != nil {
		info["erro"] = err.Error()
		return info, err
	}
	var bytes json.RawMessage
	stats := d.db.Stats()
	bytes, _ = json.Marshal(dbStats{
		stats,
		d.config.MaxIdleConns,
		d.config.ConnMaxLifetime.String(),
		d.config.ConnMaxIdleTime.String()})
	info["status"] = "ok"
	info["info"] = bytes
	return info, nil
}

func (d *DbManager) GetServiceName() string {
	return d.name
}

func (d *DbManager) Reconfig(config *config.DBPoolConfig, reconnect bool) any {
	d.mu.Lock()
	defer d.mu.Unlock()
	if reconnect {
		err := d.reconnect()
		if err != nil {
			return map[string]string{"mensagem": "erro ao se reconectar ao banco de dados", "erro": err.Error()}
		}
	}

	d.setConfigs(config)
	d.applyConfig()

	/*stts := d.db.Stats()
	var logData = dbStats{
		stts,
		d.config.MaxIdleConns,
		d.config.ConnMaxLifetime.String(),
		d.config.ConnMaxIdleTime.String()}
	*/
	data := d.GetDetailedStats()
	helpers.LogInfo(d.ctx, d.logger, d.loc, "db", "", "Modificação de atributos do banco realizada com sucesso", data)

	return data

}

func (d *DbManager) reconnect() error {

	if err := d.Close(); err != nil {
		return err
	}

	db, err := Open(config.DB_URL)
	if err != nil {
		helpers.LogError(d.ctx, d.logger, d.loc, "db", "", "Erro ao abrir nova conexão com o banco de dados", err, nil)
		return err
	}

	if err := d.PingWithTimeout(config.DB_QUERY_TIMEOUT); err != nil {
		helpers.LogError(d.ctx, d.logger, d.loc, "db", "", "Erro ao testar nova conexão com o banco de dados", err, nil)
		return err
	}

	d.db = db
	helpers.LogInfo(d.ctx, d.logger, d.loc, "db", "", "Reconexão com banco de dados realizada com sucesso", nil)
	if len(d.setters) > 0 {
		d.set()
	}
	return nil

}

func (d *DbManager) set() {
	for _, setter := range d.setters {
		if setter == nil {
			continue
		}
		setter.SetDB(d.db)

	}
}

func (d *DbManager) Close() error {
	if d.db == nil {
		err := errors.New("banco de dados é nulo")
		helpers.LogError(d.ctx, d.logger, d.loc, "db", "", "Erro ao realizar fechamento de conexões com banco de dados", err, nil)
		return err

	}
	err := d.db.Close()
	if err != nil {
		helpers.LogError(d.ctx, d.logger, d.loc, "db", "", "Erro ao realizar fechamento de conexões com banco de dados", err, nil)
		return err
	}
	helpers.LogInfo(d.ctx, d.logger, d.loc, "db", "", "Fechamento de conexão com banco de dados realizado com sucesso", nil)
	return nil

}

func (d *DbManager) GetDetailedStats() map[string]any {
	stats := d.db.Stats()

	return map[string]any{
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration_ms":    stats.WaitDuration.Milliseconds(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
		"max_open_conns":      d.config.MaxOpenConns,
		"max_idle_conns":      d.config.MaxIdleConns,
		"conn_max_lifetime":   d.config.ConnMaxLifetime.String(),
		"conn_max_idle_time":  d.config.ConnMaxIdleTime.String(),
	}
}

func (d *DbManager) GetHealthStatus() string {
	stats := d.db.Stats()

	utilization := float64(stats.InUse) / float64(d.config.MaxOpenConns)

	if utilization > 0.9 {
		return "critical"
	} else if utilization > 0.7 || stats.WaitCount > 100 {
		return "warning"
	}
	return "healthy"
}

func (d *DbManager) GetHealthWarnings() []string {
	stats := d.db.Stats()
	warnings := []string{}

	utilization := float64(stats.InUse) / float64(d.config.MaxOpenConns)

	if utilization > 0.8 {
		warnings = append(warnings, fmt.Sprintf("Alta utilização do pool: %.1f%%", utilization*100))
	}

	if stats.WaitCount > 50 {
		warnings = append(warnings, fmt.Sprintf("Muitas esperas por conexão: %d", stats.WaitCount))
	}

	if stats.WaitDuration.Milliseconds() > 100 {
		warnings = append(warnings, fmt.Sprintf("Tempo de espera alto: %dms", stats.WaitDuration.Milliseconds()))
	}

	if stats.MaxLifetimeClosed > 10 {
		warnings = append(warnings, "Muitas conexões fechadas por lifetime")
	}
	return warnings
}

func (d *DbManager) PingWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return d.db.PingContext(ctx)
}

func (d *DbManager) CheckHealth() bool {
	return d.PingWithTimeout(5*time.Second) == nil
}

func (d *DbManager) StressTest() {

	go func() {

		for range config.DB_MAX_OPEN_CONNS - 2 {
			go d.db.Query("SELECT FROM status WHERE id=16")

		}

	}()
}
