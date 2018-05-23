package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type eventCollector struct {
	descs [2]*prometheus.Desc
}

func init() {
	registerCollector("event", defaultEnabled, NewEventCollector)
}

// NewEventCollector returns a new Collector exposing session activity statistics.
func NewEventCollector() (Collector, error) {
	descs := [2]*prometheus.Desc{
		newDesc("event", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		newDesc("event", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
	}
	return &eventCollector{descs}, nil
}

func (c *eventCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(eventSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, wait_class string
		var waits, time_waited float64
		if err := rows.Scan(&name, &waits, &time_waited, &wait_class); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, name, wait_class)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, time_waited, name, wait_class)
	}
	return nil
}

const eventSQL = `SELECT event, total_waits, time_waited_micro, wait_class FROM v$system_event`