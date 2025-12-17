package config

import (
	"database/sql"
	"fmt"
	"time"
)

type DBPoolConfig struct {
	MaxOpenConns    int           `json:"maxOpenConns"`
	MaxIdleConns    int           `json:"maxIdleConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime"`
}

type APIDBPoolConfig struct {
	MaxOpenConns    int  `json:"maxOpenConns"`
	MaxIdleConns    int  `json:"maxIdleConns"`
	ConnMaxLifetime int  `json:"connMaxLifetime"`
	ConnMaxIdleTime int  `json:"connMaxIdleTime"`
	Reconnect       bool `json:"reconnect"`
	Default         bool `json:"default"`
}

func NewDBPoolConfigFromEnv() *DBPoolConfig {
	return &DBPoolConfig{
		MaxOpenConns:    DB_MAX_OPEN_CONNS,
		MaxIdleConns:    DB_MAX_IDLE_CONNS,
		ConnMaxLifetime: DB_CONN_MAX_LIFETIME,
		ConnMaxIdleTime: DB_CONN_MAX_IDLE_TIME,
	}

}

func (c *DBPoolConfig) ApplyToDatabase(db *sql.DB) {
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
}

func (c *DBPoolConfig) Validate() error {
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("MaxOpenConns deve ser maior que 0")
	}
	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("MaxIdleConns deve ser maior que 0")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("MaxIdleConns não pode ser maior que MaxOpenConns")
	}
	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("ConnMaxLifetime deve ser maior que 0")
	}
	if c.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("ConnMaxIdleTime deve ser maior que 0")
	}
	return nil
}
