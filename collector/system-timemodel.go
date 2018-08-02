package collector

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type sysTimeModelCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("systemTimeModel-10g", NewSysTimeModelCollector)
	registerCollector("systemTimeModel-11g", NewSysTimeModelCollector)
}

// NewSysTimeModelCollector
func NewSysTimeModelCollector() Collector {
	descs := make(map[string]*prometheus.Desc)
	descs["DB time"] = createNewDesc(sysTimeModelSystemName, "db_time", "Generic counter metric.", nil, nil)
	return &sysTimeModelCollector{descs}
}

func (c *sysTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(systemTimeModelSQL)
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
			err := fmt.Sprintf("system time model: %s no exist", name)
			return errors.New(err)
		}

	}
	return nil
}

const (
	sysTimeModelSystemName = "systimemodel"
	systemTimeModelSQL     = `SELECT stat_name, value FROM v$sys_time_model WHERE stat_name in ('DB time')`
)
