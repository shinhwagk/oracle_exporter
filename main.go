package main

import (
	"fmt"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	// queryTimeout      = kingpin.Flag("query.timeout", "Query timeout (in seconds). (env: QUERY_TIMEOUT)").Default(getEnv("QUERY_TIMEOUT", "5")).String()
	multidatabaseAddr = kingpin.Flag("query.addr", "multidatabase addr").Default("").String()
)

// Metric name parts.
const (
	namespace = "oracle"
	exporter  = "exporter"
)

type handler struct {
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	maxRequests             int
	logger                  log.Logger
}

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
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
	up              prometheus.Gauge
}

// NewExporter returns a new Oracle DB exporter for the provided DSN.
func NewExporter(collects []string, dbid string, logger log.Logger) *Exporter {
	return &Exporter{
		logger:   logger,
		md:       MultiDatabase{Addr: *multidatabaseAddr, DbId: dbid},
		collects: collects,
		mp:       MetricProcessor{logger: logger},
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

	if err = e.md.Ping(); err != nil {
		e.up.Set(0)
		level.Error(e.logger).Log("Error pinging oracle:", err)
		return
	} else {
		e.up.Set(1)
		level.Debug(e.logger).Log("msg", "Successfully pinged Oracle database: ")
	}

	// resolveMetricFile()

	wg := sync.WaitGroup{}

	for _, metric := range metricsToScrap.Metric {
		if !filterMetric(e.collects, metric.Name) {
			continue
		}
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
				level.Error(e.logger).Log("Error scraping for", metric.Name, "_", metric.Context, ":", err)
				e.scrapeErrors.WithLabelValues(metric.Context).Inc()
			} else {
				level.Debug(e.logger).Log("Successfully scraped metric: ", metric.Context, metric.MetricsDesc, time.Since(scrapeStart))
			}
		}()
	}
	wg.Wait()
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

func newHandler(includeExporterMetrics bool, maxRequests int, logger log.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		maxRequests:             maxRequests,
		logger:                  logger,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}
	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	collects := r.URL.Query()["collect[]"]
	level.Debug(h.logger).Log("msg", "collect query:", "collects", collects)

	params_dbid := r.URL.Query()["dbid"]

	level.Debug(h.logger).Log("msg", "dbid query:", "collects", params_dbid)
	if len(params_dbid) != 1 || len(collects) == 0 {
		// level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", "err")))
		return
	}

	// To serve filtered metrics, we create a filtering handler on the fly.
	filteredHandler, err := h.innerHandler(params_dbid[0], collects)
	if err != nil {
		level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}
	filteredHandler.ServeHTTP(w, r)
}

func (h *handler) innerHandler(dbid string, collects []string) (http.Handler, error) {
	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector("oracle_exporter"))
	exporter := NewExporter(collects, dbid, h.logger)
	if err := r.Register(exporter); err != nil {
		return nil, fmt.Errorf("couldn't register node collector: %s", err)
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: h.maxRequests,
			Registry:            h.exporterMetricsRegistry,
		},
	)
	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9521").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
		configFile = kingpin.Flag(
			"web.config",
			"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
		).Default("").String()
		fileMetrics = kingpin.Flag(
			"file.metrics",
			"File with default metrics in a yaml file. (env: FILE_METRICS)",
		).Default("").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("oracle_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting oracledb_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	if err := resolveMetrics(*fileMetrics, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	http.Handle(*metricsPath, newHandler(!*disableExporterMetrics, *maxRequests, logger))
	// http.Handle(*metricPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Oracle Database Exporter ` + version.Version + `</title></head>
			<body>
			<h1>Oracle Database Exporter ` + version.Version + `</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)

	server := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(server, *configFile, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
