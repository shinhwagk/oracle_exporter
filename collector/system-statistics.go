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
	registerCollector("systemStats-10g", NewSysstatCollector)
	registerCollector("systemStats-11g", NewSysstatCollector)
}

// NewSysstatCollector
func NewSysstatCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["user commits"] = createNewDesc("sysstat", "commit_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["user rollbacks"] = createNewDesc("sysstat", "rollback_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["execute count"] = createNewDesc("sysstat", "execute_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["parse count (total)"] = createNewDesc("sysstat", "parse_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["DB time"] = createNewDesc("sysstat", "dbtime_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["redo size"] = createNewDesc("sysstat", "redo_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["parse count (hard)"] = createNewDesc("sysstat", "parse_hard_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["parse count (failures)"] = createNewDesc("sysstat", "parse_failures_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["parse count (describe)"] = createNewDesc("sysstat", "parse_describe_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["parse time cpu"] = createNewDesc("sysstat", "parse_time_cpu_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["parse time elapsed"] = createNewDesc("sysstat", "parse_time_elapsed_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["physical read total bytes"] = createNewDesc("sysstat", "phy_read_bytes_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
	descs["physical write total bytes"] = createNewDesc("sysstat", "phy_write_bytes_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"class"}, nil)
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
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, class)
		} else {
			return errors.New("sysstat desc no exist")
		}

	}
	return nil
}

const sysstatSQL = `
SELECT name,
			 value,
			 decode(class, 1, 'User', 2,
              				 	'Read', 4,
              					'Enqueue', 8,
              					'Cache',  16,
              					'OS', 32,
              					'Real Application Clusters', 64,
              					'SQL', 128,
              					'Debug', 'null') class
  FROM v$sysstat
 WHERE name IN ('parse count (total)',
								'parse count (hard)',
								'parse count (failures)',
								'parse count (describe)',
								'parse time cpu',
								'parse time elapsed',
								'execute count',
								'user commits',
								'user rollbacks',
								'DB time',
								'redo size',
								'physical read total bytes',
								'physical write total bytes')`
