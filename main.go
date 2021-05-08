package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/godror/godror"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	// Version will be set at build time.
	Version       = "0.0.0.dev"
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry. (env: LISTEN_ADDRESS)").Default(getEnv("LISTEN_ADDRESS", ":9161")).String()
	metricPath    = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics. (env: TELEMETRY_PATH)").Default(getEnv("TELEMETRY_PATH", "/metrics")).String()
	fileMetrics   = kingpin.Flag("file.metrics", "File with default metrics in a yaml file. (env: FILE_METRICS)").Default(getEnv("FILE_METRICS", "default-metrics.toml")).String()
	queryTimeout  = kingpin.Flag("query.timeout", "Query timeout (in seconds). (env: QUERY_TIMEOUT)").Default(getEnv("QUERY_TIMEOUT", "5")).String()
	dataSource    = kingpin.Flag("database.datasource", "Number of maximum open connections in the connection pool. (env: DATABASE_MAXOPENCONNS)").String()
	maxIdleConns  = kingpin.Flag("database.maxIdleConns", "Number of maximum idle connections in the connection pool. (env: DATABASE_MAXIDLECONNS)").Default(getEnv("DATABASE_MAXIDLECONNS", "0")).Int()
	maxOpenConns  = kingpin.Flag("database.maxOpenConns", "Number of maximum open connections in the connection pool. (env: DATABASE_MAXOPENCONNS)").Default(getEnv("DATABASE_MAXOPENCONNS", "10")).Int()
)

// Metric name parts.
const (
	namespace = "oracle"
	exporter  = "exporter"
)

// global variables
var (
	metricFileShasum string
	dbCon            *sql.DB
	metricsToScrap   Metrics
)

// Metrics object description
type Metric struct {
	Name             string
	Context          string
	Labels           []string
	MetricsDesc      map[string]string
	MetricsType      map[string]string
	MetricsBuckets   map[string]map[string]string
	FieldToAppend    string
	Request          string
	IgnoreZeroResult bool
}

// Used to load multiple metrics from file
type Metrics struct {
	Metric []Metric
}

// Exporter collects Oracle DB metrics. It implements prometheus.Collector.
type Exporter struct {
	filter          []string
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
	up              prometheus.Gauge
}

// getEnv returns the value of an environment variable, or returns the provided fallback value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func createDatabaseConnect() {
	log.Infoln("Launching connection: ", *dataSource)
	db, err := sql.Open("godror", *dataSource)
	if err != nil {
		log.Errorln("Error while connecting to", *dataSource)
		panic(err)
	}
	log.Debugln("set max idle connections to ", *maxIdleConns)
	db.SetMaxIdleConns(*maxIdleConns)
	log.Debugln("set max open connections to ", *maxOpenConns)
	db.SetMaxOpenConns(*maxOpenConns)
	log.Debugln("Successfully connected to: ", *dataSource)
	dbCon = db
}

func getOracleVersion() string {
	var version string
	rows := dbCon.QueryRow("SELECT SUBSTR(version,0,2) FROM product_component_version WHERE UPPER(product) LIKE '%DATABASE%'")
	if err := rows.Scan(&version); err != nil {
		log.Errorln("quay database version faile:%s", err)
	}
	return version
}

// NewExporter returns a new Oracle DB exporter for the provided DSN.
func NewExporter(filters []string) *Exporter {
	return &Exporter{
		filter: filters,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from Oracle DB.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times Oracle DB was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occured scraping a Oracle database.",
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from Oracle DB resulted in an error (1 for error, 0 for success).",
		}),
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the Oracle database server is up.",
		}),
	}
}

