package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessionCollector struct {
	sessionActivity *prometheus.Desc
}

func init() {
	registerCollector("session", defaultEnabled, NewSessionCollector)
}

// NewSessionCollector returns a new Collector exposing session activity statistics.
func NewSessionCollector() (Collector, error) {
	return &sessionCollector{prometheus.NewDesc(prometheus.BuildFQName(namespace, "sessions", "activity"),
		"Gauge metric with count of sessions by status and type", []string{"status", "type"}, nil)}, nil
}

func (c *sessionCollector) Update(ch chan<- prometheus.Metric) error {
	var (
		rows *sql.Rows
		err  error
	)
	db, err := sql.Open("mysql", "")
	rows, err = db.Query("SELECT status, type, COUNT(*) FROM v$session GROUP BY status, type")
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			status, sessionType string
			count               float64
		)
		if err := rows.Scan(&status, &sessionType, &count); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.sessionActivity,
			prometheus.GaugeValue,
			count,
			status,
			sessionType,
		)
	}
	return nil
}
