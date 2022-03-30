package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	// queryTimeout      = kingpin.Flag("query.timeout", "Query timeout (in seconds). (env: QUERY_TIMEOUT)").Default(getEnv("QUERY_TIMEOUT", "5")).String()
	multidatabaseAddr = kingpin.Flag("mdb.addr", "multidatabase addr").Default("").String()
)

// Metric name parts.
const (
	namespace = "oracle"
	exporter  = "exporter"
)

// global variables
var (
	metricsToScrap Metrics
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
	logger          log.Logger
	md              MultiDatabase
	mp              MetricProcessor
	collects        []string
	Duration, Error prometheus.Gauge
	TotalScrapes    prometheus.Counter
	ScrapeErrors    *prometheus.CounterVec
	OracleUp        prometheus.Gauge
}

// NewExporter returns a new Oracle DB exporter for the provided DSN.
func NewExporter(collects []string, dsn string, logger log.Logger) *Exporter {
	return &Exporter{
		logger:   logger,
		md:       MultiDatabase{addr: *multidatabaseAddr, dsn: dsn},
		collects: collects,
		mp:       MetricProcessor{logger: logger},
		Duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from Oracle DB.",
		}),
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times Oracle DB was scraped for metrics.",
		}),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occured scraping a Oracle database.",
		}, []string{"collector"}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from Oracle DB resulted in an error (1 for error, 0 for success).",
		}),
		OracleUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the Oracle database server is up.",
		}),
	}
}

// Collect implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.TotalScrapes.Desc()
	ch <- e.Error.Desc()
	e.ScrapeErrors.Describe(ch)
	ch <- e.OracleUp.Desc()
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)
	ch <- e.TotalScrapes
	ch <- e.Error
	e.ScrapeErrors.Collect(ch)
	ch <- e.OracleUp
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.TotalScrapes.Inc()
	var err error
	defer func(begun time.Time) {
		e.Duration.Set(time.Since(begun).Seconds())
		if err == nil {
			e.Error.Set(0)
		} else {
			e.Error.Set(1)
		}
	}(time.Now())

	if err = e.md.Ping(); err != nil {
		e.OracleUp.Set(0)
		level.Error(e.logger).Log("Error pinging oracle", err)
		return
	}

	e.OracleUp.Set(1)

	// resolveMetricFile()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, metric := range metricsToScrap.Metric {
		if !filterMetric(e.collects, metric.Name) {
			continue
		}

		level.Debug(e.logger).Log("msg", "scrap detail", "dsn", e.md.dsn, "metric", metric.Name, "metric context", metric.Context)

		wg.Add(1)
		metric := metric //https://golang.org/doc/faq#closures_and_goroutines

		go func() {
			defer wg.Done()

			level.Debug(e.logger).Log("About to scrape metric: ")
			level.Debug(e.logger).Log("- Metric MetricsDesc: ", metric.MetricsDesc)
			level.Debug(e.logger).Log("- Metric Name: ", metric.Name)
			level.Debug(e.logger).Log("- Metric Context: ", metric.Context)
			level.Debug(e.logger).Log("- Metric MetricsType: ", metric.MetricsType)
			level.Debug(e.logger).Log("- Metric MetricsBuckets: ", metric.MetricsBuckets, "(Ignored unless Histogram type)")
			level.Debug(e.logger).Log("- Metric Labels: ", metric.Labels)
			level.Debug(e.logger).Log("- Metric FieldToAppend: ", metric.FieldToAppend)
			level.Debug(e.logger).Log("- Metric IgnoreZeroResult: ", metric.IgnoreZeroResult)
			level.Debug(e.logger).Log("- Metric Request: ", metric.Request)

			if len(metric.Request) == 0 {
				level.Error(e.logger).Log("Error scraping for ", metric.MetricsDesc, ". Did you forget to define request in your toml file?")
				return
			}

			if len(metric.MetricsDesc) == 0 {
				level.Error(e.logger).Log("Error scraping for query", metric.Request, ". Did you forget to define metricsdesc in your toml file?")
				return
			}

			for column, metricType := range metric.MetricsType {
				if metricType == "histogram" {
					_, ok := metric.MetricsBuckets[column]
					if !ok {
						level.Error(e.logger).Log("Unable to find MetricsBuckets configuration key for metric. (metric=" + column + ")")
						return
					}
				}
			}

			scrapeStart := time.Now()
			if err = e.mp.ScrapeMetric(ch, e.md, metric); err != nil {
				level.Error(e.logger).Log("msg", "Error Scrape", "dsn", e.md.dsn, "metric", metric.Name, "context", metric.Context, "err", err)
				e.ScrapeErrors.WithLabelValues(metric.Context).Inc()
			} else {
				level.Debug(e.logger).Log("Successfully scraped metric: ", metric.Context, metric.MetricsDesc, time.Since(scrapeStart))
			}
		}()
	}
}

func resolveMetrics(metricsPath string, logger log.Logger) error {
	var metricsBytes []byte
	var err error
	if strings.HasPrefix(metricsPath, "http://") || strings.HasPrefix(metricsPath, "https://") {
		var resp *http.Response
		resp, err = http.Get(metricsPath)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else {
			metricsBytes = body
		}
	} else {
		var err error
		metricsBytes, err = ioutil.ReadFile(metricsPath)
		if err != nil {
			return err
		}
	}

	metricsToScrap.Metric = []Metric{}
	// Load default metrics
	err = yaml.Unmarshal(metricsBytes, &metricsToScrap.Metric)
	return err
}

func filterMetric(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func init() {
	prometheus.MustRegister(version.NewCollector("oracle_exporter"))
}

func newHandler(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param_collects := r.URL.Query()["collect[]"]
		param_dsn := r.URL.Query()["dsn"]

		if len(param_dsn) != 1 || len(param_collects) == 0 {
			// level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Parameters are not set correctly."))
			return
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(NewExporter(param_collects, param_dsn[0], logger))

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}

		handler := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		handler.ServeHTTP(w, r)
	}
}

func main() {
	var (
		webConfig     = webflag.AddFlags(kingpin.CommandLine)
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9521").String()
		metricPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		metricFile = kingpin.Flag(
			"file.metrics",
			"File with default metrics in a yaml file. (env: FILE_METRICS)",
		).Default("").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("oracle_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting oracledb_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	if err := resolveMetrics(*metricFile, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	// http.Handle(*metricsPath, newHandler(*maxRequests, logger))
	handlerFunc := newHandler(logger)
	http.Handle(*metricPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Oracle Database Exporter ` + version.Version + `</title></head>
			<body>
			<h1>Oracle Database Exporter ` + version.Version + `</h1>
			<p><a href="` + *metricPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)

	server := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(server, *webConfig, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
