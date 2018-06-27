package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sysstatCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("sysstat", cMin, defaultEnabled, NewSysstatCollector)
}

// NewSysstatCollector returns a new Collector exposing session activity statistics.
func NewSysstatCollector() (Collector, error) {
	desc := newDesc("sysstat", "", "Generic counter metric from v$sysstat view in Oracle.", []string{"name", "class"}, nil)
	return &sysstatCollector{desc}, nil
}

func (c *sysstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, class string
		var value float64
		if err := rows.Scan(&name, &value, &class); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, name, class)
	}
	return nil
}

const sysstatSQL = `
SELECT name,
       value,
       decode(class,
              1,
              'User',
              2,
              'Read',
              4,
              'Enqueue',
              8,
              'Cache',
              16,
              'OS',
              32,
              'Real Application Clusters',
              64,
              'SQL',
              128,
              'Debug',
              'Null') class
  FROM v$sysstat`
