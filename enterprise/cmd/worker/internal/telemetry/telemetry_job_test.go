package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"

	"github.com/google/go-cmp/cmp"

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

			handler := mockTelemetryHandler(t, func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
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

	initAllowedEvents(t, db, []string{"event1", "event2"})

	t.Run("loads no events when table is empty", func(t *testing.T) {
		handler := mockTelemetryHandler(t, func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
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
		handler := mockTelemetryHandler(t, func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
			got = event
			return nil
		})
		handler.eventLogStore = db.EventLogs()

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
		handler := mockTelemetryHandler(t, func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
			got = event
			return nil
		})
		handler.eventLogStore = db.EventLogs()
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

func TestHandlerLoadsEventsWithBookmarkState(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(logger, t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	initAllowedEvents(t, db, []string{"event1", "event2", "event4"})
	testData := []*database.Event{
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
	err := db.EventLogs().BulkInsert(ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
	err = basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrape_state (bookmark_id) values (0);"))
	if err != nil {
		t.Error(err)
	}

	config := validEnabledConfiguration()
	config.ExportUsageTelemetry.BatchSize = 1
	confClient.Mock(&conf.Unified{SiteConfiguration: config})

	handler := mockTelemetryHandler(t, noopHandler())
	handler.eventLogStore = db.EventLogs() // replace mocks with real stores for a partially mocked handler
	handler.bookmarkStore = newBookmarkStore(db)

	t.Run("first execution of handler should return first event", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, got []*types.Event, config topicConfig, metadata instanceMetadata) error {
			autogold.Want("first execution of handler should return first event", []*types.Event{{
				ID:       1,
				Name:     "event1",
				UserID:   1,
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			}}).Equal(t, got)
			return nil
		}

		err = handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("second execution of handler should return second event", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, got []*types.Event, config topicConfig, metadata instanceMetadata) error {
			autogold.Want("second execution of handler should return second event", []*types.Event{{
				ID:       2,
				Name:     "event2",
				UserID:   2,
				Argument: "{}",
				Source:   "test",
				Version:  "0.0.0+dev",
			}}).Equal(t, got)
			return nil
		}

		err = handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("third execution of handler should return no events", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, event []*types.Event, config topicConfig, metadata instanceMetadata) error {
			if len(event) == 0 {
				t.Error("expected empty events")
			}
			return nil
		}

		err = handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestHandlerLoadsEventsWithAllowlist(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(logger, t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	initAllowedEvents(t, db, []string{"allowed"})
	testData := []*database.Event{
		{
			Name:   "allowed",
			UserID: 1,
			Source: "test",
		},
		{
			Name:   "not-allowed",
			UserID: 2,
			Source: "test",
		},
		{
			Name:   "allowed",
			UserID: 3,
			Source: "test",
		},
	}
	err := db.EventLogs().BulkInsert(ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
	err = basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrape_state (bookmark_id) values (0);"))
	if err != nil {
		t.Error(err)
	}

	config := validEnabledConfiguration()
	confClient.Mock(&conf.Unified{SiteConfiguration: config})

	handler := mockTelemetryHandler(t, noopHandler())
	handler.eventLogStore = db.EventLogs() // replace mocks with real stores for a partially mocked handler
	handler.bookmarkStore = newBookmarkStore(db)

	t.Run("ensure only allowed events are returned", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, got []*types.Event, config topicConfig, metadata instanceMetadata) error {
			autogold.Want("first execution of handler should return first event", []*types.Event{
				{
					ID:       1,
					Name:     "allowed",
					UserID:   1,
					Argument: "{}",
					Source:   "test",
					Version:  "0.0.0+dev",
				},
				{
					ID:       3,
					Name:     "allowed",
					UserID:   3,
					Argument: "{}",
					Source:   "test",
					Version:  "0.0.0+dev",
				},
			}).Equal(t, got)
			return nil
		}

		err = handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
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
	bookmarkStore := newBookmarkStore(db)

	confClient.Mock(&conf.Unified{SiteConfiguration: validEnabledConfiguration()})

	t.Run("handle fails when missing project name", func(t *testing.T) {
		config := validEnabledConfiguration()
		config.ExportUsageTelemetry.TopicProjectName = ""
		confClient.Mock(&conf.Unified{SiteConfiguration: config})

		handler := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), bookmarkStore, noopHandler())
		err := handler.Handle(ctx)

		autogold.Want("handle fails when missing project name", "getTopicConfig: missing project name to export usage data").Equal(t, err.Error())
	})
	t.Run("handle fails when missing topic name", func(t *testing.T) {
		config := validEnabledConfiguration()
		config.ExportUsageTelemetry.TopicName = ""
		confClient.Mock(&conf.Unified{SiteConfiguration: config})

		handler := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), bookmarkStore, noopHandler())
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

func TestGetBookmark(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(logger, t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)
	store := newBookmarkStore(db)
	eventLogStore := db.EventLogs()

	clearStateTable := func() {
		dbHandle.Exec("DELETE FROM event_logs_scrape_state;")
	}

	insert := []*database.Event{
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
	err := eventLogStore.BulkInsert(ctx, insert)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("state is empty should generate row", func(t *testing.T) {
		got, err := store.GetBookmark(ctx)
		if err != nil {
			t.Error(err)
		}
		want := 2
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s (want/got): %s", t.Name(), diff)
		}
		clearStateTable()
	})

	t.Run("state exists and returns bookmark", func(t *testing.T) {
		err := basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrape_state (bookmark_id) values (1);"))
		if err != nil {
			t.Error(err)
		}

		got, err := store.GetBookmark(ctx)
		if err != nil {
			t.Error(err)
		}
		want := 1
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s (want/got): %s", t.Name(), diff)
		}
		clearStateTable()
	})
}

func TestUpdateBookmark(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(logger, t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)
	store := newBookmarkStore(db)

	err := basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrape_state (bookmark_id) values (1);"))
	if err != nil {
		t.Error(err)
	}

	want := 6
	err = store.UpdateBookmark(ctx, want)
	if err != nil {
		t.Error(errors.Wrap(err, "UpdateBookmark"))
	}

	got, err := store.GetBookmark(ctx)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%s (want/got): %s", t.Name(), diff)
	}
}

func mockTelemetryHandler(t *testing.T, callbackFunc sendEventsCallbackFunc) *telemetryHandler {
	bms := NewMockBookmarkStore()
	bms.GetBookmarkFunc.SetDefaultReturn(0, nil)

	return &telemetryHandler{
		logger:             logtest.Scoped(t),
		eventLogStore:      database.NewMockEventLogStore(),
		globalStateStore:   database.NewMockGlobalStateStore(),
		userEmailsStore:    database.NewMockUserEmailsStore(),
		bookmarkStore:      bms,
		sendEventsCallback: callbackFunc,
	}
}

// initAllowedEvents is a helper to establish a deterministic set of allowed events. This is useful because
// the standard database migrations will create data in the allowed events table that may conflict with tests.
func initAllowedEvents(t *testing.T, db database.DB, names []string) {
	store := basestore.NewWithHandle(db.Handle())
	err := store.Exec(context.Background(), sqlf.Sprintf("delete from event_logs_export_allowlist"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Exec(context.Background(), sqlf.Sprintf("insert into event_logs_export_allowlist (event_name) values (unnest(%s::text[]))", pq.Array(names)))
	if err != nil {
		t.Fatal(err)
	}
}
