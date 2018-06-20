package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sqlCollector struct {
	descs [5]*prometheus.Desc
}

func init() {
	registerCollector("sql", cMin, defaultEnabled, NewSQLCollector)
}

// NewSQLCollector returns a new Collector exposing session activity statistics.
func NewSQLCollector() (Collector, error) {
	descs := [5]*prometheus.Desc{
		newDesc("sql", "cpu_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command"}, nil),
		newDesc("sql", "elapsed_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command"}, nil),
		newDesc("sql", "executions_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command"}, nil),
		newDesc("sql", "buffer_gets_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command"}, nil),
		newDesc("sql", "disk_read_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command"}, nil),
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
		var sqlID, username string
		var cpuTime, elapsedTime, executions, bufferGets, diskReads float64
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
			 DISK_READS,
			 (select command_name from v$sqlcommand where s.command_type = command_type)
			 FROM v$sqlarea s
			 WHERE last_active_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60`
