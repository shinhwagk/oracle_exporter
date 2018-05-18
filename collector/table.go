package collector

import (
	"database/sql"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	tableFlag = flag.Bool("collector.table", true, "for session activity collector")
)

type tableCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("table", defaultEnabled, NewTableCollector)
}

// NewTableCollector returns a new Collector exposing session activity statistics.
func NewTableCollector() (Collector, error) {
	return &tableCollector{
		newDesc("table", "bytes_total", "Generic counter metric from v$sysstat view in Oracle.", []string{"owner", "table_name"}, nil),
	}, nil
}

func (c *tableCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(`SELECT owner, segment_name, SUM(bytes) bytes
	FROM dba_segments
	WHERE owner NOT IN ('SYS', 'SYSTEM', 'WMSYS', 'DBSNMP', 'TSMSYS', 'SYSMAN', 'OLAPSYS', 'EXFSYS', 'CTXSYS', 'XDB')
		AND segment_type IN ('TABLE', 'TABLE PARTITION')
	GROUP BY owner, segment_name`)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var owner string
		var name string
		var bytes float64
		if err := rows.Scan(&owner, &name, &bytes); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, bytes, owner, name)
	}
	return nil
}
