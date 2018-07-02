package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type datafileCollector struct {
	descs [4]*prometheus.Desc
}

func init() {
	// registerCollector("datafile", cMin, defaultEnabled, NewDatafileCollector)
}

// NewDatafileCollector returns a new Collector exposing session activity statistics.
func NewDatafileCollector() (Collector, error) {
	descs := [4]*prometheus.Desc{
		newDesc("datafile", "small_read_megabytes_total", "Generic counter metric from v$iostat_file view in Oracle.", []string{"tablespace"}, nil),
		newDesc("datafile", "small_write_megabytes_total", "Generic counter metric from v$iostat_file view in Oracle.", []string{"tablespace"}, nil),
		newDesc("datafile", "large_read_megabytes_total", "Generic counter metric from v$iostat_file view in Oracle.", []string{"tablespace"}, nil),
		newDesc("datafile", "large_write_megabytes_total", "Generic counter metric from v$iostat_file view in Oracle.", []string{"tablespace"}, nil),
	}
	return &datafileCollector{descs}, nil
}

func (c *datafileCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(datafileSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tablespace string
		var sRMB, sWMB, lRMB, lWMB float64
		if err := rows.Scan(&tablespace, &sRMB, &sWMB, &lRMB, &lWMB); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, sRMB, tablespace)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, sWMB, tablespace)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, lRMB, tablespace)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, lWMB, tablespace)
	}
	return nil
}

const datafileSQL = `
SELECT ddf.tablespace_name,
       SUM(SMALL_READ_MEGABYTES) SMALL_READ_MEGABYTES,
       SUM(SMALL_WRITE_MEGABYTES) SMALL_WRITE_MEGABYTES,
       SUM(LARGE_READ_MEGABYTES) LARGE_READ_MEGABYTES,
       SUM(LARGE_WRITE_MEGABYTES) LARGE_WRITE_MEGABYTES
  FROM dba_data_files ddf, v$IOSTAT_FILE vif
 WHERE ddf.file_id = vif.FILE_NO
 GROUP BY ddf.tablespace_name`
