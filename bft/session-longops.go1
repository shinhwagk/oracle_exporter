package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sessionLongOpsCollector struct {
	desc *prometheus.Desc
}

func init() {
	// registerCollector("sessionLongOps", NewSessionLongOpsCollector)
}

// NewSessionLongOpsCollector returns a new Collector exposing session activity statistics.
func NewSessionLongOpsCollector() (Collector, error) {
	return &sessionLongOpsCollector{
		newDesc("session", "longops", "Gauge metric with count of sessions by status and type", []string{"username", "opname", "target", "sid", "serial"}, nil),
	}, nil
}

func (c *sessionLongOpsCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessionLongOpsSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var username, opname, target, sid, serial string

		if err = rows.Scan(&username, &opname, &target, &sid, &serial); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(1), username, opname, target, sid, serial)
	}
	return nil
}

const sessionLongOpsSQL = `
select username, opname, target, sid, serial, count(*)
  from (SELECT username,
               opname,
               nvl(target, '') || nvl(target_desc, '') target,
               sid,
               serial# serial
          FROM v$session_longops
         WHERE (start_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND
               start_time < TRUNC(sysdate, 'MI'))
            or (last_update_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND
               last_update_time < TRUNC(sysdate, 'MI')))
 group by username, opname, target, sid, serial`
