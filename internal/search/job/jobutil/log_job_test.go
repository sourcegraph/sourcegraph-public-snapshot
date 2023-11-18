package jobutil_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrytest"
	"github.com/sourcegraph/sourcegraph/schema"
)

type fakeEventLogStore struct {
	database.EventLogStore
	events []*database.Event
}

func (s *fakeEventLogStore) BulkInsert(_ context.Context, newEvents []*database.Event) error {
	s.events = append(s.events, newEvents...)
	return nil
}

func (s *fakeEventLogStore) loggedEventNames() []string {
	var names []string
	for _, e := range s.events {
		var present bool
		for _, n := range names {
			present = present || e.Name == n
		}
		if !present {
			names = append(names, e.Name)
		}
	}
	sort.Strings(names)
	return names
}

func TestOwnSearchEventNames(t *testing.T) {
	type wantEvents struct {
		legacy []string // we retain manual instrumentation of existing events
		new    []string // https://docs.sourcegraph.com/dev/background-information/telemetry
	}
	for literal, wantEventNames := range map[string]wantEvents{
		"file:has.owner(one@example.com)": {
			legacy: []string{"FileHasOwnerSearch", "search.latencies.file"},
			new:    []string{"search.latencies - file", "search - file.hasOwners"},
		},
		"select:file.owners": {
			legacy: []string{"SelectFileOwnersSearch", "search.latencies.repo"},
			new:    []string{"search.latencies - repo", "search - select.fileOwners"},
		},
	} {
		t.Run(literal, func(t *testing.T) {
			q, err := query.ParseLiteral(literal)
			if err != nil {
				t.Fatalf("ParseLiteral: %s", err)
			}
			inputs := &search.Inputs{
				UserSettings:        &schema.Settings{},
				PatternType:         query.SearchTypeLiteral,
				Protocol:            search.Streaming,
				OnSourcegraphDotCom: true,
				Query:               q,
			}

			gss := dbmocks.NewMockGlobalStateStore()
			gss.GetFunc.SetDefaultReturn(database.GlobalState{SiteID: "a"}, nil)

			db := dbmocks.NewMockDB()
			db.GlobalStateFunc.SetDefaultReturn(gss)
			// legacy events
			legacyEvents := &fakeEventLogStore{}
			db.EventLogsFunc.SetDefaultReturn(legacyEvents)
			// new events
			newEvents := telemetrytest.NewMockEventsExportQueueStore()
			db.TelemetryEventsExportQueueFunc.SetDefaultReturn(newEvents)

			ctx := actor.WithActor(context.Background(), actor.FromUser(42))
			childJob := mockjob.NewMockJob()
			logJob := jobutil.NewLogJob(inputs, childJob)
			if _, err := logJob.Run(ctx, job.RuntimeClients{
				Logger: logtest.Scoped(t),
				DB:     db,
			}, streaming.NewNullStream()); err != nil {
				t.Fatalf("LogJob.Run: %s", err)
			}
			// legacy events
			if diff := cmp.Diff(wantEventNames.legacy, legacyEvents.loggedEventNames()); diff != "" {
				t.Errorf("logged legacy events, -want+got: %s", diff)
			}
			// new events
			if diff := cmp.Diff(wantEventNames.new, newEvents.GetMockQueuedEvents().Summary()); diff != "" {
				t.Errorf("logged new events, -want+got: %s", diff)
			}
		})
	}
}
