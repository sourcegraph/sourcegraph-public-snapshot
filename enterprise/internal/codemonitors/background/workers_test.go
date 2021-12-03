package background

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/email"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/storetest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestActionRunner(t *testing.T) {
	tests := []struct {
		name               string
		numResults         int
		wantNumResultsText string
	}{
		{
			name:               "5 results",
			numResults:         5,
			wantNumResultsText: "There were 5 new search results for your query",
		},
		{
			name:               "1 result",
			numResults:         1,
			wantNumResultsText: "There was 1 new search result for your query",
		},
	}

	var (
		queryID int64 = 1
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := dbtest.NewDB(t)

			externalURL := "https://www.sourcegraph.com"
			testQuery := "test patternType:literal"

			// Mocks.
			got := email.TemplateDataNewSearchResults{}
			email.MockSendEmailForNewSearchResult = func(ctx context.Context, userID int32, data *email.TemplateDataNewSearchResults) error {
				got = *data
				return nil
			}
			email.MockExternalURL = func() *url.URL {
				externalURL, _ := url.Parse("https://www.sourcegraph.com")
				return externalURL
			}

			// Create a TestStore.
			now := time.Now()
			clock := func() time.Time { return now }
			s := codemonitors.NewStoreWithClock(db, clock)
			ctx, ts := storetest.NewTestStore(t, db)

			_, _, _, userCtx := storetest.NewTestUser(ctx, t, db)

			// Run a complete pipeline from creation of a code monitor to sending of an email.
			_, err := ts.InsertTestMonitor(userCtx, t)
			require.NoError(t, err)

			triggerJobs, err := ts.EnqueueQueryTriggerJobs(ctx)
			require.NoError(t, err)
			require.Len(t, triggerJobs, 1)
			triggerEventID := triggerJobs[0].ID

			err = ts.UpdateTriggerJobWithResults(ctx, triggerEventID, testQuery, tt.numResults)
			require.NoError(t, err)

			_, err = ts.EnqueueActionJobsForQuery(ctx, queryID, triggerEventID)
			require.NoError(t, err)

			record, err := ts.GetActionJob(ctx, 1)
			require.NoError(t, err)

			a := actionRunner{s}
			err = a.Handle(ctx, record)
			require.NoError(t, err)

			want := email.TemplateDataNewSearchResults{
				Priority:       "New",
				SearchURL:      externalURL + "/search?q=test+patternType%3Aliteral&utm_source=code-monitoring-email",
				Description:    "test description",
				CodeMonitorURL: externalURL + "/code-monitoring/" + string(relay.MarshalID("CodeMonitor", 1)) + "?utm_source=code-monitoring-email",
			}

			want.NumberOfResultsWithDetail = tt.wantNumResultsText
			require.Equal(t, want, got)
		})
	}
}
