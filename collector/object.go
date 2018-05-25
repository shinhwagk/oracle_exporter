package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type objectCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("object", cHour, defaultEnabled, NewObjectCollector)
}

// NewObjectCollector returns a new Collector exposing session activity statistics.
func NewObjectCollector() (Collector, error) {
	desc := newDesc("object", "total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "object_type"}, nil)
	return &objectCollector{desc}, nil
}

func (c *objectCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(objectSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var owner, object_type string
		var count float64
		if err := rows.Scan(&owner, &object_type, &count); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, count, owner, object_type)
	}
	return nil
}

const objectSQL = `
select owner, object_type, count(*)
  from dba_objects
 WHERE owner NOT IN ('SYS',
                     'SYSTEM',
                     'WMSYS',
                     'DBSNMP',
                     'TSMSYS',
                     'SYSMAN',
                     'OLAPSYS',
                     'EXFSYS',
                     'CTXSYS',
										 'XDB',
										 'DMSYS')
 group by owner, object_type`
