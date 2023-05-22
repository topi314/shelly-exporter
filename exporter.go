package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/exp/slog"
)

type exporter struct {
	cfg   Config
	meter metric.Meter

	plugPower      metric.Float64ObservableGauge
	plugPowerValid metric.Int64ObservableGauge
	overpower      metric.Float64ObservableGauge
	totalPower     metric.Int64ObservableCounter
	temperature    metric.Float64ObservableGauge
	uptime         metric.Int64ObservableGauge
	hasUpdate      metric.Int64ObservableGauge
}

func newExporter(cfg Config, meter metric.Meter) error {
	e := &exporter{
		cfg:   cfg,
		meter: meter,
	}
	var err error

	if e.plugPower, err = meter.Float64ObservableGauge("shellyplug_power",
		metric.WithDescription("Current real AC power being drawn, in Watts"),
		metric.WithUnit("W"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_power metric: %w", err)
	}

	if e.plugPowerValid, err = meter.Int64ObservableGauge("shellyplug_power_valid",
		metric.WithDescription("Whether power metering self-checks OK"),
		metric.WithUnit("1"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_power_valid metric: %w", err)
	}

	if e.overpower, err = meter.Float64ObservableGauge("shellyplug_overpower",
		metric.WithDescription("Value in Watts, on which an overpower condition is detected"),
		metric.WithUnit("W"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_overpower metric: %w", err)
	}

	if e.totalPower, err = meter.Int64ObservableCounter("shellyplug_total_power",
		metric.WithDescription("Total energy consumed by the attached electrical appliance in Watt-minute"),
		metric.WithUnit("Wmin"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_total_power metric: %w", err)
	}

	if e.temperature, err = meter.Float64ObservableGauge("shellyplug_temperature",
		metric.WithDescription("PlugS only internal device temperature in °C"),
		metric.WithUnit("°C"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_temperature metric: %w", err)
	}

	if e.uptime, err = meter.Int64ObservableGauge("shellyplug_uptime",
		metric.WithDescription("Seconds elapsed since boot"),
		metric.WithUnit("s"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_uptime metric: %w", err)
	}

	if e.hasUpdate, err = meter.Int64ObservableGauge("shellyplug_has_update",
		metric.WithDescription("Whether an update is available"),
		metric.WithUnit("1"),
	); err != nil {
		return fmt.Errorf("failed to create shellyplug_has_update metric: %w", err)
	}

	_, err = meter.RegisterCallback(e.collect, e.plugPower, e.plugPowerValid, e.overpower, e.totalPower, e.temperature, e.uptime, e.hasUpdate)
	if err != nil {
		return fmt.Errorf("failed to register callback: %w", err)
	}

	return nil
}

func (e *exporter) collect(ctx context.Context, o metric.Observer) error {
	slog.DebugCtx(ctx, "Collecting data for plugs")
	defer slog.DebugCtx(ctx, "Collected data for plug")
	var (
		errs error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)
	for i := range e.cfg.Configs {
		wg.Add(1)
		cfg := e.cfg.Configs[i]
		go func() {
			defer wg.Done()
			status, err := getPlugStatus(ctx, cfg)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = errors.Join(errs, err)
				return
			}
			slog.DebugCtx(ctx, "Got plug status", slog.Any("status", status))

			attrs := []attribute.KeyValue{
				attribute.String("serial", strconv.Itoa(status.Serial)),
				attribute.String("name", cfg.Name),
			}

			o.ObserveFloat64(e.temperature, status.Temperature, metric.WithAttributes(attrs...))
			o.ObserveInt64(e.uptime, int64(status.Uptime), metric.WithAttributes(attrs...))
			o.ObserveInt64(e.hasUpdate, boolToInt64(status.Update.HasUpdate), metric.WithAttributes(attrs...))

			for ii, meter := range status.Meters {
				attrs = append(attrs, attribute.String("meter", strconv.Itoa(ii)))

				o.ObserveFloat64(e.plugPower, meter.Power, metric.WithAttributes(attrs...))
				o.ObserveInt64(e.plugPowerValid, boolToInt64(meter.IsValid), metric.WithAttributes(attrs...))
				o.ObserveFloat64(e.overpower, meter.Overpower, metric.WithAttributes(attrs...))
				o.ObserveInt64(e.totalPower, int64(meter.Total), metric.WithAttributes(attrs...))
			}
		}()
	}
	wg.Wait()
	return errs
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

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

type PlugStatus struct {
	Serial          int     `json:"serial"`
	Meters          []Meter `json:"meters"`
	Temperature     float64 `json:"temperature"`
	Overtemperature bool    `json:"overtemperature"`
	Update          Update  `json:"update"`
	Uptime          int     `json:"uptime"`
}

type Meter struct {
	Power     float64 `json:"power"`
	IsValid   bool    `json:"is_valid"`
	Overpower float64 `json:"overpower"`
	Total     int     `json:"total"`
}

type Update struct {
	HasUpdate bool `json:"has_update"`
}
