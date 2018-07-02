package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sysEventCollector struct {
	descs [2]*prometheus.Desc
}

type sysClassCollector struct {
	descs [4]*prometheus.Desc
}

func init() {
	registerCollector("systemEvent", cMin, defaultEnabled, NewSysEventCollector)
	registerCollector("systemClass", cMin, defaultEnabled, NewSysClassCollector)
}

// NewSysEventCollector returns a new Collector exposing session activity statistics.
func NewSysEventCollector() (Collector, error) {
	descs := [2]*prometheus.Desc{
		newDesc("sysevent", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		newDesc("sysevent", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
	}
	return &sysEventCollector{descs}, nil
}

func NewSysClassCollector() (Collector, error) {
	descs := [4]*prometheus.Desc{
		newDesc("sysclass", "waits_total", "Generic counter metric from v$system_class view in Oracle.", []string{"class"}, nil),
		newDesc("sysclass", "waited_time_total", "Generic counter metric from v$system_class view in Oracle.", []string{"class"}, nil),
		newDesc("sysclass", "waits_pg_total", "Generic counter metric from v$system_class view in Oracle.", []string{"class"}, nil),
		newDesc("sysclass", "waited_time_pg_total", "Generic counter metric from v$system_class view in Oracle.", []string{"class"}, nil),
	}
	return &sysClassCollector{descs}, nil
}

func (c *sysEventCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysEventSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var event, waitClass string
		var waits, time float64
		if err := rows.Scan(&waitClass, &event, &waits, &time); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, event, waitClass)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, time, event, waitClass)
	}
	return nil
}

func (c *sysClassCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sysClassSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var class string
		var waits, time, waitspg, timepg float64
		if err := rows.Scan(&class, &waits, &time, &waitspg, timepg); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, class)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, time, class)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, waitspg, class)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, timepg, class)
	}
	return nil
}

const (
	sysEventSQL = `
SELECT n.wait_class, e.event, e.total_waits, e.time_waited_micro
	FROM v$system_event e, v$event_name n
WHERE n.name = e.event AND time_waited > 0`
	sysClassSQL = `
select wait_class, TOTAL_WAITS, TIME_WAITED, TOTAL_WAITS_FG, TIME_WAITED_FG
  from v$system_wait_class`
)
