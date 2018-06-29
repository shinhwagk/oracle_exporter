package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sqlCollector struct {
	descs [6]*prometheus.Desc
}

func init() {
	registerCollector("sql", cMin, defaultEnabled, NewSQLCollector)
}

// NewSQLCollector returns a new Collector exposing session activity statistics.
func NewSQLCollector() (Collector, error) {
	descs := [6]*prometheus.Desc{
		newDesc("sql", "cpu_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		newDesc("sql", "elapsed_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		newDesc("sql", "executions_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		newDesc("sql", "buffer_gets_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		newDesc("sql", "disk_read_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		newDesc("sql", "sort_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		// newDesc("sql", "phys_read_req_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		// newDesc("sql", "phys_read_bytes_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		// newDesc("sql", "phys_write_req_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		// newDesc("sql", "phys_write_bytes_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
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
		var sqlID, username, commandType, child string
		var cpuTime, elapsedTime, executions, bufferGets, diskReads, sort float64
		if err := rows.Scan(&sqlID, &child, &commandType, &username, &cpuTime, &elapsedTime, &bufferGets, &diskReads, &sort, &executions); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, cpuTime, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, elapsedTime, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, executions, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, bufferGets, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[4], prometheus.CounterValue, diskReads, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[5], prometheus.CounterValue, sort, username, sqlID, commandType, child)
		// ch <- prometheus.MustNewConstMetric(c.descs[6], prometheus.CounterValue, prr, username, sqlID, commandType, child)
		// ch <- prometheus.MustNewConstMetric(c.descs[7], prometheus.CounterValue, prb, username, sqlID, commandType, child)
		// ch <- prometheus.MustNewConstMetric(c.descs[8], prometheus.CounterValue, pwr, username, sqlID, commandType, child)
		// ch <- prometheus.MustNewConstMetric(c.descs[9], prometheus.CounterValue, pwb, username, sqlID, commandType, child)
	}
	return nil
}

const sqlSQL = `
select SQL_ID,
			 CHILD_NUMBER,
			 (select command_name from v$sqlcommand where s.command_type = command_type),
			 PARSING_SCHEMA_NAME,
       CPU_TIME,
			 ELAPSED_TIME,
			 BUFFER_GETS,
       DISK_READS,
       SORTS,
       EXECUTIONS
  FROM v$sql s
 WHERE last_active_time >= TRUNC(sysdate, 'MI') - 1 / 24 / 60 AND is_obsolete ='N' AND last_active_time < TRUNC(sysdate, 'MI')`
