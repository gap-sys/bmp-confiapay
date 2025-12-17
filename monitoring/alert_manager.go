package monitoring

import (
	"cobranca-bmp/config"
	"time"
)

type AlertManager struct {
	alerts []Alert
	loc    *time.Location
}

type Alert struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
}

func NewAlertManager(loc *time.Location) *AlertManager {
	return &AlertManager{
		loc:    loc,
		alerts: make([]Alert, 0),
	}
}

func (am *AlertManager) CheckMetrics(metrics *PoolMetrics) {
	// Verificar utilização crítica
	if metrics.UtilizationPercent >= 90 {
		am.addAlert(am.createAlert("critical", "Pool utilization critical", "utilization_percent",
			metrics.UtilizationPercent, 90))
	} else if metrics.UtilizationPercent > 75 {
		am.addAlert(am.createAlert("warning", "Pool utilization high", "utilization_percent",
			metrics.UtilizationPercent, 75))
	}

	// Verificar tempo de espera
	if waitDuration, ok := metrics.Stats["wait_duration_ms"].(int64); ok {
		if waitDuration > 100 {

			am.addAlert(am.createAlert("warning", "High wait duration for connections", "wait_duration_ms",
				float64(waitDuration), 100))
		}
	}

	// Verificar contagem de esperas
	if waitCount, ok := metrics.Stats["wait_count"].(int64); ok {
		if waitCount > 50 {
			am.addAlert(am.createAlert("warning", "High wait count for connections", "wait_count",
				float64(waitCount), 50))
		}
	}
}

func (am *AlertManager) addAlert(alert Alert) {
	am.alerts = append(am.alerts, alert)

	// Manter apenas os últimos alertas
	if len(am.alerts) > config.DB_POOL_MONITORING_MAX_ALERTS {
		am.alerts = am.alerts[:config.DB_POOL_MONITORING_MAX_ALERTS]
	}
}

func (am *AlertManager) createAlert(level, message, metric string, value, threshold float64) Alert {
	alert := Alert{
		Level:     level,
		Message:   message,
		Timestamp: time.Now().In(am.loc),
		Metric:    metric,
		Value:     value,
		Threshold: threshold,
	}

	return alert
}

func (am *AlertManager) GetRecentAlerts(since time.Duration) []Alert {
	cutoff := time.Now().Add(-since)
	recent := make([]Alert, 0)

	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoff) {
			recent = append(recent, alert)
		}
	}

	return recent
}
