package collector

import (
	"database/sql"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	tablespaceSQL  *string
	tablespaceFlag = flag.Bool("collector.tablespace", true, "for tablespace space collector")
)

type tablespaceCollector struct {
	tablespaceBytesDesc     *prometheus.Desc
	tablespaceMaxBytesDesc  *prometheus.Desc
	tablespaceFreeBytesDesc *prometheus.Desc
}

func init() {
	s, err := readFile("tablespace.sql")
	tablespaceSQL = s
	if err != nil {
		log.Errorln("Error opening sql file tablespace.sql:", err)
	} else {
		registerCollector("tablespace", defaultEnabled, NewTabalespaceCollector)
	}
}

// NewTabalespaceCollector returns a new Collector exposing session activity statistics.
func NewTabalespaceCollector() (Collector, error) {
	return &tablespaceCollector{
		prometheus.NewDesc(prometheus.BuildFQName(namespace, "tablespace", "bytes"),
			"Generic counter metric of tablespaces bytes in Oracle.", []string{"tablespace", "type"}, nil),
		prometheus.NewDesc(prometheus.BuildFQName(namespace, "tablespace", "max_bytes"),
			"Generic counter metric of tablespaces max bytes in Oracle.", []string{"tablespace", "type"}, nil),
		prometheus.NewDesc(prometheus.BuildFQName(namespace, "tablespace", "free"),
			"Generic counter metric of tablespaces free bytes in Oracle.", []string{"tablespace", "type"}, nil),
	}, nil
}

func (c *tablespaceCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(*tablespaceSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tablespaceName, status, contents, extentManagement string
		var bytes, maxBytes, bytesFree float64

		if err := rows.Scan(&tablespaceName, &status, &contents, &extentManagement, &bytes, &maxBytes, &bytesFree); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.tablespaceBytesDesc, prometheus.GaugeValue, float64(bytes), tablespaceName, contents)
		ch <- prometheus.MustNewConstMetric(c.tablespaceMaxBytesDesc, prometheus.GaugeValue, float64(maxBytes), tablespaceName, contents)
		ch <- prometheus.MustNewConstMetric(c.tablespaceFreeBytesDesc, prometheus.GaugeValue, float64(bytesFree), tablespaceName, contents)
	}
	return nil
}
