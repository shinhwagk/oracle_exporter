package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sysEventCollector struct {
	descs [2]*prometheus.Desc
}

func init() {
	registerCollector("sysevent", cMin, defaultEnabled, NewSysEventCollector)
}

// NewSysEventCollector returns a new Collector exposing session activity statistics.
func NewSysEventCollector() (Collector, error) {
	descs := [2]*prometheus.Desc{
		newDesc("sysevent", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		newDesc("sysevent", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
	}
	return &sysEventCollector{descs}, nil
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

const sysEventSQL = `
SELECT n.wait_class, e.event, e.total_waits, e.time_waited_micro
  FROM v$system_event e, v$event_name n
 WHERE n.name = e.event
   AND time_waited > 0`
