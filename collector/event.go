package collector

import (
	"database/sql"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

type eventCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("event", defaultEnabled, NewEventCollector)
}

// NewEventCollector returns a new Collector exposing session activity statistics.
func NewEventCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["total_waits"] = newDesc("event", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event"}, nil)
	descs["time_waited_micro"] = newDesc("event", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event"}, nil)
	return &eventCollector{descs}, nil
}

func (c *eventCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(eventSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]
		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, name)
		} else {
			return errors.New("event desc no exist")
		}
	}
	return nil
}

const eventSQL = `SELECT event, total_waits, time_waited_micro FROM v$system_event`