// Collect implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()
	e.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)
	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
	ch <- e.up
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	var err error
	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
		if err == nil {
			e.error.Set(0)
		} else {
			e.error.Set(1)
		}
	}(time.Now())
	for {
		if err = dbCon.Ping(); err != nil {
			log.Errorln("Error pinging oracle:", err)
			e.up.Set(0)
			createDatabaseConnect()
		} else {
			log.Debugln("Successfully pinged Oracle database: ")
			e.up.Set(1)
			break
		}
		time.Sleep(time.Second * 5)
	}

	resolveMetricFile()

	wg := sync.WaitGroup{}

	for _, metric := range metricsToScrap.Metric {
		if filterMetric(e.filter, metric.Name) == false {
			continue
		}
		wg.Add(1)
		metric := metric //https://golang.org/doc/faq#closures_and_goroutines

		go func() {
			defer wg.Done()

			log.Debugln("About to scrape metric: ")
			log.Debugln("- Metric MetricsDesc: ", metric.MetricsDesc)
			log.Debugln("- Metric Name: ", metric.Name)
			log.Debugln("- Metric Context: ", metric.Context)
			log.Debugln("- Metric MetricsType: ", metric.MetricsType)
			log.Debugln("- Metric MetricsBuckets: ", metric.MetricsBuckets, "(Ignored unless Histogram type)")
			log.Debugln("- Metric Labels: ", metric.Labels)
			log.Debugln("- Metric FieldToAppend: ", metric.FieldToAppend)
			log.Debugln("- Metric IgnoreZeroResult: ", metric.IgnoreZeroResult)
			log.Debugln("- Metric Request: ", metric.Request)

			if len(metric.Request) == 0 {
				log.Errorln("Error scraping for ", metric.MetricsDesc, ". Did you forget to define request in your toml file?")
				return
			}

			if len(metric.MetricsDesc) == 0 {
				log.Errorln("Error scraping for query", metric.Request, ". Did you forget to define metricsdesc in your toml file?")
				return
			}

			for column, metricType := range metric.MetricsType {
				if metricType == "histogram" {
					_, ok := metric.MetricsBuckets[column]
					if !ok {
						log.Errorln("Unable to find MetricsBuckets configuration key for metric. (metric=" + column + ")")
						return
					}
				}
			}

			scrapeStart := time.Now()
			if err = ScrapeMetric(dbCon, ch, metric); err != nil {
				log.Errorln("Error scraping for", metric.Name, "_", metric.Context, ":", err)
				e.scrapeErrors.WithLabelValues(metric.Context).Inc()
			} else {
				log.Debugln("Successfully scraped metric: ", metric.Context, metric.MetricsDesc, time.Since(scrapeStart))
			}
		}()
	}
	wg.Wait()
}

func GetMetricType(metricType string, metricsType map[string]string) prometheus.ValueType {
	var strToPromType = map[string]prometheus.ValueType{
		"gauge":     prometheus.GaugeValue,
		"counter":   prometheus.CounterValue,
		"histogram": prometheus.UntypedValue,
	}

	strType, ok := metricsType[strings.ToLower(metricType)]
	if !ok {
		return prometheus.GaugeValue
	}
	valueType, ok := strToPromType[strings.ToLower(strType)]
	if !ok {
		panic(errors.New("Error while getting prometheus type " + strings.ToLower(strType)))
	}
	return valueType
}

// interface method to call ScrapeGenericValues using Metric struct values
func ScrapeMetric(db *sql.DB, ch chan<- prometheus.Metric, metricDefinition Metric) error {
	log.Debugln("Calling function ScrapeGenericValues()")
	return ScrapeGenericValues(db, ch, metricDefinition.Context, metricDefinition.Labels,
		metricDefinition.MetricsDesc, metricDefinition.MetricsType, metricDefinition.MetricsBuckets,
		metricDefinition.FieldToAppend, metricDefinition.IgnoreZeroResult,
		metricDefinition.Request)
}

