package usagestats

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCreateBigQueryEvent(t *testing.T) {
	testUrl := "https://sourcegraph.com/test"
	event := Event{
		EventName:      "ExtensionToggled",
		UserID:         1,
		UserCookieID:   "abcd",
		FirstSourceURL: &testUrl,
		URL:            "https://sourcegraph.com/extensions",
		Source:         "test",
		FeatureFlags:   map[string]bool{},
		CohortID:       nil,
		Referrer:       &testUrl,
		Argument:       json.RawMessage("\"extension_id\":\"sourcegraph/codecov\""),
	}

	got, err := createBigQueryEvent(event)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)
	argumentField := "\"extension_id\":\"sourcegraph/codecov\""
	want := &bigQueryEvent{
		EventName:       "ExtensionToggled",
		UserID:          1,
		AnonymousUserID: "abcd",
		FirstSourceURL:  testUrl,
		Source:          "test",
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		FeatureFlags:    "{}",
		CohortID:        nil,
		Version:         "0.0.0+dev",
		Referrer:        testUrl,
		PublicArguments: &argumentField,
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}

}
func TestCreateBigQueryEvent_PrivateEvent(t *testing.T) {
	testUrl := "https://sourcegraph.com/test"
	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)
	privateArgumentField := "\"code_search\": { \"query_data\": { \"query\": \"test\", \"combined\": \"something private\", \"empty\": \"false\"}"
	eventWithPrivateArgs := Event{
		EventName:      "SearchResultsFetched",
		UserID:         1,
		UserCookieID:   "abcd",
		FirstSourceURL: &testUrl,
		URL:            "https://sourcegraph.com/extensions",
		Source:         "test",
		FeatureFlags:   map[string]bool{},
		CohortID:       nil,
		Referrer:       &testUrl,
		Argument:       json.RawMessage(privateArgumentField),
	}
	got, err := createBigQueryEvent(eventWithPrivateArgs)
	if err != nil {
		t.Fatal(err)
	}
	want := &bigQueryEvent{
		EventName:       "SearchResultsFetched",
		UserID:          1,
		AnonymousUserID: "abcd",
		FirstSourceURL:  testUrl,
		Source:          "test",
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		FeatureFlags:    "{}",
		CohortID:        nil,
		Version:         "0.0.0+dev",
		Referrer:        testUrl,
		// PublicArguments should be nil because
		// SearchResultsFetched is not in the BigQuery
		// allowlist.
		PublicArguments: nil,
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}

}
