package metrics

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const instrumName = "opentelemetry/otel"

type config struct {
	tracerProvider trace.TracerProvider
	tracer         trace.Tracer

	meterProvider metric.MeterProvider
	meter         metric.Meter

	opts []metric.ObserveOption
}

func newConfig() *config {
	c := &config{
		tracerProvider: otel.GetTracerProvider(),
		meterProvider:  otel.GetMeterProvider(),
		tracer:         nil,
		meter:          nil,
		opts:           nil,
	}
	return c
}

// ReportDBStatsMetrics reports DBStats metrics using OpenTelemetry Metrics API.
func ReportDBStatsMetrics(db *sql.DB) {
	cfg := newConfig()

	if cfg.meter == nil {
		cfg.meter = cfg.meterProvider.Meter(instrumName)
	}

	meter := cfg.meter
	opts := cfg.opts

	maxOpenConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_max_open",
		metric.WithDescription("Maximum number of open connections to the database"),
	)
	openConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_open",
		metric.WithDescription("The number of established connections both in use and idle"),
	)
	inUseConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_in_use",
		metric.WithDescription("The number of connections currently in use"),
	)
	idleConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_idle",
		metric.WithDescription("The number of idle connections"),
	)
	connsWaitCount, _ := meter.Int64ObservableCounter(
		"go.sql.connections_wait_count",
		metric.WithDescription("The total number of connections waited for"),
	)
	connsWaitDuration, _ := meter.Int64ObservableCounter(
		"go.sql.connections_wait_duration",
		metric.WithDescription("The total time blocked waiting for a new connection"),
		metric.WithUnit("nanoseconds"),
	)
	connsClosedMaxIdle, _ := meter.Int64ObservableCounter(
		"go.sql.connections_closed_max_idle",
		metric.WithDescription("The total number of connections closed due to SetMaxIdleConns"),
	)
	connsClosedMaxIdleTime, _ := meter.Int64ObservableCounter(
		"go.sql.connections_closed_max_idle_time",
		metric.WithDescription("The total number of connections closed due to SetConnMaxIdleTime"),
	)
	connsClosedMaxLifetime, _ := meter.Int64ObservableCounter(
		"go.sql.connections_closed_max_lifetime",
		metric.WithDescription("The total number of connections closed due to SetConnMaxLifetime"),
	)

	_, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats := db.Stats()

			o.ObserveInt64(maxOpenConns, int64(stats.MaxOpenConnections), opts...)
			o.ObserveInt64(openConns, int64(stats.OpenConnections), opts...)
			o.ObserveInt64(inUseConns, int64(stats.InUse), opts...)
			o.ObserveInt64(idleConns, int64(stats.Idle), opts...)
			o.ObserveInt64(connsWaitCount, stats.WaitCount, opts...)
			o.ObserveInt64(connsWaitDuration, int64(stats.WaitDuration), opts...)
			o.ObserveInt64(connsClosedMaxIdle, stats.MaxIdleClosed, opts...)
			o.ObserveInt64(connsClosedMaxIdleTime, stats.MaxIdleTimeClosed, opts...)
			o.ObserveInt64(connsClosedMaxLifetime, stats.MaxLifetimeClosed, opts...)
			return nil
		},
		maxOpenConns,
		openConns,
		inUseConns,
		idleConns,
		connsWaitCount,
		connsWaitDuration,
		connsClosedMaxIdle,
		connsClosedMaxIdleTime,
		connsClosedMaxLifetime,
	)
	if err != nil {
		panic(err)
	}
}
