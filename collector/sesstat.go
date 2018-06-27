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
	registerCollector("sesstat", cMin, defaultEnabled, NewSesstatCollector)
}

// NewSesstatCollector returns a new Collector exposing session activity statistics.
func NewSesstatCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["user commits"] = newDesc("sesstat", "commit_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	descs["user rollbacks"] = newDesc("sesstat", "rollback_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	descs["execute count"] = newDesc("sesstat", "execute_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	descs["parse count (total)"] = newDesc("sesstat", "parse_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	descs["DB time"] = newDesc("sesstat", "dbtime_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	descs["redo size"] = newDesc("sesstat", "redo_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	return &sesstatCollector{descs}, nil
}

func (c *sesstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sesstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sid, name, username string
		var value float64
		if err := rows.Scan(&sid, &name, &username, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]

		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, username, sid, name)
		} else {
			return errors.New("sesstat statistic no exist.")
		}
	}
	return nil
}

const sesstatSQL = `
SELECT s.sid, sn.name, s.USERNAME, SUM(ss.VALUE)
  FROM v$sesstat ss, v$statname sn, v$session s
 WHERE s.sid = ss.SID
   AND ss.STATISTIC# = sn.STATISTIC#
   AND s.username is not null
   AND ss.value > 0
   AND sn.name IN ('parse count (total)',
                   'execute count',
                   'user commits',
                   'user rollbacks',
                   'DB time',
                   'redo size')
 GROUP BY s.USERNAME, sn.NAME, s.sid`
