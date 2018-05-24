package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessEventCollector struct {
	descs [2]*prometheus.Desc
}

func init() {
	registerCollector("sessionevent", cMin, defaultEnabled, NewSessEventCollector)
}

// NewSessEventCollector returns a new Collector exposing session activity statistics.
func NewSessEventCollector() (Collector, error) {
	descs := [2]*prometheus.Desc{
		newDesc("sessevent", "waits_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
		newDesc("sessevent", "waited_time_total", "Generic counter metric from v$system_event view in Oracle.", []string{"event", "class"}, nil),
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
		var name, waitClass string
		var waits, timeWaited float64
		if err := rows.Scan(&name, &waits, &timeWaited, &waitClass); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, waits, name, waitClass)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, timeWaited, name, waitClass)
	}
	return nil
}

const sessEventSQL = `SELECT event, total_waits, time_waited_micro, wait_class FROM v$session_event`
