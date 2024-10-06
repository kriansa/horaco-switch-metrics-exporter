package exporter

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kriansa/switch-exporter/pkg/scraper"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PortEnabled = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "switch_port_enabled",
		Help: "Whether the switch is enabled or not",
	}, []string{"port"})

	PortConnected = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "switch_port_connected",
		Help: "Whether the switch is connected or not",
	}, []string{"port"})

	PortSpeed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "switch_port_speed",
		Help: "The maximum rated speed of the switch port in Mbit/s",
	}, []string{"port", "transmission_mode"})

	Packets = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "switch_port_packets_total",
		Help: "The number of packets (counter) transmitted or received by the switch port",
	}, []string{"port", "direction", "status"})
)

func StartServer(bindAddress string, switchScraper *scraper.SwitchScraper) error {
	prometheus.MustRegister(PortEnabled, PortConnected, PortSpeed, Packets)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	PoolMetrics(switchScraper)

	slog.Info("starting server", slog.String("address", bindAddress))
	return http.ListenAndServe(bindAddress, mux)
}

func PoolMetrics(switchScraper *scraper.SwitchScraper) {
	boolToFloat := map[bool]float64{true: 1, false: 0}
	speedToFloat := map[scraper.LinkSpeed]float64{
		scraper.LinkSpeed10Mbps:  10,
		scraper.LinkSpeed100Mbps: 100,
		scraper.LinkSpeed1Gbps:   1000,
		scraper.LinkSpeed2_5Gbps: 2500,
		scraper.LinkSpeed10Gbps:  10000,
		"":                       0,
	}

	go func() {
		for {
			metrics, err := switchScraper.FetchData()
			if err != nil {
				fmt.Println("error collecting switch metrics: %w", err)
				os.Exit(1)
			}

			for port, metric := range metrics {
				PortEnabled.WithLabelValues(port).Set(boolToFloat[metric.Enabled])
				PortConnected.WithLabelValues(port).Set(boolToFloat[metric.Connected])
				PortSpeed.WithLabelValues(port, string(metric.TransmissionMode)).Set(speedToFloat[metric.Speed])

				// Use the four attributes of the PortStats struct to set the packets metric
				Packets.WithLabelValues(port, "sent", "success").Set(float64(metric.TxGood))
				Packets.WithLabelValues(port, "sent", "error").Set(float64(metric.TxBad))
				Packets.WithLabelValues(port, "received", "success").Set(float64(metric.RxGood))
				Packets.WithLabelValues(port, "received", "error").Set(float64(metric.RxBad))
			}

			time.Sleep(1 * time.Second)
		}
	}()
}
