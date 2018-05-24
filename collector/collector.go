package collector

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

const namespace = "oracle"

var (
	scrapeDurationDesc = newDesc("scrape", "collector_duration_seconds", "oracle_exporter: Duration of a collector scrape.", []string{"collector"}, nil)
	scrapeSuccessDesc  = newDesc("scrape", "collector_success", "oracle_exporter: Whether a collector succeeded.", []string{"collector"}, nil)
	oracleUpDesc       = newDesc("", "up", "oracle_exporter: Whether the Oracle server is up.", nil, nil)
)

// func warnDeprecated(collector string) {
// 	log.Warnf("The %s collector is deprecated and will be removed in the future!", collector)
// }

const (
	defaultEnabled  = true
	defaultDisabled = false
	cHour           = "h"
	cMin            = "m"
	cDay            = "d"
)

var (
	factoriesMinute      = make(map[string]func() (Collector, error))
	collectorStateMinute = make(map[string]*bool)
	factoriesHour        = make(map[string]func() (Collector, error))
	collectorStateHour   = make(map[string]*bool)
	factoriesDay         = make(map[string]func() (Collector, error))
	collectorStateDay    = make(map[string]*bool)
)

func registerCollector(collector string, cycle string, isDefaultEnabled bool, factory func() (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)
	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Bool()
	if cycle == cMin {
		collectorStateMinute[collector] = flag
		factoriesMinute[collector] = factory
	}
	if cycle == cHour {
		collectorStateHour[collector] = flag
		factoriesHour[collector] = factory
	}
}

// OracleCollector implements the prometheus.Collector interface.
type oracleCollector struct {
	Collectors map[string]Collector
}

// NewOracleCollector creates a new OracleCollector
func NewOracleCollector(cycle string, filters ...string) (*oracleCollector, error) {
	if cycle == cMin {
		return NewOracleCollectorByCycle(factoriesMinute, collectorStateMinute, filters)
	} else if cycle == cHour {
		return NewOracleCollectorByCycle(factoriesHour, collectorStateHour, filters)
	} else {
		return NewOracleCollectorByCycle(factoriesDay, collectorStateDay, filters)
	}
}

func NewOracleCollectorByCycle(factories map[string]func() (Collector, error), collectorState map[string]*bool, filters []string) (*oracleCollector, error) {
	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if *enabled {
			collector, err := factories[key]() // get each collector desc
			if err != nil {
				return nil, err
			}
			if len(f) == 0 || f[key] {
				collectors[key] = collector
			}
		}
	}
	return &oracleCollector{collectors}, nil
}

// Describe implements the prometheus.Collector interface.
func (n oracleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
	ch <- oracleUpDesc
}

// Collect implements the prometheus.Collector interface.
func (n oracleCollector) Collect(ch chan<- prometheus.Metric) {
	begin := time.Now()
	dsn := os.Getenv("DATA_SOURCE_NAME")
	db, err := sql.Open("goracle", dsn)
	if err != nil {
		log.Errorln("Error opening connection to database:", err)
		ch <- prometheus.MustNewConstMetric(oracleUpDesc, prometheus.GaugeValue, float64(0))
		return
	}
	defer db.Close()

	isUpRows, err := db.Query("SELECT * FROM dual")
	if err != nil {
		log.Errorln("Error pinging Pracle:", err)
		ch <- prometheus.MustNewConstMetric(oracleUpDesc, prometheus.GaugeValue, float64(0))
		return
	}
	isUpRows.Close()

	ch <- prometheus.MustNewConstMetric(oracleUpDesc, prometheus.GaugeValue, float64(1))
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(begin).Seconds(), "connection")

	for name, c := range n.Collectors {
		func(name string, c Collector) {
			execute(db, name, c, ch)
		}(name, c)
	}
}

func execute(db *sql.DB, name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(db, ch)
	duration := time.Since(begin)

	var success float64 = 1

	if err != nil {
		log.Errorf("ERROR: %s collector failed after %fs: %s", name, duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("OK: %s collector succeeded after %fs.", name, duration.Seconds())
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

// Collector is the interface a collector has to implement.
type Collector interface {
	Update(db *sql.DB, ch chan<- prometheus.Metric) error
}

func newDesc(subsystem string, name string, help string, vls []string, cls prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, name), help, vls, cls)
}
