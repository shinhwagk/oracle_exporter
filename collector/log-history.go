package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type logHistoryCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("loghistory-10g", NewLogHistoryCollector)
	registerCollector("loghistory-11g", NewLogHistoryCollector)
}

// NewLogHistoryCollector returns a new Collector exposing ash activity statistics.
func NewLogHistoryCollector() (Collector, error) {
	desc := newDesc("loghistory", "sequence", "Gauge metric with count of sessions by status and type", []string{"recid", "year", "month", "day", "hour", "minute"}, nil)
	return &logHistoryCollector{desc}, nil
}

func (c *logHistoryCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(logHistorySQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var r, y, m, d, h, mi string
		var seq float64

		if err = rows.Scan(&r, &y, &m, &d, &h, &mi, &seq); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, seq, r, y, m, d, h, mi)
	}
	return nil
}

const logHistorySQL = `
	SELECT recid,
       TO_CHAR(first_time, 'YYYY'),
       TO_CHAR(first_time, 'MM'),
       TO_CHAR(first_time, 'DD'),
       TO_CHAR(first_time, 'HH24'),
			 TO_CHAR(first_time, 'MI'),
			 sequence#
  FROM v$log_history
 WHERE thread# = (SELECT instance_number FROM v$instance)
   AND first_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60`
