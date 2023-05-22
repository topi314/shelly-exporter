package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"golang.org/x/exp/slog"
)

func newMeter(cfg Config) (metric.Meter, func(), error) {
	exp, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exp),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(Name),
			semconv.ServiceNamespace(Namespace),
			semconv.ServiceInstanceID(cfg.Otel.InstanceID),
			semconv.ServiceVersion(Version),
		)),
	)
	global.SetMeterProvider(mp)

	server := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: promhttp.Handler(),
	}
	go func() {
		if listenErr := server.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			slog.Error("failed to listen metrics server", slog.Any("err", listenErr))
		}
	}()

	return mp.Meter(Name), func() {
		if closeErr := server.Close(); closeErr != nil {
			slog.Error("failed to close metrics server", slog.Any("err", closeErr))
		}
	}, nil
}
