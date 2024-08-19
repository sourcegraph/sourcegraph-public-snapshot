package jobutil_test

import (
	"context"
	"sort"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

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
		legacy autogold.Value // we retain manual instrumentation of existing events
		new    autogold.Value // https://docs-legacy.sourcegraph.com/dev/background-information/telemetry
	}

	cases := []struct {
		query      string
		searchType query.SearchType
		want       wantEvents
	}{
		// result types
		{
			query:      "sourcegraph type:repo",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.repo"}),
				new:    autogold.Expect([]string{"search.latencies - repo"}),
			},
		},
		{
			query:      "sourcegraph type:diff",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.diff"}),
				new:    autogold.Expect([]string{"search.latencies - diff"}),
			},
		},
		{
			query:      "search results type:file",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.keyword"}),
				new:    autogold.Expect([]string{"search.latencies - keyword"}),
			},
		},
		{
			query:      "search results type:path",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.file"}),
				new:    autogold.Expect([]string{"search.latencies - file"}),
			},
		},
		// pattern types
		{
			query:      "bytes buffer",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.keyword"}),
				new:    autogold.Expect([]string{"search.latencies - keyword"}),
			},
		},
		{
			query:      "bytes",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.keyword"}),
				new:    autogold.Expect([]string{"search.latencies - keyword"}),
			},
		},
		{
			query:      "bytes buffer",
			searchType: query.SearchTypeStandard,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.standard"}),
				new:    autogold.Expect([]string{"search.latencies - standard"}),
			},
		},
		{
			query:      "bytes buffer",
			searchType: query.SearchTypeRegex,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.regex"}),
				new:    autogold.Expect([]string{"search.latencies - regex"}),
			},
		},
		{
			query:      "if ... else ...",
			searchType: query.SearchTypeStructural,
			want: wantEvents{
				legacy: autogold.Expect([]string{"search.latencies.structural"}),
				new:    autogold.Expect([]string{"search.latencies - structural"}),
			},
		},
		// other
		{
			query:      "file:has.owner(one@example.com)",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"FileHasOwnerSearch", "search.latencies.keyword"}),
				new:    autogold.Expect([]string{"search.latencies - keyword", "search - file.hasOwners"}),
			},
		},
		{
			query:      "select:file.owners",
			searchType: query.SearchTypeKeyword,
			want: wantEvents{
				legacy: autogold.Expect([]string{"SelectFileOwnersSearch", "search.latencies.keyword"}),
				new:    autogold.Expect([]string{"search.latencies - keyword", "search - select.fileOwners"}),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			q, err := query.ParseSearchType(c.query, c.searchType)
			require.NoError(t, err)
			inputs := &search.Inputs{
				UserSettings:        &schema.Settings{},
				PatternType:         c.searchType,
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
			_, err = logJob.Run(ctx, job.RuntimeClients{
				Logger: logtest.Scoped(t),
				DB:     db,
			}, streaming.NewNullStream())
			require.NoError(t, err)

			c.want.legacy.Equal(t, legacyEvents.loggedEventNames())
			c.want.new.Equal(t, newEvents.GetMockQueuedEvents().Summary())
		})
	}
}
