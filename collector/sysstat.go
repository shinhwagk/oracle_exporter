package collector

import (
	"database/sql"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	sysstatSQL *string
	// sysstatFlag = flag.Bool("collector.sysstat", true, "for session activity collector")
)

type sysstatCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	s, err := readFile("sysstat.sql")
	sysstatSQL = s
	if err != nil {
		log.Errorln("Error opening sql file sysstat.sql:", err)
	} else {
		registerCollector("sysstat", defaultEnabled, NewSessionCollector)
	}
}

// NewSysstatCollector returns a new Collector exposing session activity statistics.
func NewSysstatCollector() (Collector, error) {
	descs := make(map[string]*prometheus.Desc)
	descs["commit_total"] = newDesc("sysstat", "commit_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["rollback_total"] = newDesc("sysstat", "rollback_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["execute_total"] = newDesc("sysstat", "execute_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	descs["parse_total"] = newDesc("sysstat", "parse_total", "Generic counter metric from v$sysstat view in Oracle.", nil, nil)
	return &sysstatCollector{descs}, nil
}

func (c *sysstatCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(*sysstatSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}

		desc, ok := c.descs[name]
		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value)
		} else {
			return errors.New("sysstat desc no exist")
		}
	}
	return nil
}