// generic method for retrieving metrics.
func ScrapeGenericValues(db *sql.DB, ch chan<- prometheus.Metric, context string, labels []string,
	metricsDesc map[string]string, metricsType map[string]string, metricsBuckets map[string]map[string]string, fieldToAppend string, ignoreZeroResult bool, request string) error {
	metricsCount := 0
	genericParser := func(row map[string]string) error {
		// Construct labels value
		labelsValues := []string{}
		for _, label := range labels {
			labelsValues = append(labelsValues, row[label])
		}
		// Construct Prometheus values to sent back
		for metric, metricHelp := range metricsDesc {

			value, err := strconv.ParseFloat(strings.TrimSpace(row[metric]), 64)
			// If not a float, skip current metric
			if err != nil {
				log.Errorln("Unable to convert current value to float (metric=" + metric + ",metricHelp=" + metricHelp + ",value=<" + row[metric] + ">)")
				continue
			}
			log.Debugln("Query result looks like: ", value)
			// If metric do not use a field content in metric's name
			if strings.Compare(fieldToAppend, "") == 0 {
				desc := prometheus.NewDesc(
					prometheus.BuildFQName(namespace, context, metric),
					metricHelp,
					labels, nil,
				)
				if metricsType[strings.ToLower(metric)] == "histogram" {
					count, err := strconv.ParseUint(strings.TrimSpace(row["count"]), 10, 64)
					if err != nil {
						log.Errorln("Unable to convert count value to int (metric=" + metric + ",metricHelp=" + metricHelp + ",value=<" + row["count"] + ">)")
						continue
					}
					buckets := make(map[float64]uint64)
					for field, le := range metricsBuckets[metric] {
						lelimit, err := strconv.ParseFloat(strings.TrimSpace(le), 64)
						if err != nil {
							log.Errorln("Unable to convert bucket limit value to float (metric=" + metric + ",metricHelp=" + metricHelp + ",bucketlimit=<" + le + ">)")
							continue
						}
						counter, err := strconv.ParseUint(strings.TrimSpace(row[field]), 10, 64)
						if err != nil {
							log.Errorln("Unable to convert ", field, " value to int (metric="+metric+",metricHelp="+metricHelp+",value=<"+row[field]+">)")
							continue
						}
						buckets[lelimit] = counter
					}
					ch <- prometheus.MustNewConstHistogram(desc, count, value, buckets, labelsValues...)
				} else {
					ch <- prometheus.MustNewConstMetric(desc, GetMetricType(metric, metricsType), value, labelsValues...)
				}
				// If no labels, use metric name
			} else {
				desc := prometheus.NewDesc(
					prometheus.BuildFQName(namespace, context, cleanName(row[fieldToAppend])),
					metricHelp,
					nil, nil,
				)
				if metricsType[strings.ToLower(metric)] == "histogram" {
					count, err := strconv.ParseUint(strings.TrimSpace(row["count"]), 10, 64)
					if err != nil {
						log.Errorln("Unable to convert count value to int (metric=" + metric + ",metricHelp=" + metricHelp + ",value=<" + row["count"] + ">)")
						continue
					}
					buckets := make(map[float64]uint64)
					for field, le := range metricsBuckets[metric] {
						lelimit, err := strconv.ParseFloat(strings.TrimSpace(le), 64)
						if err != nil {
							log.Errorln("Unable to convert bucket limit value to float (metric=" + metric + ",metricHelp=" + metricHelp + ",bucketlimit=<" + le + ">)")
							continue
						}
						counter, err := strconv.ParseUint(strings.TrimSpace(row[field]), 10, 64)
						if err != nil {
							log.Errorln("Unable to convert ", field, " value to int (metric="+metric+",metricHelp="+metricHelp+",value=<"+row[field]+">)")
							continue
						}
						buckets[lelimit] = counter
					}
					ch <- prometheus.MustNewConstHistogram(desc, count, value, buckets)
				} else {
					ch <- prometheus.MustNewConstMetric(desc, GetMetricType(metric, metricsType), value)
				}
			}
			metricsCount++
		}
		return nil
	}
	err := GeneratePrometheusMetrics(db, genericParser, request)
	log.Debugln("ScrapeGenericValues() - metricsCount: ", metricsCount)
	if err != nil {
		return err
	}
	// if !ignoreZeroResult && metricsCount == 0 {
	// 	return errors.New("No metrics found while parsing")
	// }
	return err
}

