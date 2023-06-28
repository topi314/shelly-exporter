package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
)

func startCollector(ctx context.Context, cfg Config) {
	for i := range cfg.Configs {
		config := cfg.Configs[i]
		if config.Interval == 0 {
			config.Interval = cfg.Global.ScrapeInterval
		}
		if config.Timeout == 0 {
			config.Timeout = cfg.Global.ScrapeTimeout
		}
		go func() {
			logger := slog.With(
				slog.String("name", config.Name),
				slog.String("address", config.Address),
				slog.Bool("insecure", config.Insecure),
				slog.String("username", config.Username),
				slog.String("password", config.Password),
				slog.Duration("interval", config.Interval),
				slog.Duration("timeout", config.Timeout),
			)
			collect(ctx, logger, config)
		}()
	}
}

func collect(ctx context.Context, logger *slog.Logger, cfg PlugConfig) {
	timer := time.NewTicker(cfg.Interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			logger.DebugCtx(ctx, "Collecting data for plug")
			status, err := getPlugStatus(ctx, cfg)
			if err != nil {
				logger.Error("Failed to get plug status", slog.Any("err", err))
				continue
			}
			logger.DebugCtx(ctx, "Got plug status", slog.Any("status", status))

			labels := prometheus.Labels{
				"name": cfg.Name,
			}
			temperature.With(labels).Set(status.Temperature)
			uptime.With(labels).Set(float64(status.Uptime))
			hasUpdate.With(labels).Set(boolToFloat64(status.Update.HasUpdate))

			for i, meter := range status.Meters {
				labels["meter"] = strconv.Itoa(i)
				power.With(labels).Set(meter.Power)
				powerValid.With(labels).Set(boolToFloat64(meter.IsValid))
				overpower.With(labels).Set(meter.Overpower)
				totalPower.With(labels).Add(float64(meter.Total))
			}
			logger.DebugCtx(ctx, "Collected data for plug")
		}
	}
}

func getPlugStatus(ctx context.Context, cfg PlugConfig) (*PlugStatus, error) {
	scheme := "https"
	if cfg.Insecure {
		scheme = "http"
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s://%s/status", scheme, cfg.Address), nil)
	if err != nil {
		return nil, err
	}
	if cfg.Username != "" {
		rq.SetBasicAuth(cfg.Username, cfg.Password)
	}

	rs, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	var status PlugStatus
	if err = json.NewDecoder(rs.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
