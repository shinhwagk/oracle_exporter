package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessEventCollector struct {
	descs [3]*prometheus.Desc
}

func init() {
	registerCollector("sessionevent", cMin, defaultEnabled, NewSessEventCollector)
}

// NewSessEventCollector returns a new Collector exposing session activity statistics.
func NewSessEventCollector() (Collector, error) {
	descs := [3]*prometheus.Desc{
		newDesc("sessevent", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"username", "event", "class", "sid"}, nil),
		newDesc("sessevent", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"username", "event", "class", "sid"}, nil),
		newDesc("sessevent", "timeout_total", "Generic counter metric from v$system_event view in Oracle.", []string{"username", "event", "class", "sid"}, nil),
	}
	return &sessEventCollector{descs}, nil
}

func (c *sessEventCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessEventSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var username, event, waitClass, sid string
		var waits, timeWaited, timeOut float64
		if err := rows.Scan(&sid, &username, &event, &waits, &timeWaited, &timeOut, &waitClass); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, username, event, waitClass, sid)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeWaited, username, event, waitClass, sid)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeOut, username, event, waitClass, sid)
	}
	return nil
}

const sessEventSQL = `
SELECT ss.sid,
       ss.username,
       se.event,
       sum(se.total_waits),
			 sum(se.time_waited_micro),
			 sum(se.TOTAL_TIMEOUTS)
       se.wait_class
  FROM v$session_event se, v$session ss
 where ss.sid = se.sid
   and se.total_waits > 0
   and ss.username is not null
 group by ss.sid, ss.username, se.event, se.wait_class`
