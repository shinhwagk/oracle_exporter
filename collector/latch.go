package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type latchCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("latch-10g", NewLatchCollector)
	registerCollector("latch-11g", NewLatchCollector)
}

// NewLatchCollector
func NewLatchCollector() Collector {
	desc := createNewDesc("latch", "time_model", "oracle_system_time_model", []string{"name"}, nil)
	return &latchCollector{desc}
}

func (c *latchCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(latchSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var statName string
		var value float64
		if err := rows.Scan(&statName, &value); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, statName)
	}
	return nil
}

const (
	latchSQL = `SELECT stat_name, value FROM v$sys_time_model`
)
