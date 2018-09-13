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
	desc := createNewDesc("session", "time_model", "oracle session level time model.", []string{"sid", "serial", "name", "username"}, nil)
	return &sessTimeModelCollector{desc}
}

func (c *sessTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessTimeModelSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sid, serial, uname, sname string
		var value float64
		if err := rows.Scan(&sid, &serial, &uname, &sname, &value); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, sid, serial, sname, uname)
	}
	return nil
}

const (
	sessTimeModelSystemName = "sesstimemodel"
	sessTimeModelSQL        = `
	SELECT s.sid, s.serial#, s.username, stm.stat_name, stm.value
	FROM v$sess_time_model stm, v$session s
  WHERE stm.sid = s.sid AND s.username IS NOT NULL`
)
