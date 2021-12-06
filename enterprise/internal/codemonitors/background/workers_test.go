package background

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/slack-go/slack"
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

			_, err = ts.EnqueueActionJobsForMonitor(ctx, 1, triggerEventID)
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

func TestSlackWebhook(t *testing.T) {
	payload := &slack.WebhookMessage{
		Blocks: &slack.Blocks{BlockSet: []slack.Block{
			slack.NewSectionBlock(slack.NewTextBlockObject("plain_text", "New code monitoring event", false, false), nil, nil),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("`%s`", "repo:test"), false, true), nil, nil),
			slack.NewSectionBlock(slack.NewTextBlockObject("plain_text", fmt.Sprintf("There were %d new search results for your query", 5), false, false), nil, nil),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(`<%s|View search on Sourcegraph>`, "https://google.com"), false, false), nil, nil),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(`<%s|View code monitor>`, "https://facebook.com"), false, false), nil, nil),
		}},
	}
	b, _ := json.Marshal(payload)
	println(string(b))

	err := slack.PostWebhookCustomHTTPContext(context.Background(), "https://hooks.slack.com/services/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", http.DefaultClient, payload)
	require.NoError(t, err)
}
