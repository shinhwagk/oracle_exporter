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
	descs["user commits"] = newDesc("sysstat", "commit_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["user rollbacks"] = newDesc("sysstat", "rollback_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["execute count"] = newDesc("sysstat", "execute_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["parse count (total)"] = newDesc("sysstat", "parse_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["DB time"] = newDesc("sysstat", "dbtime_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["redo size"] = newDesc("sysstat", "redo_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["user calls"] = newDesc("sysstat", "useralls_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["db block changes"] = newDesc("sysstat", "dbblockchanges_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["parse count (hard)"] = newDesc("sysstat", "parsehard_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["physical reads"] = newDesc("sysstat", "physicalreads_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["physical writes"] = newDesc("sysstat", "physicalwrites_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["logons cumulative"] = newDesc("sysstat", "logons_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["session logical reads"] = newDesc("sysstat", "logicalreads_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["physical write total bytes"] = newDesc("sysstat", "logicalwrite_bytes_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["physical read total bytes"] = newDesc("sysstat", "logicalread_bytes_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	return &sysstatCollector{descs}, nil
}

func (c *sysstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]
		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value)
		} else {
			return errors.New("sysstat desc no exist")
		}
	}
	return nil
}

const sysstatSQL = `
SELECT name, value
  FROM v$sysstat
 WHERE name IN ('execute count',
                'user commits',
                'user rollbacks',
                'DB time',
                'redo size',
                'user calls',
                'db block changes',
                'parse count (total)',
                'parse count (hard)',
                'physical reads',
                'physical writes',
								'logons cumulative',
								'session logical reads',
								'physical write total bytes',
								'physical read total bytes')`
