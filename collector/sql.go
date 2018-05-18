package collector

import (
	"database/sql"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	sqlFlag = flag.Bool("collector.sql", true, "for session activity collector")
)

type sqlCollector struct {
	descs [5]*prometheus.Desc
}

func init() {
	registerCollector("sql", defaultEnabled, NewSQLCollector)
}

// NewSQLCollector returns a new Collector exposing session activity statistics.
func NewSQLCollector() (Collector, error) {
	descs := [5]*prometheus.Desc{
		newDesc("sql", "cpu_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "sample_time"}, nil),
		newDesc("sql", "elapsed_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "sample_time"}, nil),
		newDesc("sql", "executions_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "sample_time"}, nil),
		newDesc("sql", "buffer_gets_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "sample_time"}, nil),
		newDesc("sql", "disk_read_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "sample_time"}, nil),
	}
	return &sqlCollector{descs}, nil
}

func (c *sqlCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sqlSQL)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sqlID string
		var cpuTime float64
		var elapsedTime float64
		var executions float64
		var sampleTime string
		var bufferGets float64
		var username string
		var diskReads float64
		if err := rows.Scan(&sqlID, &cpuTime, &elapsedTime, &executions, &sampleTime, &bufferGets, &username, &diskReads); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, cpuTime, username, sqlID, sampleTime)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, elapsedTime, username, sqlID, sampleTime)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, executions, username, sqlID, sampleTime)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, bufferGets, username, sqlID, sampleTime)
		ch <- prometheus.MustNewConstMetric(c.descs[4], prometheus.CounterValue, diskReads, username, sqlID, sampleTime)
	}
	return nil
}

const sqlSQL = `
with vash as
 (select sql_id,
         sample_time,
         user_id,
         row_number() over(PARTITION BY sql_id order by sample_time desc) rn
    from V$ACTIVE_SESSION_HISTORY
   where sample_time >= trunc(sysdate, 'MI') - 1 / 24 / 60)
SELECT *
  FROM (select s.sql_id,
               s.CPU_TIME,
               s.ELAPSED_TIME,
               s.EXECUTIONS,
               vash.sample_time,
               S.BUFFER_GETS,
               s.LAST_ACTIVE_TIME,
               du.username,
               S.DISK_READS
          from v$sqlstats s, vash vash, dba_users du
         where vash.sql_id = s.sql_id
           and du.user_id = vash.user_id
           and vash.rn = 1
           and vash.sql_id is not null)
`
