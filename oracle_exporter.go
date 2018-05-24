package main

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/shinhwagk/oracle_exporter/collector"

	_ "gopkg.in/goracle.v2"
)

func init() {
	prometheus.MustRegister(version.NewCollector("oracle_exporter"))
}

func cycleHandler(cycle string) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		filters := r.URL.Query()["collect[]"]
		log.Debugln("collect query:", filters)

		nc, err := collector.NewOracleCollector(cycle, filters...)
		if err != nil {
			log.Warnln("Couldn't create", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Couldn't create %s", err)))
			return
		}

		registry := prometheus.NewRegistry()
		err = registry.Register(nc)
		if err != nil {
			log.Errorln("Couldn't register collector:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't register collector: %s", err)))
			return
		}

		gatherers := prometheus.Gatherers{prometheus.DefaultGatherer, registry}

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.InstrumentMetricHandler(
			registry,
			promhttp.HandlerFor(gatherers,
				promhttp.HandlerOpts{
					ErrorLog:      log.NewErrorLogger(),
					ErrorHandling: promhttp.ContinueOnError,
				}),
		)
		h.ServeHTTP(w, r)
	}
	return handler
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9100").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		cycle         = kingpin.Flag("web.cycle", "Address on which to expose metrics and web interface.").Default("m").String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("oracle_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting oracle_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	nc, err := collector.NewOracleCollector(*cycle)
	if err != nil {
		log.Fatalf("Couldn't create collector: %s", err)
	}

	// print endable collectors with sort.
	log.Infof("Enabled collectors:")
	collectors := []string{}
	for n := range nc.Collectors {
		collectors = append(collectors, n)
	}
	sort.Strings(collectors)
	for _, n := range collectors {
		log.Infof(" - %s", n)
	}

	http.HandleFunc(*metricsPath, cycleHandler(*cycle))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Oracle Exporter</title></head>
			<body>
			<h1>Oracle Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
