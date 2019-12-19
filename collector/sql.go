package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type sql11GCollector struct {
	descs [18]*prometheus.Desc
}

type sql10GCollector struct {
	descs [14]*prometheus.Desc
}

func init() {
	registerCollector("sql-11g", NewSQL11GCollector)
	registerCollector("sql-10g", NewSQL10GCollector)
}

// NewSQL11GCollector
func NewSQL11GCollector() Collector {
	descs := [18]*prometheus.Desc{
		createNewDesc(sQLSystemName, "cpu_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "elapsed_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "executions_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "buffer_gets_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "disk_read_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "sort_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "phy_read_bytes_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "phy_read_request_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "phy_write_bytes_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "phy_write_request_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "parse_call_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "application_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "concurrency_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "cluster_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "user_io_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "plsql_exec_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "java_exec_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "rows_processed_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
	}

	return &sql11GCollector{descs}
}

// NewSQL10GCollector  .
func NewSQL10GCollector() Collector {
	descs := [14]*prometheus.Desc{
		createNewDesc(sQLSystemName, "cpu_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "elapsed_time_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "executions_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "buffer_gets_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "disk_read_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "sort_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "parse_call_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "application_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "concurrency_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "cluster_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "user_io_wait_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "plsql_exec_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "java_exec_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
		createNewDesc(sQLSystemName, "rows_processed_total", "Generic counter metric from v$sesstat view in Oracle.", []string{"username", "sql_id", "command", "child"}, nil),
	}
	return &sql10GCollector{descs}
}

func (c *sql11GCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sql11GSQL)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sqlID, username, commandType, child string
		var cpuTime, elapsedTime, executions, bufferGets, diskReads, sort, phyRB, phyRR, phyWB, phyWR, pc, awt, cwt, cwt2, uiwt, pet, jet, rp float64
		if err := rows.Scan(&sqlID, &child, &commandType, &username, &cpuTime, &elapsedTime, &bufferGets, &diskReads, &sort, &executions, &phyRB, &phyRR, &phyWB, &phyWR, &pc, &awt, &cwt, &cwt2, &uiwt, &pet, &jet, &rp); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, cpuTime, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, elapsedTime, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, executions, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, bufferGets, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[4], prometheus.CounterValue, diskReads, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[5], prometheus.CounterValue, sort, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[6], prometheus.CounterValue, phyRB, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[7], prometheus.CounterValue, phyRR, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[8], prometheus.CounterValue, phyWB, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[9], prometheus.CounterValue, phyWR, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[10], prometheus.CounterValue, pc, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[11], prometheus.CounterValue, awt, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[12], prometheus.CounterValue, cwt, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[13], prometheus.CounterValue, cwt2, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[14], prometheus.CounterValue, uiwt, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[15], prometheus.CounterValue, pet, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[16], prometheus.CounterValue, jet, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[17], prometheus.CounterValue, rp, username, sqlID, commandType, child)
	}
	return nil
}

func (c *sql10GCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sql10GSQL)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sqlID, username, commandType, child string
		var cpuTime, elapsedTime, executions, bufferGets, diskReads, sort, pc, awt, cwt, cwt2, uiwt, pet, jet, rp float64
		if err := rows.Scan(&sqlID, &child, &commandType, &username, &cpuTime, &elapsedTime, &bufferGets, &diskReads, &sort, &executions, &pc, &awt, &cwt, &cwt2, &uiwt, &pet, &jet, &rp); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(c.descs[0], prometheus.CounterValue, cpuTime, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[1], prometheus.CounterValue, elapsedTime, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[2], prometheus.CounterValue, executions, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[3], prometheus.CounterValue, bufferGets, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[4], prometheus.CounterValue, diskReads, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[5], prometheus.CounterValue, sort, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[6], prometheus.CounterValue, pc, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[7], prometheus.CounterValue, awt, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[8], prometheus.CounterValue, cwt, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[9], prometheus.CounterValue, cwt2, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[10], prometheus.CounterValue, uiwt, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[11], prometheus.CounterValue, pet, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[12], prometheus.CounterValue, jet, username, sqlID, commandType, child)
		ch <- prometheus.MustNewConstMetric(c.descs[13], prometheus.CounterValue, rp, username, sqlID, commandType, child)

	}
	return nil
}

const (
	sQLSystemName = "sql"
	sql11GSQL     = `
SELECT sql_id,
			 child_number,
			 (SELECT command_name FROM v$sqlcommand WHERE s.command_type = command_type),
			 parsing_schema_name,
       cpu_time,
			 elapsed_time,
			 buffer_gets,
       disk_reads,
       sorts,
			 executions,
			 physical_read_bytes,
			 physical_read_requests,
			 physical_write_bytes,
			 physical_write_requests,
			 parse_calls,
			 application_wait_time,
			 concurrency_wait_time,
			 cluster_wait_time,
			 user_io_wait_time,
			 plsql_exec_time,
			 java_exec_time,
			 rows_processed
  FROM v$sql s
 WHERE last_active_time >= sysdate - 1 / 24 / 60 AND is_obsolete ='N'`

	sql10GSQL = `
SELECT sql_id,
			 child_number,
			 (SELECT name FROM audit_actions WHERE s.command_type = action),
			 parsing_schema_name,
			 cpu_time,
			 elapsed_time,
			 buffer_gets,
			 disk_reads,
			 sorts,
			 executions,
			 user_io_wait_time,
			 parse_calls,
			 application_wait_time,
			 concurrency_wait_time,
			 cluster_wait_time,
			 plsql_exec_time,
			 java_exec_time,
			 rows_processed
FROM v$sql s
WHERE last_active_time >= sysdate - 1 / 24 / 60 AND is_obsolete ='N'`
)
