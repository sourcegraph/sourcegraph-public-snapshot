package telemetry

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/hexops/autogold"
	"github.com/hexops/valast"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestInitializeJob(t *testing.T) {
	confClient = conf.MockClient()
	defer func() {
		confClient = conf.DefaultClient()
	}()

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

			if have, want := isEnabled(), test.shouldInit; have != want {
				t.Errorf("unexpected isEnabled return value have=%t want=%t", have, want)
			}
		})
	}
}

func TestHandlerEnabledDisabled(t *testing.T) {
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

			handler := newTelemetryHandler(logtest.Scoped(t), database.NewMockEventLogStore(), func(ctx context.Context, event []*types.Event) error {
				return nil
			})
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

func TestHandlerLoadsEvents(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(logger, t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	confClient.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExportUsageTelemetry: &schema.ExportUsageTelemetry{Enabled: true}}})

	t.Run("loads no events when table is empty", func(t *testing.T) {
		handler := newTelemetryHandler(logtest.Scoped(t), db.EventLogs(), func(ctx context.Context, event []*types.Event) error {
			if len(event) != 0 {
				t.Errorf("expected empty events but got event array with size: %d", len(event))
			}
			return nil
		})

		err := handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

	want := []*database.Event{
		{
			Name:   "event1",
			UserID: 1,
			Source: "test",
		},
		{
			Name:   "event2",
			UserID: 2,
			Source: "test",
		},
	}
	err := db.EventLogs().BulkInsert(ctx, want)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("loads events without error", func(t *testing.T) {
		var got []*types.Event
		handler := newTelemetryHandler(logtest.Scoped(t), db.EventLogs(), func(ctx context.Context, event []*types.Event) error {
			got = event
			return nil
		})

		err := handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("loads events without error", []*types.Event{
			{
				ID:       1,
				Name:     "event1",
				UserID:   valast.Addr(int32(1)).(*int32),
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			},
			{
				ID:       2,
				Name:     "event2",
				UserID:   valast.Addr(int32(2)).(*int32),
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			},
		}).Equal(t, got)
	})

	t.Run("loads using specified batch size from settings", func(t *testing.T) {
		confClient.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExportUsageTelemetry: &schema.ExportUsageTelemetry{Enabled: true, BatchSize: 1}}})

		var got []*types.Event
		handler := newTelemetryHandler(logtest.Scoped(t), db.EventLogs(), func(ctx context.Context, event []*types.Event) error {
			got = event
			return nil
		})
		err := handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("loads using specified batch size from settings", []*types.Event{
			{
				ID:       1,
				Name:     "event1",
				UserID:   valast.Addr(int32(1)).(*int32),
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			},
		}).Equal(t, got)
	})
}
