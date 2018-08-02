package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessEventCollector struct {
	descs [3]*prometheus.Desc
}

type sessClassCollector struct {
	descs [2]*prometheus.Desc
}

func init() {
	registerCollector("sessionEvent-10g", NewSessEventCollector)
	registerCollector("sessionEvent-11g", NewSessEventCollector)
}

// NewSessEventCollector
func NewSessEventCollector() Collector {
	descs := [3]*prometheus.Desc{
		createNewDesc("sessevent", "waits_total", "Generic counter metric from v$session_event view in Oracle.", []string{"username", "event", "class", "sid"}, nil),
		createNewDesc("sessevent", "waited_time_total", "Generic counter metric from v$session_event view in Oracle.", []string{"username", "event", "class", "sid"}, nil),
		createNewDesc("sessevent", "timeout_total", "Generic counter metric from v$session_event view in Oracle.", []string{"username", "event", "class", "sid"}, nil),
	}
	return &sessEventCollector{descs}
}

// NewSessClassCollector
func NewSessClassCollector() Collector {
	descs := [2]*prometheus.Desc{
		createNewDesc("sessclass", "waits_total", "Generic counter metric from v$session_event view in Oracle.", []string{"username", "serial", "class", "sid"}, nil),
		createNewDesc("sessclass", "waited_time_total", "Generic counter metric from v$session_event view in Oracle.", []string{"username", "serial", "class", "sid"}, nil),
	}
	return &sessClassCollector{descs}
}

func (c *sessEventCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessEventSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var username, event, class, sid string
		var waits, timeWaited, timeOut float64
		if err := rows.Scan(&sid, &username, &event, &class, &waits, &timeWaited, &timeOut); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, username, event, class, sid)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeWaited, username, event, class, sid)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, timeOut, username, event, class, sid)
	}
	return nil
}

func (c *sessClassCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessClassSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var username, class, sid, serial string
		var waits, timeWaited float64
		if err := rows.Scan(&sid, &serial, &class, &waits, &timeWaited, &username); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, username, serial, class, sid)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeWaited, username, serial, class, sid)
	}
	return nil
}

const (
	sessEventSQL = `
SELECT ss.sid,
       ss.username,
			 se.event,
			 se.wait_class,
       sum(se.total_waits),
			 sum(se.time_waited_micro),
			 sum(se.TOTAL_TIMEOUTS)
  FROM v$session_event se, v$session ss
 where ss.sid = se.sid
	 and se.total_waits > 0
	 and se.wait_class <> 'Idle'
   and ss.username is not null
 group by ss.sid, ss.username, se.event, se.wait_class`

	sessClassSQL = `
select swc.sid,
	swc.serial#,
	swc.wait_class,
	swc.TOTAL_WAITS,
	swc.time_waited,
	s.username
from V$SESSION_WAIT_CLASS swc, v$session s
where swc.sid = s.sid
	and swc.serial# = s.serial#
	and s.username is not null
	and swc.wait_class <> 'Idle'
	and swc.TOTAL_WAITS > 0`
)
