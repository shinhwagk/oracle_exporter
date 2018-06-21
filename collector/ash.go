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
		newDesc("ash", "wait_class", "Gauge metric with count of sessions by status and type", []string{"class", "sql_id", "inst_id", "username", "event"}, nil),
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
		var sqlID, instID, username, event string
		var cpu, bcpu, scheduler, userio, systemio, concurrency, application, commit, configuration, administrative, network, queueing, cluster, other float64

		if err = rows.Scan(&instID, &sqlID, &event, &username, &cpu, &bcpu, &scheduler, &userio, &systemio, &concurrency, &application, &commit, &configuration, &administrative, &network, &queueing, &cluster, &other); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, cpu, "Cpu", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, bcpu, "Bcpu", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, scheduler, "Scheduler", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, userio, "User I/O", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, systemio, "System I/O", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, concurrency, "Concurrency", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, application, "Application", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, commit, "Commit", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, configuration, "Configuration", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, administrative, "Administrative", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, network, "Network", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, queueing, "Queueing", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, cluster, "Cluster", sqlID, instID, username, event)
		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, other, "Other", sqlID, instID, username, event)
	}
	return nil
}

const ashSQL = `
SELECT 
	inst_id,
	sql_id,
	event,
	(select username from dba_users where user_id = ash.user_id),
	SUM(DECODE(session_state, 'ON CPU',	DECODE(session_type, 'BACKGROUND', 0, 1),	0)),
	SUM(DECODE(session_state, 'ON CPU', DECODE(session_type, 'BACKGROUND', 1, 0), 0)),
	SUM(DECODE(wait_class, 'Scheduler', 1, 0)),
	SUM(DECODE(wait_class, 'User I/O', 1, 0)),
	SUM(DECODE(wait_class, 'System I/O', 1, 0)),
	SUM(DECODE(wait_class, 'Concurrency', 1, 0)),
	SUM(DECODE(wait_class, 'Application', 1, 0)),
	SUM(DECODE(wait_class, 'Commit', 1, 0)),
	SUM(DECODE(wait_class, 'Configuration', 1, 0)),
	SUM(DECODE(wait_class, 'Administrative', 1, 0)),
	SUM(DECODE(wait_class, 'Network', 1, 0)),
	SUM(DECODE(wait_class, 'Queueing', 1, 0)),
	SUM(DECODE(wait_class, 'Cluster', 1, 0)),
	SUM(DECODE(wait_class, 'Other', 1, 0))
FROM gv$active_session_history ash
WHERE SAMPLE_TIME >= TRUNC(sysdate, 'MI') - 1 / 24 AND SAMPLE_TIME < TRUNC(sysdate, 'MI')
group by sql_id, user_id, inst_id, event`
