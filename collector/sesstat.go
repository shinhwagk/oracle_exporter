package collector

import (
	"database/sql"
	"errors"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	sessstatFlag = flag.Bool("collector.sesstat", true, "for session activity collector")
)

type sesstatCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("sesstat", defaultEnabled, NewSysstatCollector)
}

// NewSesstatCollector returns a new Collector exposing session activity statistics.
func NewSesstatCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["commit_total"] = newDesc("sesstat", "commit_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username"}, nil)
	descs["rollback_total"] = newDesc("sesstat", "rollback_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username"}, nil)
	descs["execute_total"] = newDesc("sesstat", "execute_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username"}, nil)
	descs["parse_total"] = newDesc("sesstat", "parse_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username"}, nil)
	return &sysstatCollector{descs}, nil
}

func (c *sesstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sesstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, username string
		var value float64
		if err := rows.Scan(&username, &name, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]
		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, username)
		} else {
			return errors.New("sesstat desc no exist")
		}
	}
	return nil
}

const sesstatSQL = `
SELECT
	USERNAME,
  CASE NAME
    WHEN 'parse count (total)' THEN 'parse_total'
    WHEN 'execute count'       THEN 'execute_total'
    WHEN 'user commits'        THEN 'commit_total'
    WHEN 'user rollbacks'      THEN 'rollback_total'
  END name,
  VALUE
FROM
  (SELECT sn.NAME, SUM(ss.VALUE) value, s.USERNAME
  FROM v$sesstat ss, v$statname sn, v$session s
  WHERE s.sid       = ss.SID
  AND ss.STATISTIC# = sn.STATISTIC#
  AND sn.name      IN ('parse count (total)', 'execute count', 'user commits', 'user rollbacks')
  GROUP BY s.USERNAME, sn.NAME)`
