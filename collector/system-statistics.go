package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sysstatCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("systemStats-10g", NewSysstatCollector)
	registerCollector("systemStats-11g", NewSysstatCollector)
}

// NewSysstatCollector
func NewSysstatCollector() Collector {
	desc := createNewDesc("system", "statistic", "empty", []string{"class", "name"}, nil)
	return &sysstatCollector{desc}
}

func (c *sysstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var class, name string
		var value float64
		if err := rows.Scan(&class, &name, &value); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, class, name)

	}
	return nil
}

const sysstatSQL = `
SELECT decode(class, 1, 'User', 2, 'Read', 4, 'Enqueue', 8, 'Cache', 16, 'OS', 32, 'Real Application Clusters', 64, 'SQL', 128, 'Debug', 'Other'),
			 name,
			 value
  FROM v$sysstat WHERE value >= 1`
