package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type segmentSizeCollector struct {
	descs *prometheus.Desc
}

func init() {
	registerCollector("segmentSize-10g", NewSegmentSizeCollector)
	registerCollector("segmentSize-11g", NewSegmentSizeCollector)
}

// NewSegmentSizeCollector
func NewSegmentSizeCollector() Collector {
	descs :=   createNewDesc("segmentSize", "bytes", "Generic counter metric of segmentSizes bytes in Oracle.", []string{"ownerm","name","type","tablespace"}, nil)

	return &segmentSizeCollector{descs}
}

func (c *segmentSizeCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(segmentSizeSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var owner,segmentName,segmentType,tablespaceName string
		var bytes  float64

		if err := rows.Scan(&owner,&segmentName,&segmentType,&tablespaceName,&bytes); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs, prometheus.GaugeValue, bytes, tablespaceName)
	}
	return nil
}

const segmentSizeSQL = `
SELECT owner, segment_name, segment_type, tablespace_name, sum(bytes)
  FROM dba_segments
 WHERE tablespace_name NOT IN ('SYSTEM','SYSAUX')
 GROUP BY owner, segment_name, segment_type, tablespace_name`
