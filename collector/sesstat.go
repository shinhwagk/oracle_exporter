package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sesstatCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("sesstat", cMin, defaultEnabled, NewSesstatCollector)
}

// NewSesstatCollector returns a new Collector exposing session activity statistics.
func NewSesstatCollector() (Collector, error) {
	desc := newDesc("sesstat", "", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sid", "name"}, nil)
	return &sesstatCollector{desc}, nil
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

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, username, sid, name)
	}
	return nil
}

const sesstatSQL = `
SELECT s.sid, sn.name, s.USERNAME, SUM(ss.VALUE)
  FROM v$sesstat ss, v$statname sn, v$session s
 WHERE s.sid = ss.SID
	 AND ss.STATISTIC# = sn.STATISTIC#
   AND s.username is not null
 GROUP BY s.USERNAME, sn.NAME, s.sid`
