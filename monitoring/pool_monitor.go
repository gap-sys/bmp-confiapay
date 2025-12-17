package monitoring

import (
	"cobranca-bmp/config"
	"cobranca-bmp/db"
	"cobranca-bmp/helpers"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PoolMonitor struct {
	dbManager            *db.DbManager
	metrics              *PoolMetrics
	PrometheusCollectors *PrometheusCollectors
	AlertManager         *AlertManager
	ticker               *time.Ticker
	loc                  *time.Location
	logger               *slog.Logger
	mu                   sync.RWMutex
	ctx                  context.Context
	restart              chan int
}

type PoolMetrics struct {
	Timestamp          time.Time      `json:"timestamp"`
	Stats              map[string]any `json:"stats"`
	HealthStatus       string         `json:"health_status"`
	Warnings           []string       `json:"warnings"`
	UtilizationPercent float64        `json:"utilization_percent"`
	IdlePercent        float64        `json:"idle_percent"`
}

type PrometheusCollectors struct {
	UtilizationPercent *prometheus.GaugeVec
	IddlePercent       *prometheus.GaugeVec
	AlertNum           *prometheus.HistogramVec
}

func NewPoolMonitor(dbManager *db.DbManager, loc *time.Location, logger *slog.Logger, prometheusCollector *PrometheusCollectors) *PoolMonitor {
	return &PoolMonitor{
		dbManager:            dbManager,
		metrics:              &PoolMetrics{},
		AlertManager:         NewAlertManager(loc),
		loc:                  loc,
		logger:               logger,
		PrometheusCollectors: prometheusCollector,
		restart:              make(chan int),
	}
}

func (pm *PoolMonitor) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	pm.ticker = ticker
	pm.ctx = ctx
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}

			helpers.LogDebug(ctx, pm.logger, pm.loc, "pool manager", "", "coletando métricas", nil)
			pm.collectMetrics()
			pm.checkAlerts()

		case <-ctx.Done():
			return

		case <-pm.restart:
			return

		}
	}
}

func (pm *PoolMonitor) collectMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stats := pm.dbManager.GetDetailedStats()

	pm.metrics = &PoolMetrics{
		Timestamp:    time.Now(),
		Stats:        stats,
		HealthStatus: pm.dbManager.GetHealthStatus(),
		Warnings:     pm.dbManager.GetHealthWarnings(),
	}

	// Calcular métricas derivadas
	if maxOpen, ok := stats["max_open_conns"].(int); ok && maxOpen > 0 {
		if inUse, ok := stats["in_use"].(int); ok {
			pm.metrics.UtilizationPercent = float64(inUse) / float64(maxOpen) * 100
			pm.PrometheusCollectors.UtilizationPercent.WithLabelValues().Set(pm.metrics.UtilizationPercent)
		}
	}

	if maxIdle, ok := stats["max_idle_conns"].(int); ok && maxIdle > 0 {
		if idle, ok := stats["idle"].(int); ok {
			pm.metrics.IdlePercent = float64(idle) / float64(maxIdle) * 100
			pm.PrometheusCollectors.IddlePercent.WithLabelValues().Set(pm.metrics.IdlePercent)
		}
	}
}

func (pm *PoolMonitor) GetCurrentMetrics() *PoolMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Retornar cópia para evitar race conditions
	metricsCopy := *pm.metrics
	return &metricsCopy
}

func (pm *PoolMonitor) checkAlerts() {
	// Obter métricas atuais para verificação
	metrics := pm.GetCurrentMetrics()
	alerts := pm.AlertManager.GetRecentAlerts(config.DB_POOL_MONITORING_ALERT_INTERVAL)
	var alerTypeCount = make(map[string]int)
	for _, a := range alerts {
		alerTypeCount[a.Level]++

	}

	for k, v := range alerTypeCount {
		pm.PrometheusCollectors.AlertNum.WithLabelValues(k).Observe(float64(v))

	}

	// Verificar utilização do pool

	if metrics.UtilizationPercent >= 90 {
		pm.AlertManager.addAlert(Alert{
			Level:     "critical",
			Message:   fmt.Sprintf("Alta utilização do pool de conexões: %.2f%%", metrics.UtilizationPercent),
			Timestamp: time.Now().In(pm.loc),
			Metric:    "utilization_percent",
			Value:     metrics.UtilizationPercent,
			Threshold: 90,
		})
	} else if metrics.UtilizationPercent > 75 {
		pm.AlertManager.addAlert(Alert{
			Level:     "warning",
			Message:   fmt.Sprintf("Alta utilização do pool de conexões: %.2f%%", metrics.UtilizationPercent),
			Timestamp: time.Now().In(pm.loc),
			Metric:    "utilization_percent",
			Value:     metrics.UtilizationPercent,
			Threshold: 75,
		})
	}

	// Verificar se há muitas conexões aguardando
	if waiting, ok := metrics.Stats["wait_count"].(int64); ok && waiting > 10 {
		pm.AlertManager.addAlert(Alert{
			Level:     "warning",
			Message:   fmt.Sprintf("Muitas conexões aguardando: %d", waiting),
			Timestamp: time.Now().In(pm.loc),
			Metric:    "wait_count",
			Value:     float64(waiting),
			Threshold: 10,
		})
	}

	// Verificar tempo de espera por conexões
	if waitDuration, ok := metrics.Stats["wait_duration"].(time.Duration); ok && waitDuration.Milliseconds() > 500 {
		pm.AlertManager.addAlert(Alert{
			Level:     "critical",
			Message:   fmt.Sprintf("Tempo de espera por conexões elevado: %s", waitDuration),
			Timestamp: time.Now().In(pm.loc),
			Metric:    "wait_duration",
			Value:     float64(waitDuration.Milliseconds()),
			Threshold: 500,
		})
	}

	// Verificar erros de conexão
	if maxLifetimeClosed, ok := metrics.Stats["max_lifetime_closed"].(int64); ok && maxLifetimeClosed > 5 {
		pm.AlertManager.addAlert(Alert{
			Level:     "info",
			Message:   fmt.Sprintf("Conexões fechadas por tempo de vida: %d", maxLifetimeClosed),
			Timestamp: time.Now().In(pm.loc),
			Metric:    "max_lifetime_closed",
			Value:     float64(maxLifetimeClosed),
		})
	}

	// Verificar saúde geral do pool
	if metrics.HealthStatus != "healthy" {
		pm.AlertManager.addAlert(Alert{
			Level:     "critical",
			Message:   fmt.Sprintf("Pool de conexões não está saudável: %v", metrics.Warnings),
			Timestamp: time.Now().In(pm.loc),
			Metric:    "health_status",
			Value:     0,
		})
	}
}

func (pm *PoolMonitor) SetInterval(interval time.Duration) {
	pm.ticker.Stop()
	pm.restart <- 1
	go pm.Start(pm.ctx, interval)
}
