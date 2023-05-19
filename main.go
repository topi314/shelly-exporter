package main

import (
	"context"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	power = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_power",
		Help: "Current power drawn in watts",
	}, []string{"serial", "name", "meter"})

	overpower = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_overpower",
		Help: "Overpower drawn in watts/minute",
	}, []string{"serial", "name", "meter"})

	totalPower = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "shellyplug_total_power",
		Help: "Total power drawn in watts",
	}, []string{"serial", "name", "meter"})

	temperature = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_temperature",
		Help: "Temperature in degrees celsius",
	}, []string{"serial", "name"})

	uptime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_uptime",
		Help: "Uptime in seconds",
	}, []string{"serial", "name"})
)

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func main() {
	cfgPath := flag.String("config", "shelly.yml", "Path to config file")
	debug := flag.Bool("debug", false, "debug mode")
	logType := flag.String("log", "json", "log format, one of: json, text")
	logLevel := flag.String("log-level", "info", "log level, one of: debug, info, warn, error")
	flag.Parse()

	setupLogger(*debug, *logLevel, *logType)

	cfg := Config{
		Global: GlobalConfig{
			ScrapeInterval: 30 * time.Second,
			ScrapeTimeout:  5 * time.Second,
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

func setupLogger(debug bool, level string, logType string) {
	opts := &slog.HandlerOptions{
		AddSource: debug,
		Level:     levels[level],
	}
	var handler slog.Handler
	if logType == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
