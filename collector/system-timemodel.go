package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sysTimeModelCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("systemTimeModel-10g", NewSysTimeModelCollector)
	registerCollector("systemTimeModel-11g", NewSysTimeModelCollector)
}

// NewSysTimeModelCollector
func NewSysTimeModelCollector() Collector {
	desc := createNewDesc("system", "time_model", "oracle_system_time_model", []string{"name"}, nil)
	return &sysTimeModelCollector{desc}
}

func (c *sysTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(systemTimeModelSQL)
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
	systemTimeModelSQL = `SELECT stat_name, value FROM v$sys_time_model`
)
