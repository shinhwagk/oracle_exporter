package collector

import (
	"database/sql"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	sessionSQL  *string
	sessionFlag = flag.Bool("collector.session", true, "for session activity collector")
)

type sessionCollector struct {
	activity *prometheus.Desc
}

func init() {
	s, err := readFile("session.sql")
	sessionSQL = s
	if err != nil {
		log.Errorln("Error opening sql file session.sql:", err)
	} else {
		registerCollector("session", defaultEnabled, NewSessionCollector)
	}
}

// NewSessionCollector returns a new Collector exposing session activity statistics.
func NewSessionCollector() (Collector, error) {
	return &sessionCollector{
		newDesc("sessions", "activity", "Gauge metric with count of sessions by status and type", []string{"status", "type"}, nil),
	}, nil
}

func (c *sessionCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(*sessionSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			status, sessionType string
			count               float64
		)
		if err = rows.Scan(&status, &sessionType, &count); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.activity, prometheus.GaugeValue, count, status, sessionType)
	}
	return nil
}
