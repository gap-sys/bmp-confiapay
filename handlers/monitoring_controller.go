package handlers

import (
	"cobranca-bmp/db"
	"cobranca-bmp/monitoring"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthChecker interface {
	Check() (map[string]any, error)
}

type MonitoringController struct {
	dbManager       *db.DbManager
	poolMonitor     *monitoring.PoolMonitor
	loc             *time.Location
	servicesChecker HealthChecker
	redisChecker    HealthChecker
	rmqChecker      HealthChecker
}

func (m *MonitoringController) GetPrefix() string {
	return "monitoring"
}

func NewMonitoringController(loc *time.Location, dbManager *db.DbManager, poolMonitor *monitoring.PoolMonitor, checker, redisChecker, rmqCHecker HealthChecker) *MonitoringController {
	return &MonitoringController{
		loc:             loc,
		dbManager:       dbManager,
		poolMonitor:     poolMonitor,
		servicesChecker: checker,
		redisChecker:    redisChecker,
		rmqChecker:      rmqCHecker,
	}
}

func (m *MonitoringController) Route(r fiber.Router) {
	r.Get("/health", m.Health())
	r.Get("/rmq-health", m.CheckRabbitMQ())
	r.Get("/redis-health", m.CheckRedis())
	r.Get("/db-pool", m.GetDatabasePoolStats)
	r.Get("/db-health", m.GetDatabaseHealth)
	r.Get("/db-config", m.GetDatabaseConfig)
	r.Get("/db-alerts", m.GetDatabaseAlerts)

}

func (mc *MonitoringController) GetDatabasePoolStats(c *fiber.Ctx) error {
	metrics := mc.poolMonitor.GetCurrentMetrics()

	response := fiber.Map{
		"timestamp":     time.Now().In(mc.loc),
		"pool_stats":    metrics.Stats,
		"health_status": metrics.HealthStatus,
		"warnings":      metrics.Warnings,
		"metrics": fiber.Map{
			"utilization_percent": metrics.UtilizationPercent,
			"idle_percent":        metrics.IdlePercent,
		},
	}

	return c.JSON(response)
}

func (mc *MonitoringController) GetDatabaseHealth(c *fiber.Ctx) error {
	isHealthy := mc.dbManager.CheckHealth()
	healthStatus := mc.dbManager.GetHealthStatus()
	warnings := mc.dbManager.GetHealthWarnings()

	response := fiber.Map{
		"healthy":            isHealthy,
		"health_status":      healthStatus,
		"warnings":           warnings,
		"timestamp":          time.Now().In(mc.loc),
		"database_reachable": isHealthy,
	}

	// Definir status HTTP baseado na saúde
	statusCode := fiber.StatusOK
	if !isHealthy {
		statusCode = fiber.StatusServiceUnavailable
	} else if healthStatus == "critical" {
		statusCode = fiber.StatusInternalServerError
	} else if healthStatus == "warning" {
		statusCode = fiber.StatusAccepted
	}

	return c.Status(statusCode).JSON(response)
}

func (mc *MonitoringController) GetDatabaseConfig(c *fiber.Ctx) error {
	stats := mc.dbManager.GetDetailedStats()

	response := fiber.Map{
		"max_open_conns":     stats["max_open_conns"],
		"max_idle_conns":     stats["max_idle_conns"],
		"conn_max_lifetime":  stats["conn_max_lifetime"],
		"conn_max_idle_time": stats["conn_max_idle_time"],
		"timestamp":          time.Now().In(mc.loc),
	}

	return c.JSON(response)
}

func (mc *MonitoringController) GetDatabaseAlerts(c *fiber.Ctx) error {
	// Parâmetro opcional para filtrar alertas por período
	sinceParam := c.Query("since", "1h")
	since, err := time.ParseDuration(sinceParam)
	if err != nil {
		since = time.Hour
	}

	alerts := mc.poolMonitor.AlertManager.GetRecentAlerts(since)

	response := fiber.Map{
		"alerts":    alerts,
		"count":     len(alerts),
		"since":     since.String(),
		"timestamp": time.Now().In(mc.loc),
	}

	return c.JSON(response)
}

func (m *MonitoringController) Health() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := m.servicesChecker.Check()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err)
		}
		return c.JSON(data)

	}

}

func (m *MonitoringController) CheckRedis() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := m.redisChecker.Check()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err)
		}
		return c.JSON(data)

	}

}

func (m *MonitoringController) CheckRabbitMQ() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := m.rmqChecker.Check()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err)
		}
		return c.JSON(data)

	}

}
