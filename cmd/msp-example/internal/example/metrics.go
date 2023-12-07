package example

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("msp-example")

func getRequestCounter() (metric.Int64Counter, error) {
	return meter.Int64Counter("request_count")
}
