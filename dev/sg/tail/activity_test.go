package tail

import (
	"fmt"
	"testing"
)

func TestActivityParse(t *testing.T) {
	tests := []struct {
		raw         string
		wantName    string
		wantLevel   string
		wantMessage string
	}{
		{
			raw:         `otel-collector: 2023-11-19T09:50:38.408Z        info    service/telemetry.go:104        Serving Prometheus metrics      {"address": ":8888", "level": "Basic"}`,
			wantName:    "otel-collector",
			wantLevel:   "INFO",
			wantMessage: `service/telemetry.go:104        Serving Prometheus metrics      {"address": ":8888", "level": "Basic"}`,
		},
		{
			raw:         `searcher: 2023/11/19 15:36:46 tmpfriend: Removing /tmp/.searcher.tmp/tmpfriend-87880-3133534317`,
			wantName:    `searcher`,
			wantLevel:   ``,
			wantMessage: `tmpfriend: Removing /tmp/.searcher.tmp/tmpfriend-87880-3133534317`,
		},
		{
			raw:         `telemetry-gateway: INFO telemetry-gateway.tracing shared/tracing.go:48 initializing OTLP exporter`,
			wantName:    "telemetry-gateway",
			wantLevel:   "INFO",
			wantMessage: "telemetry-gateway.tracing shared/tracing.go:48 initializing OTLP exporter",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("raw_%d", i), func(t *testing.T) {
			a := parseActivity(tt.raw)
			if a.name != tt.wantName {
				t.Errorf("got name %q, want %q", a.name, tt.wantName)
			}
			if a.level != tt.wantLevel {
				t.Errorf("got level %q, want %q", a.level, tt.wantLevel)
			}
			if a.data != tt.wantMessage {
				t.Errorf("got data %q, want %q", a.data, tt.wantMessage)
			}
		})
	}
}
