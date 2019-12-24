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
func NewASH11GCollector() Collector {
	desc := createNewDesc("ash", "sample", "Gauge metric with count of sessions by status and type", []string{"sample_id", "sid", "serial", "event", "type", "username", "sql_id", "opname", "program", "machine", "blocking"}, nil)
	return &ash11GCollector{desc}
}

// NewASH10GCollector
func NewASH10GCollector() Collector {
	desc := createNewDesc("ash", "sample", "Gauge metric with count of sessions by status and type", []string{"sample_id", "sid", "serial", "event", "type", "username", "sql_id", "opname", "program", "machine", "blocking"}, nil)
	return &ash10GCollector{desc}
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
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(1), sei, si, ssi, e, st, u, sli, so, p, m, bs)
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
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(1), sei, si, ssi, e, st, u, sli, so, p, m, bs)
	}
	return nil
}

const (
	ashSQL11G = `
SELECT sample_id,
       session_id,
       session_serial#,
       DECODE(session_state, 'ON CPU', 'Wait for CPU', 'WAITING', event),
       session_type,
       (SELECT username FROM dba_users WHERE user_id = ash.user_id),
       NVL(sql_id, 'null'),
       NVL(sql_opname, 'null'),
       NVL(program, 'null'),
	   NVL(machine, 'null'),
	   TO_CHAR(NVL(blocking_session, 0))
  FROM v$active_session_history ash
 WHERE sample_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND sample_time < TRUNC(sysdate, 'MI')`
	ashSQL10G = `
SELECT sample_id,
       session_id,
       session_serial#,
       DECODE(session_state, 'ON CPU', 'Wait for CPU', 'WAITING', event),
       session_type,
       (SELECT username FROM dba_users WHERE user_id = ash.user_id),
       NVL(sql_id, 'null'),
       (SELECT name FROM audit_actions WHERE ash.sql_opcode = action),
       NVL(program, 'null'),
       NVL(machine, 'null'),
       TO_CHAR(NVL(blocking_session, 0))
  FROM v$active_session_history ash
 WHERE sample_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND sample_time < TRUNC(sysdate, 'MI')`
)
