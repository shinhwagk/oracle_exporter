package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sysEvent11GCollector struct {
	descs [sysEvent11GCollectorNumber]*prometheus.Desc
}

type sysEvent10GCollector struct {
	descs [sysEvent10GCollectorNumber]*prometheus.Desc
}

func init() {
	registerCollector("systemEvent-10g", NewSysEvent10GCollector)
	registerCollector("systemEvent-11g", NewSysEvent11GCollector)
}

// NewSysEventCollector
func NewSysEvent11GCollector() Collector {
	descs := [sysEvent11GCollectorNumber]*prometheus.Desc{
		createNewDesc("sysevent", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "timeout_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "waits_pg_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "waited_time_pg_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "timeout_pg_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
	}
	return &sysEvent11GCollector{descs}
}

// NewSysEventCollector
func NewSysEvent10GCollector() Collector {
	descs := [sysEvent10GCollectorNumber]*prometheus.Desc{
		createNewDesc("sysevent", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		createNewDesc("sysevent", "timeout_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
	}
	return &sysEvent10GCollector{descs}
}

func (c *sysEvent11GCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysEvent11GSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var event, class string
		var waits, timeWaited, timeOut, waitsfg, timeWaitedfg, timeOutfg float64
		if err := rows.Scan(&class, &event, &waits, &timeWaited, &timeOut, &waitsfg, &timeWaitedfg, &timeOutfg); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeWaited, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, timeOut, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, waitsfg, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[4], prometheus.CounterValue, timeWaitedfg, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[5], prometheus.CounterValue, timeOutfg, event, class)
	}
	return nil
}

func (c *sysEvent10GCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysEvent10GSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var event, class string
		var waits, timeWaited, timeOut float64
		if err := rows.Scan(&class, &event, &waits, &timeWaited, &timeOut); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeWaited, event, class)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, timeOut, event, class)
	}
	return nil
}

const (
	sysEvent11GCollectorNumber = 6
	sysEvent10GCollectorNumber = 3
	sysEvent11GSQL             = `
	SELECT n.wait_class,
			 	 e.event,
				 e.total_waits,
				 e.time_waited_micro,
				 e.total_timeouts,
				 e.total_waits_fg,
				 e.time_waited_micro_fg,
				 e.total_timeouts_fg
FROM v$system_event e, v$event_name n
WHERE n.name = e.event`
	sysEvent10GSQL = `
SELECT n.wait_class,
				e.event,
			 e.total_waits,
			 e.time_waited_micro,
			 e.total_timeouts
FROM v$system_event e, v$event_name n
WHERE n.name = e.event`
)
