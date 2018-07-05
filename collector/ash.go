package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type ashCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("ash", defaultEnabled, NewASHCollector)
}

// NewASHCollector returns a new Collector exposing ash activity statistics.
func NewASHCollector() (Collector, error) {
	desc := newDesc("ash", "sample", "Gauge metric with count of sessions by status and type", []string{"sample_id", "session_id", "session_serial", "event", "session_type", "username", "sql_id", "opname", "program", "machine"}, nil)
	return &ashCollector{desc}, nil
}

func (c *ashCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(ashSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sei, si, ssi, e, st, u, sli, so, p, m string

		if err = rows.Scan(&sei, &si, &ssi, &e, &st, &u, &sli, &so, &p, &m); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(1), sei, si, ssi, e, st, u, sli, so, p, m)
	}
	return nil
}

const ashSQL = `
select sample_id,
       session_id,
			 session_serial#,
       decode(session_state,'ON CPU', 'Wait for CPU', 'WAITING', event),
       session_type,
       (select username from dba_users where user_id = ash.user_id),
       nvl(sql_id, 'null'),
       nvl(sql_opname, 'null'),
       nvl(Program,'null'),
			 nvl(machine,'null'),
			 nvl(blocking_session,'null')
  from v$active_session_history ash
 where SAMPLE_TIME >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND SAMPLE_TIME < TRUNC(sysdate, 'MI')`
