package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	// Version will be set at build time.
	Version           = "0.0.0.dev"
	listenAddress     = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry. (env: LISTEN_ADDRESS)").Default(getEnv("LISTEN_ADDRESS", ":9161")).String()
	metricPath        = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics. (env: TELEMETRY_PATH)").Default(getEnv("TELEMETRY_PATH", "/metrics")).String()
	fileMetrics       = kingpin.Flag("file.metrics", "File with default metrics in a yaml file. (env: FILE_METRICS)").Default(getEnv("FILE_METRICS", "default-metrics.toml")).String()
	queryTimeout      = kingpin.Flag("query.timeout", "Query timeout (in seconds). (env: QUERY_TIMEOUT)").Default(getEnv("QUERY_TIMEOUT", "5")).String()
	multidatabaseAddr = os.Getenv("MULTIDATABASE_ADDR") // prometheus oracle exporter
	multidatabaseDbId = os.Getenv("MULTIDATABASE_DBID")
)

// Metric name parts.
const (
	namespace = "oracle"
	exporter  = "exporter"
)

// global variables
var (
	metricFileShasum string
	metricsToScrap   Metrics
	md               MultiDatabase
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
		if err = md.Ping(); err != nil {
			log.Errorln("Error pinging oracle:", err)
			e.up.Set(0)
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
		if !filterMetric(e.filter, metric.Name) {
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
			if err = ScrapeMetric(ch, metric); err != nil {
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
func ScrapeMetric(ch chan<- prometheus.Metric, metricDefinition Metric) error {
	log.Debugln("Calling function ScrapeGenericValues()")
	return ScrapeGenericValues(ch, metricDefinition.Context, metricDefinition.Labels,
		metricDefinition.MetricsDesc, metricDefinition.MetricsType, metricDefinition.MetricsBuckets,
		metricDefinition.FieldToAppend, metricDefinition.IgnoreZeroResult,
		metricDefinition.Request)
}

// generic method for retrieving metrics.
func ScrapeGenericValues(ch chan<- prometheus.Metric, context string, labels []string,
	metricsDesc map[string]string, metricsType map[string]string, metricsBuckets map[string]map[string]string, fieldToAppend string, ignoreZeroResult bool, sqlText string) error {
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
	err := GeneratePrometheusMetrics(genericParser, sqlText)
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
func GeneratePrometheusMetrics(parse func(row map[string]string) error, sqlText string) error {
	rows, err := md.Query(sqlText)

	// if ctx.Err() == context.DeadlineExceeded {
	// 	return errors.New("Oracle query timed out")
	// }

	if err != nil {
		return err
	}

	for _, row := range rows {

		m := make(map[string]string)
		for col, val := range row {
			m[strings.ToLower(col)] = fmt.Sprintf("%v", val)
		}

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

type MultiDatabase struct {
	Addr string
	DbId string
}

type MultiDatabasePost struct {
	DbId    string        `json:"db_id"`
	SqlText string        `json:"sql_text"`
	Binds   []interface{} `json:"binds"`
}

type MultiDatabaseResult struct {
	Code   int                      `json:"code"`
	DbId   string                   `json:"db_id"`
	Error  string                   `json:"error"`
	Result []map[string]interface{} `json:"result"`
}

func (md MultiDatabase) Ping() error {
	_, err := md.Query("select * from dual")
	return err
}

func (md MultiDatabase) Query(sqlText string) ([]map[string]interface{}, error) {
	json_data, err := json.Marshal(MultiDatabasePost{md.DbId, sqlText, []interface{}{}})

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s/query", md.Addr)
	resp, err := http.Post(url, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		return nil, err
	}

	var mdr MultiDatabaseResult

	json.NewDecoder(resp.Body).Decode(&mdr)

	if mdr.Code == 1 {
		return nil, errors.New(mdr.Error)
	}
	return mdr.Result, nil
}

func main() {

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version("oracle_exporter " + Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting oracledb_exporter " + Version)

	// init
	resolveMetricFile()
	md = MultiDatabase{Addr: multidatabaseAddr, DbId: multidatabaseDbId}

	handlerFunc := newHandler()
	http.Handle(*metricPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><head><title>Oracle DB Exporter " + Version + "</title></head><body><h1>Oracle DB Exporter " + Version + "</h1><p><a href='" + *metricPath + "'>Metrics</a></p></body></html>"))
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
