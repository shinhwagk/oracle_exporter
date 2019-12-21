package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type tablespaceCollector struct {
	descs [3]*prometheus.Desc
}

func init() {
	registerCollector("tablespace-10g", NewTabalespaceCollector)
	registerCollector("tablespace-11g", NewTabalespaceCollector)
}

// NewTabalespaceCollector
func NewTabalespaceCollector() Collector {
	descs := [3]*prometheus.Desc{
		createNewDesc("tablespace", "alloc_bytes", "Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace"}, nil),
		createNewDesc("tablespace", "max_bytes", "Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace"}, nil),
		createNewDesc("tablespace", "alloc_free_bytes", "Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace"}, nil),
	}
	return &tablespaceCollector{descs}
}

func (c *tablespaceCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(tablespaceSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tablespaceName string
		var bytesFree, bytes, maxBytes float64

		if err := rows.Scan(&tablespaceName, &bytesFree, &bytes, &maxBytes); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.GaugeValue, bytes, tablespaceName)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.GaugeValue, maxBytes, tablespaceName)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.GaugeValue, bytesFree, tablespaceName)
	}
	return nil
}

const tablespaceSQL = `
SELECT ddf.tablespace_name, NVL(dfs.bytes, 0) free, ddf.bytes, ddf.maxbytes
  FROM (SELECT tablespace_name, SUM(DECODE(maxbytes, 0, bytes, maxbytes)) maxbytes, SUM(bytes) bytes
          FROM dba_data_files
         GROUP BY tablespace_name) ddf,
       dba_tablespaces dt,
       (SELECT tablespace_name, SUM(bytes) bytes FROM dba_free_space GROUP BY tablespace_name) dfs
 WHERE ddf.tablespace_name = dt.tablespace_name AND dt.tablespace_name = dfs.tablespace_name(+)`
