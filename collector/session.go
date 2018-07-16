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
		newDesc("", "session", "Gauge metric with count of sessions by status and type", []string{"username", "status", "type", "machine"}, nil),
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
SELECT nvl(USERNAME, 'null'), STATUS, MACHINE, TYPE, COUNT(*)
  FROM V$SESSION
 GROUP BY USERNAME, STATUS, MACHINE, TYPE`
