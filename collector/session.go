package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessionCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("session", defaultEnabled, NewSessionCollector)
}

// NewSessionCollector returns a new Collector exposing session activity statistics.
func NewSessionCollector() (Collector, error) {
	return &sessionCollector{
		newDesc("sessions", "activity", "Gauge metric with count of sessions by status and type", []string{"status", "type"}, nil),
	}, nil
}

func (c *sessionCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessionSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			status, sessionType string
			count               float64
		)
		if err = rows.Scan(&status, &sessionType, &count); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, count, status, sessionType)
	}
	return nil
}

const sessionSQL = "SELECT status, type, COUNT(*) FROM v$session GROUP BY status, type"
