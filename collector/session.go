package collector

import (
	"database/sql"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	sessionFlag = flag.Bool("collector.session", true, "for session activity collector")
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

func (c *sessionCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query("SELECT status, type, COUNT(*) FROM v$session GROUP BY status, type")
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
		ch <- prometheus.MustNewConstMetric(c.sessionActivity,
			prometheus.GaugeValue,
			count,
			status,
			sessionType,
		)
	}
	return nil
}
