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
	desc := createNewDesc("session", "statistic", "empty", []string{"class", "name", "username"}, nil)
	return &sesstatCollector{desc}
}

func (c *sesstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sesstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, uname, class string
		var value float64
		if err := rows.Scan(&name, &uname, &class, &value); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value, class, name, uname)
	}
	return nil
}

const sesstatSQL = `
SELECT sn.name, s.username,
       decode(sn.class,
              1,  'User',
              2,  'Read',
              4,  'Enqueue',
              8,  'Cache',
              16, 'OS',
              32, 'Real Application Clusters',
              64, 'SQL',
              128, 'Debug', 'null'),
				sum(ss.value)
FROM v$sesstat ss, v$statname sn, v$session s
WHERE s.sid = ss.sid
	AND ss.statistic# = sn.statistic#
	AND s.username IS NOT NULL
GROUP BY s.username, sn.name, sn.class`
