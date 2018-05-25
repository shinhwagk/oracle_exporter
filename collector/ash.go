package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type ashCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("ash", cMin, defaultEnabled, NewASHCollector)
}

// NewASHCollector returns a new Collector exposing ash activity statistics.
func NewASHCollector() (Collector, error) {
	return &ashCollector{
		newDesc("ash", "info", "Gauge metric with count of sessions by status and type", []string{"username", "machine", "type"}, nil),
	}, nil
}

func (c *ashCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(ashSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionType, username, machine string

		if err = rows.Scan(&sessionType, &username, &machine); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, 1, username, machine, sessionType)
	}
	return nil
}

const ashSQL = `
select ash.session_type, du.username, ash.machine
  from v$active_session_history ash, dba_users du
 where du.user_id = ash.user_id
   and ash.SAMPLE_TIME >= trunc(sysdate, 'MI') - 1 / 24 / 60
   and ash.SAMPLE_TIME < trunc(sysdate, 'MI')`
