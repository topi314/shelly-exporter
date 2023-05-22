package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/exp/slog"
)

// These variables are set via the -ldflags option in go build
var (
	Name      = "gobin"
	Namespace = "github.com/topisenpai/shelly-exporter"

	Version   = "unknown"
	Commit    = "unknown"
	BuildTime = "unknown"
)

var (
	power = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_power",
		Help: "Current real AC power being drawn, in Watts",
	}, []string{"serial", "name", "meter"})

	powerValid = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_power_valid",
		Help: "Whether power metering self-checks OK",
	}, []string{"serial", "name", "meter"})

	overpower = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_overpower",
		Help: "Value in Watts, on which an overpower condition is detected",
	}, []string{"serial", "name", "meter"})

	totalPower = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shellyplug_total_power",
		Help: "Total energy consumed by the attached electrical appliance in Watt-minute",
	}, []string{"serial", "name", "meter"})

	temperature = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_temperature",
		Help: "PlugS only internal device temperature in °C",
	}, []string{"serial", "name"})

	uptime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_uptime",
		Help: "Seconds elapsed since boot",
	}, []string{"serial", "name"})

	hasUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_has_update",
		Help: "Whether an update is available",
	}, []string{"serial", "name"})
)

func main() {
	cfgPath := flag.String("config", "shelly.yml", "Path to config file")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		slog.Error("Failed to load config", slog.Any("err", err))
		os.Exit(1)
	}

	setupLogger(cfg.Log)
	slog.Info("Starting ShellyPlug Exporter", slog.Any("config", cfg))

	meter, closeFn, err := newMeter(cfg)
	if err != nil {
		slog.Error("Failed to create otel meter", slog.Any("err", err))
		os.Exit(1)
	}
	defer closeFn()

	if err = newExporter(cfg, meter); err != nil {
		slog.Error("Failed to create exporter", slog.Any("err", err))
		os.Exit(1)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
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
