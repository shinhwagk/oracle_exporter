package collector

import (
	"database/sql"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

type sessTimeModelCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("sessionTimeModel-10g", NewSessTimeModelCollector)
	registerCollector("sessionTimeModel-11g", NewSessTimeModelCollector)
}

// NewSessTimeModelCollector returns a new Collector exposing session activity statistics.
func NewSessTimeModelCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["DB time"] = newDesc(sessTimeModelSystemName, "db_time", "Generic counter metric from v$sesstat view in Oracle.", []string{"sid", "username"}, nil)
	return &sessTimeModelCollector{descs}, nil
}

func (c *sessTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessTimeModelSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, username, sid string
		var value float64
		if err := rows.Scan(&sid, &username, &name, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]

		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, sid, username)
		} else {
			return errors.New("system time model no exist")
		}

	}
	return nil
}

const (
	sessTimeModelSystemName = "sesstimemodel"
	sessTimeModelSQL        = `
	SELECT s.sid, s.username, stm.stat_name, stm.value
	FROM v$sess_time_model stm, v$session s
	WHERE stm.sid = s.sid AND s.username is not null AND stm.stat_name in ('DB time')`
)
