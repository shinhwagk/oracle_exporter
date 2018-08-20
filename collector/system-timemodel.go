package collector

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type sysTimeModelCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("systemTimeModel-10g", NewSysTimeModelCollector)
	registerCollector("systemTimeModel-11g", NewSysTimeModelCollector)
}

// NewSysTimeModelCollector
func NewSysTimeModelCollector() Collector {
	descs := make(map[string]*prometheus.Desc)
	descs["3649082374"] = createNewDesc(sysTimeModelSystemName, "db_time", "Generic counter metric.", nil, nil)
	descs["2748282437"] = createNewDesc(sysTimeModelSystemName, "db_cpu", "Generic counter metric.", nil, nil)
	descs["4157170894"] = createNewDesc(sysTimeModelSystemName, "bgd_elap_time", "Generic counter metric.", nil, nil)
	descs["2451517896"] = createNewDesc(sysTimeModelSystemName, "bgd_cpu_time", "Generic counter metric.", nil, nil)
	descs["4127043053"] = createNewDesc(sysTimeModelSystemName, "seq_load_elap_time", "Generic counter metric.", nil, nil)
	descs["1431595225"] = createNewDesc(sysTimeModelSystemName, "parse_elap", "Generic counter metric.", nil, nil)
	descs["372226525"] = createNewDesc(sysTimeModelSystemName, "hard_parse_elap_time", "Generic counter metric.", nil, nil)
	descs["2821698184"] = createNewDesc(sysTimeModelSystemName, "sql_exec_elap_time", "Generic counter metric.", nil, nil)
	descs["1990024365"] = createNewDesc(sysTimeModelSystemName, "conn_manage_call_elap_time", "Generic counter metric.", nil, nil)
	descs["1824284809"] = createNewDesc(sysTimeModelSystemName, "failed_parse_elap_time", "Generic counter metric.", nil, nil)
	descs["4125607023"] = createNewDesc(sysTimeModelSystemName, "faile_parse_outofsharedmemory_elap_time", "Generic counter metric.", nil, nil)
	descs["3138706091"] = createNewDesc(sysTimeModelSystemName, "hard_parse_sharingcriteria_elap_time", "Generic counter metric.", nil, nil)
	descs["268357648"] = createNewDesc(sysTimeModelSystemName, "hard_parse_bind_mismatch_elap_time", "Generic counter metric.", nil, nil)
	descs["2643905994"] = createNewDesc(sysTimeModelSystemName, "plsql_exec_elap_time", "Generic counter metric.", nil, nil)
	descs["290749718"] = createNewDesc(sysTimeModelSystemName, "inbound_plsql_rpc_elap_time", "Generic counter metric.", nil, nil)
	descs["1311180441"] = createNewDesc(sysTimeModelSystemName, "plsql_compilation_elap_time", "Generic counter metric.", nil, nil)
	descs["751169994"] = createNewDesc(sysTimeModelSystemName, "java_exec_elap_time", "Generic counter metric.", nil, nil)
	descs["1159091985"] = createNewDesc(sysTimeModelSystemName, "repeated_bind_elap_time", "Generic counter metric.", nil, nil)
	descs["2411117902"] = createNewDesc(sysTimeModelSystemName, "rman_cpu_time", "Generic counter metric.", nil, nil)
	return &sysTimeModelCollector{descs}
}

func (c *sysTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(systemTimeModelSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var value float64
		if err := rows.Scan(&id, &value); err != nil {
			return err
		}

		desc, ok := c.descs[id]

		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value)
		} else {
			err := fmt.Sprintf("system time model: %s no exist", id)
			return errors.New(err)
		}

	}
	return nil
}

const (
	sysTimeModelSystemName = "systimemodel"
	systemTimeModelSQL     = `SELECT stat_id, value FROM v$sys_time_model`
)
