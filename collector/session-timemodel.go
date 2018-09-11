package collector

import (
	"database/sql"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

type sessTimeModelCollector struct {
	descs map[string]*prometheus.Desc
}

func init() {
	registerCollector("sessionTimeModel-10g", NewSessTimeModelCollector)
	registerCollector("sessionTimeModel-11g", NewSessTimeModelCollector)
}

// NewSessTimeModelCollector.
func NewSessTimeModelCollector() Collector {
	descs := make(map[string]*prometheus.Desc)
	descs["3649082374"] = createNewDesc(sessTimeModelSystemName, "db_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["2748282437"] = createNewDesc(sessTimeModelSystemName, "db_cpu", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["4157170894"] = createNewDesc(sessTimeModelSystemName, "bgd_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["2451517896"] = createNewDesc(sessTimeModelSystemName, "bgd_cpu_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["4127043053"] = createNewDesc(sessTimeModelSystemName, "seq_load_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["1431595225"] = createNewDesc(sessTimeModelSystemName, "parse_elap", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["372226525"] = createNewDesc(sessTimeModelSystemName, "hard_parse_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["2821698184"] = createNewDesc(sessTimeModelSystemName, "sql_exec_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["1990024365"] = createNewDesc(sessTimeModelSystemName, "conn_manage_call_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["1824284809"] = createNewDesc(sessTimeModelSystemName, "failed_parse_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["4125607023"] = createNewDesc(sessTimeModelSystemName, "faile_parse_outofsharedmemory_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["3138706091"] = createNewDesc(sessTimeModelSystemName, "hard_parse_sharingcriteria_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["268357648"] = createNewDesc(sessTimeModelSystemName, "hard_parse_bind_mismatch_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["2643905994"] = createNewDesc(sessTimeModelSystemName, "plsql_exec_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["290749718"] = createNewDesc(sessTimeModelSystemName, "inbound_plsql_rpc_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["1311180441"] = createNewDesc(sessTimeModelSystemName, "plsql_compilation_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["751169994"] = createNewDesc(sessTimeModelSystemName, "java_exec_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["1159091985"] = createNewDesc(sessTimeModelSystemName, "repeated_bind_elap_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	descs["2411117902"] = createNewDesc(sessTimeModelSystemName, "rman_cpu_time", "Generic counter metric.", []string{"sid", "username"}, nil)
	return &sessTimeModelCollector{descs}
}

func (c *sessTimeModelCollector) Update(db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.Query(sessTimeModelSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, username, sid string
		var value float64
		if err := rows.Scan(&sid, &username, &id, &value); err != nil {
			return err
		}

		desc, ok := c.descs[id]

		if ok {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, sid, username)
		} else {
			return errors.New("system time model no exist")
		}

	}
	return nil
}

const (
	sessTimeModelSystemName = "sesstimemodel"
	sessTimeModelSQL        = `
	SELECT s.sid, s.username, stm.stat_id, stm.value
	FROM v$sess_time_model stm, v$session s
	WHERE stm.sid = s.sid AND s.username is not null AND stm.stat_name in ('DB time')`
)
