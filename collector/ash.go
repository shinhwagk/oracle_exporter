package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type ashCollector struct {
	descs [1]*prometheus.Desc
}

func init() {
	registerCollector("ash", cMin, defaultEnabled, NewASHCollector)
}

// NewASHCollector returns a new Collector exposing ash activity statistics.
func NewASHCollector() (Collector, error) {
	descs := [1]*prometheus.Desc{
		newDesc("ash", "wait_class", "Gauge metric with count of sessions by status and type", []string{"class"}, nil),
	}
	return &ashCollector{descs}, nil
}

func (c *ashCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(ashSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cpu, bcpu, scheduler, userio, systemio, concurrency, application, commit, configuration, administrative, network, queueing, cluster, other float64

		if err = rows.Scan(&cpu, &bcpu, &scheduler, &userio, &systemio, &concurrency, &application, &commit, &configuration, &administrative, &network, &queueing, &cluster, &other); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, cpu, "Cpu")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, bcpu, "Bcpu")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, scheduler, "Scheduler")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, userio, "User I/O")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, systemio, "System I/O")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, concurrency, "Concurrency")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, application, "Application")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, commit, "Commit")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, configuration, "Configuration")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, administrative, "Administrative")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, network, "Network")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, queueing, "Queueing")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, cluster, "Cluster")
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, other, "Other")
	}
	return nil
}

const ashSQL = `
SELECT 
	DECODE(session_state, 'ON CPU',	DECODE(session_type, 'BACKGROUND', 0, 1),	0),
	DECODE(session_state, 'ON CPU', DECODE(session_type, 'BACKGROUND', 1, 0), 0),
	DECODE(wait_class, 'Scheduler', 1, 0),
	DECODE(wait_class, 'User I/O', 1, 0),
	DECODE(wait_class, 'System I/O', 1, 0),
	DECODE(wait_class, 'Concurrency', 1, 0),
	DECODE(wait_class, 'Application', 1, 0),
	DECODE(wait_class, 'Commit', 1, 0),
	DECODE(wait_class, 'Configuration', 1, 0),
	DECODE(wait_class, 'Administrative', 1, 0),
	DECODE(wait_class, 'Network', 1, 0),
	DECODE(wait_class, 'Queueing', 1, 0),
	DECODE(wait_class, 'Cluster', 1, 0),
	DECODE(wait_class, 'Other', 1, 0)
FROM v$active_session_history
WHERE SAMPLE_TIME >= TRUNC(sysdate, 'MI') - 1 / 24 AND SAMPLE_TIME < TRUNC(sysdate, 'MI')`
