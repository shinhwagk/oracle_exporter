package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type dataguardCollector struct {
	desc *prometheus.Desc
}

func init() {
	registerCollector("dataguard-10g", NewDadataguardCollector)
	registerCollector("dataguard-11g", NewDadataguardCollector)
}

// NewDadataguardCollector
func NewDadataguardCollector() Collector {
	return &dataguardCollector{
		createNewDesc("lag", "second", "dataguard lag", []string{"dgtype", "name"}, nil),
	}
}

func (c *dataguardCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(dataguardSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			dgtype, name string
			second       float64
		)
		if err = rows.Scan(&dgtype, &name, &second); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, second, dgtype, name)
	}
	return nil
}

const dataguardSQL = `
SELECT 'physical', name,
       EXTRACT(DAY FROM itval) * 24 * 60 * 60 +
	   EXTRACT(HOUR FROM itval) * 60 * 60 + 
	   EXTRACT(MINUTE FROM itval) * 60 +
       EXTRACT(SECOND FROM itval)
  FROM (SELECT ds.name, TO_DSINTERVAL(value) itval
          FROM v$dataguard_stats ds, v$database d
         WHERE d.database_role = 'PHYSICAL STANDBY'
           AND ds.name IN ('apply lag', 'transport lag'))
UNION ALL
SELECT 'logical', 'apply lag', (SYSDATE - lp.applied_time) * 24 * 60 * 60
  FROM v$logstdby_progress lp, v$database d
 WHERE d.database_role = 'LOGICAL STANDBY'`
