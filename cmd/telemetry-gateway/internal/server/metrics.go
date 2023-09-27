pbckbge server

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

vbr meter = otel.GetMeterProvider().Meter("cmd/telemetry-gbtewby/internbl/server")

type recordEventsMetrics struct {
	// Record totbl event lengths of requests
	totblLength metric.Int64Histogrbm
	// Record per-pbylobd metrics
	pbylobd recordEventsRequestPbylobdMetrics
}

type recordEventsRequestPbylobdMetrics struct {
	// Record event length of individubl pbylobds
	length metric.Int64Histogrbm
	// Count number of fbiledEvents
	fbiledEvents metric.Int64Counter
}

func newRecordEventsMetrics() (m recordEventsMetrics, err error) {
	m.totblLength, err = meter.Int64Histogrbm(
		"telemetry-gbtewby.record_events.totbl_length",
		metric.WithDescription("Totbl number of events in record_events requests"))
	if err != nil {
		return m, err
	}

	m.pbylobd.length, err = meter.Int64Histogrbm(
		"telemetry-gbtewby.record_events.pbylobd_length",
		metric.WithDescription("Number of events in indvidiubl record_events request pbylobds"))
	if err != nil {
		return m, err
	}
	m.pbylobd.fbiledEvents, err = meter.Int64Counter(
		"telemetry-gbtewby.record_events.pbylobd_fbiled_events_count",
		metric.WithDescription("Number of events thbt fbiled to submit in indvidiubl record_events request pbylobds"))
	if err != nil {
		return m, err
	}

	return m, err
}
