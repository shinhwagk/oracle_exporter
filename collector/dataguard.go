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
		createNewDesc("lag", "second", "dataguard lag", []string{"name"}, nil),
	}
}

func (c *dataguardCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessionSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name   string
			second float64
		)
		if err = rows.Scan(&name, &second); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, second, name)
	}
	return nil
}

const dataguardSQL = `
SELECT name,
       EXTRACT(DAY FROM itval) * 24 * 60 * 60 +
	   EXTRACT(HOUR FROM itval) * 60 * 60 +
	   EXTRACT(MINUTE FROM itval) * 60 +
       EXTRACT(SECOND FROM itval)
  FROM (SELECT name, TO_DSINTERVAL(value) itval
          FROM v$dataguard_stats
         WHERE name IN ('apply lag', 'transport lag'))`
