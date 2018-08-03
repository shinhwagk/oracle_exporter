package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type logHistoryCollector struct {
	descs [3]*prometheus.Desc
}

func init() {
	registerCollector("logHistory-10g", NewLogHistoryCollector)
	registerCollector("logHistory-11g", NewLogHistoryCollector)
}

// NewLogHistoryCollector
func NewLogHistoryCollector() Collector {
	descs := [3]*prometheus.Desc{
		createNewDesc("loghistory", "count", "Gauge metric with count of sessions by status and type", nil, nil),
		createNewDesc("loghistory", "sequence_min", "Gauge metric with count of sessions by status and type", nil, nil),
		createNewDesc("loghistory", "sequence_max", "Gauge metric with count of sessions by status and type", nil, nil),
	}
	return &logHistoryCollector{descs}
}

func (c *logHistoryCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(logHistorySQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cnt, min, max float64

		if err = rows.Scan(&cnt, &min, &max); err != nil {
			return err
		}

		if cnt >= 1 {
			ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, cnt)
			ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, min)
			ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, max)
		}

	}
	return nil
}

const logHistorySQL = `
	SELECT COUNT(*), MIN(sequence#), MAX(sequence#) FROM v$log_history
 WHERE thread# = (SELECT instance_number FROM v$instance)
   AND first_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60`
