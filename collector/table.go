package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type tableCollector struct {
	desc *prometheus.Desc
}

func init() {
	// registerCollector("table", cHour, defaultEnabled, NewTableCollector)
}

// NewTableCollector returns a new Collector exposing session activity statistics.
func NewTableCollector() (Collector, error) {
	return &tableCollector{
		newDesc("table", "bytes_total", "Generic counter metric from v$sysstat view in Oracle.", []string{"owner", "table_name"}, nil),
	}, nil
}

func (c *tableCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(tableSQL)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var owner, name string
		var value float64
		if err := rows.Scan(&owner, &name, &value); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value, owner, name)
	}
	return nil
}

const tableSQL = `
SELECT owner, segment_name, SUM(bytes)
  FROM DBA_EXTENTS
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
                     'MDSYS')
   AND segment_type IN ('TABLE', 'TABLE PARTITION')
 GROUP BY owner, segment_name, SEGMENT_TYPE`
