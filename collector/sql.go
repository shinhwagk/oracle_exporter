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
		newDesc("sql", "cpu_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id"}, nil),
		newDesc("sql", "elapsed_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id"}, nil),
		newDesc("sql", "executions_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id"}, nil),
		newDesc("sql", "buffer_gets_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id"}, nil),
		newDesc("sql", "disk_read_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id"}, nil),
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
		var bufferGets float64
		var username string
		var diskReads float64
		if err := rows.Scan(&sqlID, &cpuTime, &elapsedTime, &executions, &bufferGets, &username, &diskReads); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, cpuTime, username, sqlID)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, elapsedTime, username, sqlID)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, executions, username, sqlID)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, bufferGets, username, sqlID)
		ch <- prometheus.MustNewConstMetric(c.descs[4], prometheus.CounterValue, diskReads, username, sqlID)
	}
	return nil
}

const sqlSQL = `
select sql_id,
       cpu_time,
       elapsed_time,
       executions,
       buffer_gets,
       PARSING_SCHEMA_NAME,
       DISK_READS
  from v$sqlarea
 where last_active_time >= trunc(sysdate, 'MI') - 1 / 24 / 60
   and PARSING_SCHEMA_ID in
       (select ASH.user_id
          from V$ACTIVE_SESSION_HISTORY ASH, DBA_USERS du
         where ASH.sample_time >= trunc(sysdate, 'MI') - 1 / 24 / 60
           AND ash.user_id = du.user_id
           and du.username not in ('SYS', 'SYSTEM')
         GROUP BY ASH.USER_ID)
`