// inspired by https://kylewbanks.com/blog/query-result-to-map-in-golang
// Parse SQL result and call parsing function to each row
func GeneratePrometheusMetrics(db *sql.DB, parse func(row map[string]string) error, query string) error {

	// Add a timeout
	timeout, err := strconv.Atoi(*queryTimeout)
	if err != nil {
		log.Fatal("error while converting timeout option value: ", err)
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, query)

	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("Oracle query timed out")
	}

	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	defer rows.Close()

	for rows.Next() {

		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]string)
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[strings.ToLower(colName)] = fmt.Sprintf("%v", *val)
		}
		// Call function to parse row
		if err := parse(m); err != nil {
			return err
		}
	}
	return nil
}

// Oracle gives us some ugly names back. This function cleans things up for Prometheus.
func cleanName(s string) string {
	s = strings.Replace(s, " ", "_", -1) // Remove spaces
	s = strings.Replace(s, "(", "", -1)  // Remove open parenthesis
	s = strings.Replace(s, ")", "", -1)  // Remove close parenthesis
	s = strings.Replace(s, "/", "", -1)  // Remove forward slashes
	s = strings.Replace(s, "*", "", -1)  // Remove asterisks
	s = strings.ToLower(s)
	return s
}

func hashFile(h hash.Hash, fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	return nil
}

func checkMetricFileChanged(content []byte) (bool, string) {
	newShasum := sha256sum(content)
	isChange := newShasum == metricFileShasum
	return !isChange, newShasum
}

func sha256sum(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return string(h.Sum(nil))
}

func readMetricFile() ([]byte, error) {
	metricsPath := *fileMetrics
	var metricsBytes []byte
	var err error
	if strings.HasPrefix(metricsPath, "http://") || strings.HasPrefix(metricsPath, "https://") {
		var resp *http.Response
		resp, err = http.Get(metricsPath)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return nil, err
		} else {
			metricsBytes = body
		}
	} else {
		var err error
		metricsBytes, err = ioutil.ReadFile(metricsPath)
		if err != nil {
			return nil, err
		}
	}
	return metricsBytes, nil
}

func resolveMetrics(content []byte) {
	// Truncate metricsToScrap
	metricsToScrap.Metric = []Metric{}
	// Load default metrics
	if err := yaml.Unmarshal(content, &metricsToScrap.Metric); err != nil {
		log.Errorln(err)
		panic(errors.New("Error loading metric file content " + *fileMetrics))
	} else {
		log.Infoln("Successfully loaded metrics yaml from: " + *fileMetrics)
	}
}

func filterMetric(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func resolveMetricFile() {
	if metricContentByte, err := readMetricFile(); err != nil {
		// panic(errors.New("Read Metric File:" + err.Error()))
		var errReadFile error = errors.New("Read Metric File:" + err.Error())
		log.Errorf("error:%v", errReadFile)
	} else {
		if isChanged, shasum := checkMetricFileChanged(metricContentByte); isChanged {
			metricFileShasum = shasum
			resolveMetrics(metricContentByte)
			for _, n := range metricsToScrap.Metric {
				log.Infof(" - %s", n.Name)
			}
		}
	}
}

func newHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := r.URL.Query()["collect[]"]
		log.Debugln("collect query:", filters)
		registry := prometheus.NewRegistry()
		exporter := NewExporter(filters)
		err := registry.Register(exporter)
		if err != nil {
			log.Errorln("Prometheus register error %s", err)
		}
		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version("oracle_exporter " + Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting oracledb_exporter " + Version)

	// init
	createDatabaseConnect()
	resolveMetricFile()
	// dbVersion = getOracleVersion()

	handlerFunc := newHandler()
	http.Handle(*metricPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><head><title>Oracle DB Exporter " + Version + "</title></head><body><h1>Oracle DB Exporter " + Version + "</h1><p><a href='" + *metricPath + "'>Metrics</a></p></body></html>"))
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
