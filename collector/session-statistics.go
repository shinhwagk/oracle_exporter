package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sesstatCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("sessionStats-10g", NewSesstatCollector)
	registerCollector("sessionStats-11g", NewSesstatCollector)
}

// NewSesstatCollector .
func NewSesstatCollector() Collector {
	desc := createNewDesc("session", "statistic", "empty", []string{"class", "name", "username", "sid"}, nil)
	return &sesstatCollector{desc}
}

func (c *sesstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sesstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sid, name, uname, class string
		var value float64
		if err := rows.Scan(&sid, &name, &uname, &class, &value); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, class, name, uname, sid)
	}
	return nil
}

const sesstatSQL = `
SELECT s.sid, sn.name, s.username,
       decode(sn.class,
              1,  'User',
              2,  'Read',
              4,  'Enqueue',
              8,  'Cache',
              16, 'OS',
              32, 'Real Application Clusters',
              64, 'SQL',
              128, 'Debug', 'null'),
				ss.value
FROM v$sesstat ss, v$statname sn, v$session s
WHERE s.sid = ss.sid
	AND ss.statistic# = sn.statistic#
	AND s.username IS NOT NULL
  AND ss.value >= 1
  AND sn.name IN ('parse count (total)',	
                   'parse count (hard)',	
                   'parse count (failures)',	
                   'parse count (describe)',	
                   'execute count',	
                   'user commits',	
									 'user rollbacks',
									 'physical read total bytes',
									 'physical write total bytes',
									 'redo size',
									 'leaf node 90-10 splits',
									 'leaf node splits',
									 'session logical reads')`
