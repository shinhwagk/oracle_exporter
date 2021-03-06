package collector

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "oracle"

var (
	scrapeDurationDesc = createNewDesc("scrape", "collector_duration_seconds", "oracle_exporter: Duration of a collector scrape.", []string{"collector"}, nil)
	scrapeSuccessDesc  = createNewDesc("scrape", "collector_success", "oracle_exporter: Whether a collector succeeded.", []string{"collector"}, nil)
	oracleUpDesc       = createNewDesc("", "up", "oracle_exporter: Whether the Oracle server is up.", nil, nil)
)

var (
	factories = make(map[string]func() Collector)
)

// OracleCollector implements the prometheus.Collector interface.
type oracleCollector struct {
	Collectors map[string]Collector
}

func registerCollector(collector string, factory func() Collector) {
	factories[collector] = factory
}

// NewOracleCollector creates a new OracleCollector
func NewOracleCollector(filters ...string) (*oracleCollector, error) {
	f := make(map[string]bool)
	for _, filter := range filters {
		_, exist := factories[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		f[filter] = true
	}

	collectors := make(map[string]Collector)
	for c, factory := range factories {
		if len(f) == 0 || f[c] {
			collectors[c] = factory()
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
		log.Errorf("ERROR: metric [%s] collector failed after %fs: %s", name, duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("OK: metric [%s] collector succeeded after %fs.", name, duration.Seconds())
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

// Collector is the interface a collector has to implement.
type Collector interface {
	Update(db *sql.DB, ch chan<- prometheus.Metric) error
}

func createNewDesc(subsystem string, name string, help string, vls []string, cls prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, name), help, vls, cls)
}
