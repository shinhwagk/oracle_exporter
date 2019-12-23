package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type segmentCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("segment-10g", NewSegmentCollector)
	registerCollector("segment-11g", NewSegmentCollector)
}

// NewSegmentCollector
func NewSegmentCollector() Collector {
	descs := createNewDesc("segment", "bytes", "collect segment size", []string{"owner", "name", "type", "tablespace"}, nil)
	return &segmentCollector{descs}
}

func (c *segmentCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(segmentSizeSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var owner, sName, sType, tsName string
		var bytes float64

		if err := rows.Scan(&owner, &sName, &sType, &tsName, &bytes); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, bytes, owner, sName, sType, tsName)
	}
	return nil
}

const segmentSizeSQL = `
SELECT owner, segment_name, segment_type, tablespace_name, sum(bytes)
  FROM dba_segments
 WHERE tablespace_name NOT IN ('SYSTEM','SYSAUX') AND tablespace_name not like 'UNDOTBS%'
 GROUP BY owner, segment_name, segment_type, tablespace_name`
