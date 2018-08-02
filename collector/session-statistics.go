package collector

import (
	"database/sql"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

type sesstatCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("sessionStats-10g", NewSesstatCollector)
	registerCollector("sessionStats-11g", NewSesstatCollector)
}

// NewSesstatCollector returns a new Collector exposing session activity statistics.
func NewSesstatCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["user commits"] = createNewDesc("sesstat", "commit_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["user rollbacks"] = createNewDesc("sesstat", "rollback_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["execute count"] = createNewDesc("sesstat", "execute_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["parse count (total)"] = createNewDesc("sesstat", "parse_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["DB time"] = createNewDesc("sesstat", "dbtime_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["redo size"] = createNewDesc("sesstat", "redo_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["parse count (hard)"] = createNewDesc("sesstat", "parse_hard_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["parse count (failures)"] = createNewDesc("sesstat", "parse_failures_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["parse count (describe)"] = createNewDesc("sesstat", "parse_describe_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["parse time cpu"] = createNewDesc("sesstat", "parse_time_cpu_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	descs["parse time elapsed"] = createNewDesc("sesstat", "parse_time_elapsed_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"serial", "username", "sid", "class"}, nil)
	return &sesstatCollector{descs}, nil
}

func (c *sesstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sesstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var serial, sid, name, username, class string
		var value float64
		if err := rows.Scan(&sid, &serial, &name, &username, &class, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]

		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, serial, username, sid, class)
		} else {
			return errors.New("sesstat statistic no exist")
		}

	}
	return nil
}

const sesstatSQL = `
SELECT s.sid,
       s.serial#,
       sn.name,
       s.username,
       decode(sn.class,
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
              'null'),
       ss.value
  FROM v$sesstat ss, v$statname sn, v$session s
 WHERE s.sid = ss.sid
   AND ss.statistic# = sn.statistic#
   AND s.username IS NOT NULL
   AND ss.value > 0
   AND sn.name IN ('parse count (total)',
                   'parse count (hard)',
                   'parse count (failures)',
                   'parse count (describe)',
                   'parse time cpu',
                   'parse time elapsed',
                   'execute count',
                   'user commits',
                   'user rollbacks',
                   'DB time',
                   'redo size')`
