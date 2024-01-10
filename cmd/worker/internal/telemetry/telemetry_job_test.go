package telemetry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"

	"github.com/keegancsmith/sqlf"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"

	"github.com/sourcegraph/sourcegraph/internal/version"

	"github.com/sourcegraph/log/logtest"

	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"

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
		name       string
		setting    bool
		shouldInit bool
	}{
		{
			name:       "job set disabled",
			setting:    false,
			shouldInit: false,
		},
		{
			name:       "setting exists and is enabled",
			setting:    true,
			shouldInit: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockEnvVars(t, test.setting)
			if have, want := isEnabled(), test.shouldInit; have != want {
				t.Errorf("unexpected isEnabled return value have=%t want=%t", have, want)
			}
		})
	}
}

func TestHandlerEnabledDisabled(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setting   bool
		expectErr error
	}{
		{
			name:      "job set disabled",
			setting:   false,
			expectErr: disabledErr,
		},
		{
			name:      "setting exists and is enabled",
			setting:   true,
			expectErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confClient.Mock(&conf.Unified{SiteConfiguration: validConfiguration()})
			mockEnvVars(t, test.setting)
			handler := mockTelemetryHandler(t, func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error {
				return nil
			})
			err := handler.Handle(ctx)
			if !errors.Is(err, test.expectErr) {
				t.Errorf("unexpected error from Handle function, expected error: %v, received: %s", test.expectErr, err.Error())
			}
		})
	}
}

