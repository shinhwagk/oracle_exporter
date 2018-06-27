package collector

import (
	"database/sql"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

type sysstatCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("sysstat", cMin, defaultEnabled, NewSysstatCollector)
}

// NewSysstatCollector returns a new Collector exposing session activity statistics.
func NewSysstatCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["user commits"] = newDesc("sesstat", "commit_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"name", "class"}, nil)
	descs["user rollbacks"] = newDesc("sesstat", "rollback_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"name", "class"}, nil)
	descs["execute count"] = newDesc("sesstat", "execute_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"name", "class"}, nil)
	descs["parse count (total)"] = newDesc("sesstat", "parse_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"name", "class"}, nil)
	descs["DB time"] = newDesc("sesstat", "dbtime_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"name", "class"}, nil)
	descs["redo size"] = newDesc("sesstat", "redo_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"name", "class"}, nil)
	return &sysstatCollector{descs}, nil
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

		desc, ok := c.descs[name]

		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value)
		} else {
			return errors.New("sesstat desc no exist")
		}

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
  FROM v$sysstat
 WHERE AND sn.name IN ('parse count (total)',
                    'execute count',
                    'user commits',
                    'user rollbacks',
                    'DB time',
                    'redo size')`
