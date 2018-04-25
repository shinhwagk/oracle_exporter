package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type upCollector struct {
	up *prometheus.Desc
}

func init() {
	registerCollector("up", defaultEnabled, NewUpCollector)
}

// NewUpCollector returns a new Collector exposing up statistics.
func NewUpCollector() (Collector, error) {
	return &upCollector{prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Whether the Oracle database server is up.", []string{}, nil)}, nil
}

func (c *upCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		rows *sql.Rows
		err  error
	)
	rows, err = db.Query("SELECT status, type, COUNT(*) FROM v$session GROUP BY status, type")
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			status, sessionType string
			count               float64
		)
		if err := rows.Scan(&status, &sessionType, &count); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.up,
			prometheus.GaugeValue,
			count,
			status,
			sessionType,
		)
	}
	return nil
}
