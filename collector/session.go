package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessionCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("session-10g", NewSessionCollector)
	registerCollector("session-11g", NewSessionCollector)
}

// NewSessionCollector
func NewSessionCollector() Collector {
	return &sessionCollector{
		createNewDesc("", "session", "Gauge metric with count of sessions by status and type", []string{"username", "status", "type", "machine"}, nil),
	}
}

func (c *sessionCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessionSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			username, status, sessType, machine string
			count                               float64
		)
		if err = rows.Scan(&username, &status, &machine, &sessType, &count); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, count, username, status, sessType, machine)
	}
	return nil
}

const sessionSQL = `
SELECT NVL(username, 'null'), status, machine, type, COUNT(*)
  FROM v$session
 GROUP BY username, status, machine, type`
