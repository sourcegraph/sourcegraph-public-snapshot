package telemetry

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestInitializeJob(t *testing.T) {
	ctx := context.Background()

	confClient = conf.MockClient()

	tests := []struct {
		name         string
		mockedConfig schema.SiteConfiguration
		shouldInit   bool
	}{
		{
			name:         "missing setting",
			mockedConfig: schema.SiteConfiguration{},
			shouldInit:   false,
		},
		{
			name: "setting exists but enabled field missing",
			mockedConfig: schema.SiteConfiguration{
				ExportUsageTelemetry: &schema.ExportUsageTelemetry{},
			},
			shouldInit: false,
		},
		{
			name: "setting exists but enabled field set false",
			mockedConfig: schema.SiteConfiguration{
				ExportUsageTelemetry: &schema.ExportUsageTelemetry{Enabled: false},
			},
			shouldInit: false,
		},
		{
			name: "setting exists and is enabled",
			mockedConfig: schema.SiteConfiguration{
				ExportUsageTelemetry: &schema.ExportUsageTelemetry{Enabled: true},
			},
			shouldInit: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confClient.Mock(&conf.Unified{SiteConfiguration: test.mockedConfig})

			job := NewTelemetryJob()
			routines, err := job.Routines(ctx, logtest.Scoped(t))
			if err != nil {
				t.Error(err)
			}

			if test.shouldInit {
				if len(routines) != 1 {
					t.Error("expected one routine")
				}
			} else {
				if len(routines) != 0 {
					t.Error("expected no routines")
				}
			}
		})
	}
}

func TestHandler(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		mockedConfig schema.SiteConfiguration
		expectErr    error
	}{
		{
			name:         "missing setting",
			mockedConfig: schema.SiteConfiguration{},
			expectErr:    disabledErr,
		},
		{
			name: "setting exists but enabled field missing",
			mockedConfig: schema.SiteConfiguration{
				ExportUsageTelemetry: &schema.ExportUsageTelemetry{},
			},
			expectErr: disabledErr,
		},
		{
			name: "setting exists but enabled field set false",
			mockedConfig: schema.SiteConfiguration{
				ExportUsageTelemetry: &schema.ExportUsageTelemetry{Enabled: false},
			},
			expectErr: disabledErr,
		},
		{
			name: "setting exists and is enabled",
			mockedConfig: schema.SiteConfiguration{
				ExportUsageTelemetry: &schema.ExportUsageTelemetry{Enabled: true},
			},
			expectErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confClient.Mock(&conf.Unified{SiteConfiguration: test.mockedConfig})

			handler := telemetryHandler{logger: logtest.Scoped(t)}
			err := handler.Handle(ctx)
			if err != nil {
				if !errors.Is(err, disabledErr) {
					t.Error("unexpected error from Handle function, expected disabled error")
				}
			} else {
				if test.expectErr != nil {
					t.Error("expected error but did not receive one")
				}
			}
		})
	}
}
