package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
)

var (
	power = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_power",
		Help: "Current real AC power being drawn, in Watts",
	}, []string{"name", "meter"})

	powerValid = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_power_valid",
		Help: "Whether power metering self-checks OK",
	}, []string{"name", "meter"})

	overpower = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_overpower",
		Help: "Value in Watts, on which an overpower condition is detected",
	}, []string{"name", "meter"})

	totalPower = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shellyplug_total_power",
		Help: "Total energy consumed by the attached electrical appliance in Watt-minute",
	}, []string{"name", "meter"})

	temperature = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_temperature",
		Help: "PlugS only internal device temperature in °C",
	}, []string{"name"})

	uptime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_uptime",
		Help: "Seconds elapsed since boot",
	}, []string{"name"})

	hasUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_has_update",
		Help: "Whether an update is available",
	}, []string{"name"})
)

func main() {
	cfgPath := flag.String("config", "shelly.yml", "Path to config file")
	flag.Parse()

	cfg := Config{
		Global: GlobalConfig{
			ScrapeInterval: 1 * time.Minute,
			ScrapeTimeout:  10 * time.Second,
		},
		Log: LogConfig{
			Level:     slog.LevelInfo,
			Format:    "json",
			AddSource: false,
		},
		Server: ServerConfig{
			ListenAddr: ":2112",
			Endpoint:   "/metrics",
		},
	}
	file, err := os.Open(*cfgPath)
	if err != nil {
		slog.Error("Failed to open config file", slog.Any("err", err))
		os.Exit(1)
	}
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		slog.Error("Failed to parse config file", slog.Any("err", err))
		os.Exit(1)
	}

	setupLogger(cfg.Log)
	slog.Info("Starting ShellyPlug Exporter", slog.Any("config", cfg))

	mux := http.NewServeMux()
	mux.Handle(cfg.Server.Endpoint, promhttp.Handler())
	server := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: mux,
	}
	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", slog.Any("err", err))
		}
	}()
	defer server.Shutdown(context.Background())

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()
	go startCollector(ctx, cfg)

	slog.Info("Started ShellyPlug Exporter", slog.String("addr", cfg.Server.ListenAddr))
	<-s
}

func setupLogger(cfg LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