func TestHandlerLoadsEvents(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	confClient.Mock(&conf.Unified{SiteConfiguration: validConfiguration()})
	mockEnvVars(t, true)

	initAllowedEvents(t, db, []string{"event1", "event2"})

	t.Run("loads no events when table is empty", func(t *testing.T) {
		handler := mockTelemetryHandler(t, func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error {
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
	flags := make(map[string]bool)
	flags["testflag"] = true

	want := []*database.Event{
		{
			Name:             "event1",
			UserID:           1,
			Source:           "test",
			EvaluatedFlagSet: flags,
			DeviceID:         pointers.Ptr("device-1"),
			InsertID:         pointers.Ptr("insert-1"),
		},
		{
			Name:     "event2",
			UserID:   2,
			Source:   "test",
			DeviceID: pointers.Ptr("device-2"),
			InsertID: pointers.Ptr("insert-2"),
		},
	}
	//lint:ignore SA1019 existing usage of deprecated functionality. This telemetry worker will be removed entirely when all events are migrated to V2.
	err := db.EventLogs().BulkInsert(ctx, want)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("loads events without error", func(t *testing.T) {
		var got []*database.Event
		handler := mockTelemetryHandler(t, func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error {
			got = event
			return nil
		})
		handler.eventLogStore = db.EventLogs()

		err := handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*database.Event{
			{
				ID:     1,
				Name:   "event1",
				UserID: 1,
				Argument: json.RawMessage{
					123,
					125,
				},
				PublicArgument: json.RawMessage{
					123,
					125,
				},
				Source:           "test",
				Version:          "0.0.0+dev",
				EvaluatedFlagSet: featureflag.EvaluatedFlagSet{"testflag": true},
				DeviceID:         valast.Addr("device-1").(*string),
				InsertID:         valast.Addr("insert-1").(*string),
			},
			{
				ID:     2,
				Name:   "event2",
				UserID: 2,
				Argument: json.RawMessage{
					123,
					125,
				},
				PublicArgument: json.RawMessage{
					123,
					125,
				},
				Source:   "test",
				Version:  "0.0.0+dev",
				DeviceID: valast.Addr("device-2").(*string),
				InsertID: valast.Addr("insert-2").(*string),
			},
		}).Equal(t, got)
	})

	t.Run("loads using specified batch size from settings", func(t *testing.T) {
		config := validConfiguration()
		config.ExportUsageTelemetry.BatchSize = 1
		confClient.Mock(&conf.Unified{SiteConfiguration: config})

		var got []*database.Event
		handler := mockTelemetryHandler(t, func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error {
			got = event
			return nil
		})
		handler.eventLogStore = db.EventLogs()
		err := handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*database.Event{{
			ID:     1,
			Name:   "event1",
			UserID: 1,
			Argument: json.RawMessage{
				123,
				125,
			},
			PublicArgument: json.RawMessage{
				123,
				125,
			},
			Source:           "test",
			Version:          "0.0.0+dev",
			EvaluatedFlagSet: featureflag.EvaluatedFlagSet{"testflag": true},
			DeviceID:         valast.Addr("device-1").(*string),
			InsertID:         valast.Addr("insert-1").(*string),
		}}).Equal(t, got)
	})
}

func TestHandlerLoadsEventsWithBookmarkState(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	initAllowedEvents(t, db, []string{"event1", "event2", "event4"})
	testData := []*database.Event{
		{
			Name:     "event1",
			UserID:   1,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
		{
			Name:     "event2",
			UserID:   2,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
	}

	//lint:ignore SA1019 existing usage of deprecated functionality. This telemetry worker will be removed entirely when all events are migrated to V2.
	err := db.EventLogs().BulkInsert(ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
	err = basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrape_state (bookmark_id) values (0);"))
	if err != nil {
		t.Error(err)
	}

	config := validConfiguration()
	config.ExportUsageTelemetry.BatchSize = 1
	confClient.Mock(&conf.Unified{SiteConfiguration: config})
	mockEnvVars(t, true)

	handler := mockTelemetryHandler(t, noopHandler())
	handler.eventLogStore = db.EventLogs() // replace mocks with real stores for a partially mocked handler
	handler.bookmarkStore = newBookmarkStore(db)

	t.Run("first execution of handler should return first event", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, got []*database.Event, config topicConfig, metadata instanceMetadata) error {
			autogold.Expect([]*database.Event{{
				ID:     1,
				Name:   "event1",
				UserID: 1,
				Argument: json.RawMessage{
					123,
					125,
				},
				PublicArgument: json.RawMessage{
					123,
					125,
				},
				Source:   "test",
				Version:  "0.0.0+dev",
				DeviceID: valast.Addr("device").(*string),
				InsertID: valast.Addr("insert").(*string),
			}}).Equal(t, got)
			return nil
		}

		err = handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("second execution of handler should return second event", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, got []*database.Event, config topicConfig, metadata instanceMetadata) error {
			autogold.Expect([]*database.Event{{
				ID:     2,
				Name:   "event2",
				UserID: 2,
				Argument: json.RawMessage{
					123,
					125,
				},
				PublicArgument: json.RawMessage{
					123,
					125,
				},
				Source:   "test",
				Version:  "0.0.0+dev",
				DeviceID: valast.Addr("device").(*string),
				InsertID: valast.Addr("insert").(*string),
			}}).Equal(t, got)
			return nil
		}

		err = handler.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("third execution of handler should return no events", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error {
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
	dbHandle := dbtest.NewDB(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)

	initAllowedEvents(t, db, []string{"allowed"})
	testData := []*database.Event{
		{
			Name:     "allowed",
			UserID:   1,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
		{
			Name:     "not-allowed",
			UserID:   2,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
		{
			Name:     "allowed",
			UserID:   3,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
	}

	//lint:ignore SA1019 existing usage of deprecated functionality. This telemetry worker will be removed entirely when all events are migrated to V2.
	err := db.EventLogs().BulkInsert(ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
	err = basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrape_state (bookmark_id) values (0);"))
	if err != nil {
		t.Error(err)
	}

	config := validConfiguration()
	confClient.Mock(&conf.Unified{SiteConfiguration: config})
	mockEnvVars(t, true)

	handler := mockTelemetryHandler(t, noopHandler())
	handler.eventLogStore = db.EventLogs() // replace mocks with real stores for a partially mocked handler
	handler.bookmarkStore = newBookmarkStore(db)

	t.Run("ensure only allowed events are returned", func(t *testing.T) {
		handler.sendEventsCallback = func(ctx context.Context, got []*database.Event, config topicConfig, metadata instanceMetadata) error {
			autogold.Expect([]*database.Event{
				{
					ID:     1,
					Name:   "allowed",
					UserID: 1,
					Argument: json.RawMessage{
						123,
						125,
					},
					PublicArgument: json.RawMessage{
						123,
						125,
					},
					Source:   "test",
					Version:  "0.0.0+dev",
					DeviceID: valast.Addr("device").(*string),
					InsertID: valast.Addr("insert").(*string),
				},
				{
					ID:     3,
					Name:   "allowed",
					UserID: 3,
					Argument: json.RawMessage{
						123,
						125,
					},
					PublicArgument: json.RawMessage{
						123,
						125,
					},
					Source:   "test",
					Version:  "0.0.0+dev",
					DeviceID: valast.Addr("device").(*string),
					InsertID: valast.Addr("insert").(*string),
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

func validConfiguration() schema.SiteConfiguration {
	return schema.SiteConfiguration{ExportUsageTelemetry: &schema.ExportUsageTelemetry{}}
}

func TestHandleInvalidConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbHandle)
	bookmarkStore := newBookmarkStore(db)

	confClient.Mock(&conf.Unified{SiteConfiguration: validConfiguration()})
	mockEnvVars(t, true)

	obsContext := observation.TestContextTB(t)

	t.Run("handle fails when missing project name", func(t *testing.T) {
		projectName = ""
		handler := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), bookmarkStore, noopHandler(), newHandlerMetrics(obsContext))
		err := handler.Handle(ctx)

		autogold.Expect("getTopicConfig: missing project name to export usage data").Equal(t, err.Error())
	})
	t.Run("handle fails when missing topic name", func(t *testing.T) {
		topicName = ""
		handler := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), bookmarkStore, noopHandler(), newHandlerMetrics(obsContext))
		err := handler.Handle(ctx)

		autogold.Expect("getTopicConfig: missing topic name to export usage data").Equal(t, err.Error())
	})
}

func TestBuildBigQueryObject(t *testing.T) {
	atTime := time.Date(2022, 7, 22, 0, 0, 0, 0, time.UTC)
	flags := make(featureflag.EvaluatedFlagSet)
	flags["testflag"] = true

	event := &database.Event{
		ID:               1,
		Name:             "GREAT_EVENT",
		URL:              "https://sourcegraph.com/search",
		UserID:           5,
		AnonymousUserID:  "anonymous",
		PublicArgument:   json.RawMessage("public_argument"),
		Source:           "src",
		Version:          "1.1.1",
		Timestamp:        atTime,
		EvaluatedFlagSet: flags,
		CohortID:         pointers.Ptr("cohort1"),
		FirstSourceURL:   pointers.Ptr("first_source_url"),
		LastSourceURL:    pointers.Ptr("last_source_url"),
		Referrer:         pointers.Ptr("reff"),
		DeviceID:         pointers.Ptr("devid"),
		InsertID:         pointers.Ptr("insertid"),
	}

	metadata := &instanceMetadata{
		DeployType:        "docker",
		Version:           "1.2.3",
		SiteID:            "site-id-1",
		LicenseKey:        "license-key-1",
		InitialAdminEmail: "admin@place.com",
	}

	got := buildBigQueryObject(event, metadata)
	autogold.Expect(&bigQueryEvent{
		SiteID: "site-id-1", LicenseKey: "license-key-1",
		InitialAdminEmail: "admin@place.com",
		DeployType:        "docker",
		EventName:         "GREAT_EVENT",
		AnonymousUserID:   "anonymous",
		FirstSourceURL:    "first_source_url",
		LastSourceURL:     "last_source_url",
		UserID:            5,
		Source:            "src",
		Timestamp:         "2022-07-22T00:00:00Z",
		Version:           "1.1.1",
		FeatureFlags:      `{"testflag":true}`,
		CohortID:          valast.Addr("cohort1").(*string),
		Referrer:          "reff",
		PublicArgument:    "public_argument",
		DeviceID:          valast.Addr("devid").(*string),
		InsertID:          valast.Addr("insertid").(*string),
	}).Equal(t, got)
}

func TestGetInstanceMetadata(t *testing.T) {
	ctx := context.Background()

	stateStore := dbmocks.NewMockGlobalStateStore()
	userEmailStore := dbmocks.NewMockUserEmailsStore()
	version.Mock("fake-Version-1")
	confClient.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{LicenseKey: "mock-license"}})
	deploy.Mock("fake-deploy-type")
	mockEnvVars(t, true)

	stateStore.GetFunc.SetDefaultReturn(database.GlobalState{
		SiteID:      "fake-site-id",
		Initialized: true,
	}, nil)

	userEmailStore.GetInitialSiteAdminInfoFunc.SetDefaultReturn("fake@place.com", true, nil)

	got, err := getInstanceMetadata(ctx, stateStore, userEmailStore)
	if err != nil {
		t.Fatal(err)
	}

	autogold.Expect(instanceMetadata{
		DeployType:        "fake-deploy-type",
		Version:           "fake-Version-1",
		SiteID:            "fake-site-id",
		LicenseKey:        "mock-license",
		InitialAdminEmail: "fake@place.com",
	}).Equal(t, got)
}

func noopHandler() sendEventsCallbackFunc {
	return func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error {
		return nil
	}
}

func Test_getBatchSize(t *testing.T) {
	tests := []struct {
		name   string
		config *conf.Unified
		want   int
	}{
		{
			name:   "null config object",
			config: nil,
			want:   MaxEventsCountDefault,
		},
		{
			name:   "null inner config object",
			config: &conf.Unified{},
			want:   MaxEventsCountDefault,
		},
		{
			name:   "null export data object",
			config: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want:   MaxEventsCountDefault,
		},
		{
			name:   "no batch size specified",
			config: &conf.Unified{SiteConfiguration: validConfiguration()},
			want:   MaxEventsCountDefault,
		},
		{
			name:   "override batch size",
			config: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExportUsageTelemetry: &schema.ExportUsageTelemetry{BatchSize: 5}}},
			want:   5,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confClient.Mock(test.config)
			got := getBatchSize()
			if got != test.want {
				t.Errorf("unexpected batch size want:%d, got:%d", test.want, got)
			}
		})
	}
}

func TestGetBookmark(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHandle := dbtest.NewDB(t)
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

	//lint:ignore SA1019 existing usage of deprecated functionality. This telemetry worker will be removed entirely when all events are migrated to V2.
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
	dbHandle := dbtest.NewDB(t)
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

	logger := logtest.Scoped(t)

	obsContext := observation.TestContextTB(t)

	return &telemetryHandler{
		logger:             logger,
		eventLogStore:      dbmocks.NewMockEventLogStore(),
		globalStateStore:   dbmocks.NewMockGlobalStateStore(),
		userEmailsStore:    dbmocks.NewMockUserEmailsStore(),
		bookmarkStore:      bms,
		sendEventsCallback: callbackFunc,
		metrics:            newHandlerMetrics(obsContext),
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

func mockEnvVars(t *testing.T, flag bool) {
	prevEnabled := enabled
	prevTopicName := topicName
	prevProjectName := projectName

	t.Cleanup(func() {
		enabled = prevEnabled
		topicName = prevTopicName
		projectName = prevProjectName
	})

	enabled = flag
	topicName = "test-name"
	projectName = "project-name"
}
