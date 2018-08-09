package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type logCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("log-10g", NewLogCollector)
	registerCollector("log-11g", NewLogCollector)
}

// NewLogCollector desc.
func NewLogCollector() Collector {
	desc := createNewDesc("log", "sequence", "Gauge metric with count of sessions by status and type", nil, nil)
	return &logCollector{desc}
}

func (c *logCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(logSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var seq float64

		if err := rows.Scan(&seq); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, seq)
	}
	return nil
}

const logSQL = `
SELECT sequence# FROM v$log
 WHERE thread# = (SELECT instance_number FROM v$instance) AND status = 'CURRENT'
 AND first_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60`
