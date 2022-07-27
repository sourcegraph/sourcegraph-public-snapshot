package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"

	"github.com/sourcegraph/sourcegraph/internal/version"

	"github.com/sourcegraph/log/logtest"

	"github.com/hexops/autogold"
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
			name:         "setting exists and is enabled",
			mockedConfig: validEnabledConfiguration(),
			shouldInit:   true,
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
			name:         "setting exists and is enabled",
			mockedConfig: validEnabledConfiguration(),
			expectErr:    nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confClient.Mock(&conf.Unified{SiteConfiguration: test.mockedConfig})

			handler := newTelemetryHandler(logtest.Scoped(t), database.NewMockEventLogStore(), database.NewMockUserEmailsStore(), database.NewMockGlobalStateStore(), func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
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

	confClient.Mock(&conf.Unified{SiteConfiguration: validEnabledConfiguration()})

	t.Run("loads no events when table is empty", func(t *testing.T) {
		handler := newTelemetryHandler(logtest.Scoped(t), db.EventLogs(), db.UserEmails(), db.GlobalState(), func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
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
		handler := newTelemetryHandler(logtest.Scoped(t), db.EventLogs(), db.UserEmails(), db.GlobalState(), func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
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
				UserID:   1,
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			},
			{
				ID:       2,
				Name:     "event2",
				UserID:   2,
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			},
		}).Equal(t, got)
	})

	t.Run("loads using specified batch size from settings", func(t *testing.T) {
		config := validEnabledConfiguration()
		config.ExportUsageTelemetry.BatchSize = 1
		confClient.Mock(&conf.Unified{SiteConfiguration: config})

		var got []*types.Event
		handler := newTelemetryHandler(logtest.Scoped(t), db.EventLogs(), db.UserEmails(), db.GlobalState(), func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
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
				UserID:   1,
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			},
		}).Equal(t, got)
	})
}

func validEnabledConfiguration() schema.SiteConfiguration {
	return schema.SiteConfiguration{ExportUsageTelemetry: &schema.ExportUsageTelemetry{
		Enabled:          true,
		TopicName:        "test-topic",
		TopicProjectName: "test-project",
	}}
}

func TestHandleInvalidConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(logger, t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	confClient.Mock(&conf.Unified{SiteConfiguration: validEnabledConfiguration()})

	t.Run("handle fails when missing project name", func(t *testing.T) {
		config := validEnabledConfiguration()
		config.ExportUsageTelemetry.TopicProjectName = ""
		confClient.Mock(&conf.Unified{SiteConfiguration: config})

		handler := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), noopHandler())
		err := handler.Handle(ctx)

		autogold.Want("handle fails when missing project name", "getTopicConfig: missing project name to export usage data").Equal(t, err.Error())
	})
	t.Run("handle fails when missing topic name", func(t *testing.T) {
		config := validEnabledConfiguration()
		config.ExportUsageTelemetry.TopicName = ""
		confClient.Mock(&conf.Unified{SiteConfiguration: config})

		handler := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), noopHandler())
		err := handler.Handle(ctx)

		autogold.Want("handle fails when missing topic name", "getTopicConfig: missing topic name to export usage data").Equal(t, err.Error())
	})
}

func TestBuildBigQueryObject(t *testing.T) {
	atTime := time.Date(2022, 7, 22, 0, 0, 0, 0, time.UTC)
	event := &types.Event{
		ID:              1,
		Name:            "GREAT_EVENT",
		URL:             "https://sourcegraph.com/search",
		UserID:          5,
		AnonymousUserID: "anonymous",
		Argument:        "argument",
		Source:          "src",
		Version:         "1.1.1",
		Timestamp:       atTime,
	}

	metadata := &instanceMetadata{
		DeployType:        "docker",
		Version:           "1.2.3",
		SiteID:            "site-id-1",
		LicenseKey:        "license-key-1",
		InitialAdminEmail: "admin@place.com",
	}

	got := buildBigQueryObject(event, metadata)
	autogold.Want("build big query object", &bigQueryEvent{
		SiteID:            "site-id-1",
		LicenseKey:        "license-key-1",
		InitialAdminEmail: "admin@place.com",
		DeployType:        "docker",
		EventName:         "GREAT_EVENT",
		AnonymousUserID:   "anonymous",
		UserID:            5,
		Source:            "src",
		Timestamp:         "2022-07-22T00:00:00Z",
		Version:           "1.1.1",
		PublicArgument:    "argument",
	}).Equal(t, got)
}

func TestGetInstanceMetadata(t *testing.T) {
	ctx := context.Background()

	stateStore := database.NewMockGlobalStateStore()
	userEmailStore := database.NewMockUserEmailsStore()
	version.Mock("fake-Version-1")
	confClient.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{LicenseKey: "mock-license"}})
	deploy.Mock("fake-deploy-type")

	stateStore.GetFunc.SetDefaultReturn(database.GlobalState{
		SiteID:      "fake-site-id",
		Initialized: true,
	}, nil)

	userEmailStore.GetInitialSiteAdminInfoFunc.SetDefaultReturn("fake@place.com", true, nil)

	got, err := getInstanceMetadata(ctx, stateStore, userEmailStore)
	if err != nil {
		t.Fatal(err)
	}

	autogold.Want("check that instance metadata equals mocked values", instanceMetadata{
		DeployType:        "fake-deploy-type",
		Version:           "fake-Version-1",
		SiteID:            "fake-site-id",
		LicenseKey:        "mock-license",
		InitialAdminEmail: "fake@place.com",
	}).Equal(t, got)
}

func noopHandler() sendEventsCallbackFunc {
	return func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
		return nil
	}
}
