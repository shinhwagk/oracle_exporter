package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type ash11GCollector struct {
	desc *prometheus.Desc
}
type ash10GCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("ash-10g", NewASH10GCollector)
	registerCollector("ash-11g", NewASH11GCollector)
}

// NewASH11GCollector
func NewASH11GCollector() (Collector, error) {
	desc := createNewDesc("ash", "sample", "Gauge metric with count of sessions by status and type", []string{"sample_id", "session_id", "session_serial", "event", "session_type", "username", "sql_id", "opname", "program", "machine"}, nil)
	return &ash11GCollector{desc}, nil
}

// NewASH10GCollector
func NewASH10GCollector() (Collector, error) {
	desc := createNewDesc("ash", "sample", "Gauge metric with count of sessions by status and type", []string{"sample_id", "session_id", "session_serial", "event", "session_type", "username", "sql_id", "opcode", "program", "machine"}, nil)
	return &ash10GCollector{desc}, nil
}

func (c *ash11GCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(ashSQL11G)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sei, si, ssi, e, st, u, sli, so, p, m, bs string

		if err = rows.Scan(&sei, &si, &ssi, &e, &st, &u, &sli, &so, &p, &m, &bs); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(1), sei, si, ssi, e, st, u, sli, so, p, m)
	}
	return nil
}

func (c *ash10GCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(ashSQL10G)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sei, si, ssi, e, st, u, sli, so, p, m, bs string

		if err = rows.Scan(&sei, &si, &ssi, &e, &st, &u, &sli, &so, &p, &m, &bs); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(1), sei, si, ssi, e, st, u, sli, so, p, m)
	}
	return nil
}

const (
	ashSQL11G = `
SELECT sample_id,
       session_id,
       session_serial#,
       decode(session_state, 'ON CPU', 'Wait for CPU', 'WAITING', event),
       session_type,
       (SELECT username FROM dba_users WHERE user_id = ash.user_id),
       nvl(sql_id, 'null'),
       nvl(sql_opname, 'null'),
       nvl(program, 'null'),
       nvl(machine, 'null'),
       decode(blocking_session, null, 'null', to_char(blocking_session))
  FROM v$active_session_history ash
 WHERE SAMPLE_TIME >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND SAMPLE_TIME < TRUNC(sysdate, 'MI')`
	ashSQL10G = `
SELECT sample_id,
       session_id,
       session_serial#,
       decode(session_state, 'ON CPU', 'Wait for CPU', 'WAITING', event),
       session_type,
       (SELECT username FROM dba_users WHERE user_id = ash.user_id),
       nvl(sql_id, 'null'),
       sql_opcode,
       nvl(program, 'null'),
       nvl(machine, 'null'),
       decode(blocking_session, null, 'null', to_char(blocking_session))
  FROM v$active_session_history ash
 WHERE SAMPLE_TIME >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND SAMPLE_TIME < TRUNC(sysdate, 'MI')`
)
