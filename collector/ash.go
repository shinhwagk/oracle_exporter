package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type ashCollector struct {
	descs [2]*prometheus.Desc
}

func init() {
	registerCollector("ash", cMin, defaultEnabled, NewASHCollector)
}

// NewASHCollector returns a new Collector exposing ash activity statistics.
func NewASHCollector() (Collector, error) {
	descs := [2]*prometheus.Desc{
		newDesc("ash", "waitting", "Gauge metric with count of sessions by status and type", []string{"class", "sql_id", "username", "event", "opname"}, nil),
		newDesc("ash", "on_cpu", "Gauge metric with count of sessions by status and type", []string{"sql_id", "username", "opname", "type"}, nil),
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
		var ss, sqlID, username, event, opname, sessionType string
		var cpu, bcpu, scheduler, userio, systemio, concurrency, application, commit, configuration, administrative, network, queueing, cluster, other float64

		if err = rows.Scan(&ss, &sqlID, &event, &opname, &username, &sessionType, &cpu, &bcpu, &scheduler, &userio, &systemio, &concurrency, &application, &commit, &configuration, &administrative, &network, &queueing, &cluster, &other); err != nil {
			return err
		}
		if ss == "WATTING" {
			// ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, cpu, "Cpu", sqlID, username, event, opname)
			// ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, bcpu, "Bcpu", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, scheduler, "Scheduler", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, userio, "User I/O", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, systemio, "System I/O", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, concurrency, "Concurrency", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, application, "Application", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, commit, "Commit", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, configuration, "Configuration", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, administrative, "Administrative", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, network, "Network", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, queueing, "Queueing", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, cluster, "Cluster", sqlID, username, event, opname)
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, other, "Other", sqlID, username, event, opname)
		}
	}
	return nil
}

const ashSQL = `
SELECT
	session_state,
	nvl(sql_id,'null'),
	nvl(event,'null'),
	SQL_OPNAME,
	SESSION_TYPE,
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
group by nvl(sql_id, 'null'), user_id, nvl(event,'null'), SQL_OPNAME, session_state, SESSION_TYPE`
