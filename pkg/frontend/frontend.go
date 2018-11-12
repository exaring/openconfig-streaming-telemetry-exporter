package frontend

import (
	"fmt"
	"net/http"

	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/collector"
	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promlog "github.com/prometheus/common/log"
	log "github.com/sirupsen/logrus"
)

// Frontend represents an HTTP frontend
type Frontend struct {
	cfg       *config.Config
	collector *collector.Collector
}

// New creates a new HTTP frontend
func New(cfg *config.Config, collector *collector.Collector) *Frontend {
	return &Frontend{
		cfg:       cfg,
		collector: collector,
	}
}

// Start starts the frontend
func (fe *Frontend) Start() {
	log.Infof("Starting OpenConfig Streaming Telemetry Exporter (Version: %s)\n", fe.cfg.Version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>OpenConfig Streaming Telemetry Exporter (Version ` + fe.cfg.Version + `)</title></head>
			<body>
			<h1>OpenConfig Streaming Telemetry Exporter</h1>
			<p><a href="` + fe.cfg.MetricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/exaring/openconfig-streaming-telemetry-exporter">github.com/exaring/openconfig-streaming-telemetry-exporterr</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(fe.cfg.MetricsPath, fe.handleMetricsRequest)
	http.HandleFunc("/debug/dump", fe.handleDumpRequest)

	log.Infof("Listening for %s on %s\n", fe.cfg.MetricsPath, fe.cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(fe.cfg.ListenAddress, nil))
}

func (fe *Frontend) handleDumpRequest(w http.ResponseWriter, r *http.Request) {
	for _, line := range fe.collector.Dump() {
		w.Write([]byte(line))
	}
}

func (fe *Frontend) handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(fe.collector)

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      promlog.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}

func (fe *Frontend) targetsForRequest(r *http.Request) ([]*config.Target, error) {
	reqTarget := r.URL.Query().Get("target")
	if reqTarget == "" {
		return fe.cfg.Targets, nil
	}

	for _, t := range fe.cfg.Targets {
		if t.Hostname == reqTarget {
			return []*config.Target{t}, nil
		}
	}

	return nil, fmt.Errorf("the target '%s' is not defined in the configuration file", reqTarget)
}
