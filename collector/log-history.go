package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type logHistoryCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("logHistory-10g", NewLogHistoryCollector)
	registerCollector("logHistory-11g", NewLogHistoryCollector)
}

// NewLogHistoryCollector desc.
func NewLogHistoryCollector() Collector {
	desc := createNewDesc("loghistory", "sequence", "Gauge metric with count of sessions by status and type", nil, nil)
	return &logHistoryCollector{desc}
}

func (c *logHistoryCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(logHistorySQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var max float64

		if err := rows.Scan(&max); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, max)
	}
	return nil
}

const logHistorySQL = `SELECT MAX(sequence#) FROM v$log_history WHERE thread# = (SELECT instance_number FROM v$instance)`
