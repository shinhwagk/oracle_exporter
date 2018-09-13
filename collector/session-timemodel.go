package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessTimeModelCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("sessionTimeModel-10g", NewSessTimeModelCollector)
	registerCollector("sessionTimeModel-11g", NewSessTimeModelCollector)
}

// NewSessTimeModelCollector .
func NewSessTimeModelCollector() Collector {
	desc := createNewDesc("session", "time_model", "oracle session level time model.", []string{"name", "username"}, nil)
	return &sessTimeModelCollector{desc}
}

func (c *sessTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessTimeModelSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var uname, sname string
		var value float64
		if err := rows.Scan(&uname, &sname, &value); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, sname, uname)
	}
	return nil
}

const (
	sessTimeModelSystemName = "sesstimemodel"
	sessTimeModelSQL        = `
	SELECT s.username, stm.stat_name, sum(stm.value) FROM v$sess_time_model stm, v$session s
  WHERE stm.sid = s.sid AND s.username IS NOT NULL
  GROUP BY s.username, stm.stat_name`
)
