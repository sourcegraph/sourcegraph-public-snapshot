package usagestats

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// pubSubDotComEventsTopicID is the topic ID of the topic that forwards messages to Sourcegraph.com events' pub/sub subscribers.
var pubSubDotComEventsTopicID = env.Get("PUBSUB_DOTCOM_EVENTS_TOPIC_ID", "", "Pub/sub dotcom events topic ID is the pub/sub topic id where Sourcegraph.com events are published.")

// Event represents a request to log telemetry.
type Event struct {
	EventName    string
	UserID       int32
	UserCookieID string
	// FirstSourceURL is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	FirstSourceURL *string
	// LastSourceURL is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	LastSourceURL *string
	URL           string
	Source        string
	FeatureFlags  featureflag.FlagSet
	CohortID      *string
	// Referrer is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	Referrer       *string
	Argument       json.RawMessage
	PublicArgument json.RawMessage
	UserProperties json.RawMessage
	DeviceID       *string
	InsertID       *string
	EventID        *int32
}

// LogBackendEvent is a convenience function for logging backend events.
func LogBackendEvent(db database.DB, userID int32, deviceID, eventName string, argument, publicArgument json.RawMessage, featureFlags featureflag.FlagSet, cohortID *string) error {
	insertID, _ := uuid.NewRandom()
	insertIDFinal := insertID.String()
	eventID := int32(rand.Int())
	return LogEvent(context.Background(), db, Event{
		EventName:      eventName,
		UserID:         userID,
		UserCookieID:   "backend", // Use a non-empty string here to avoid the event_logs table's user existence constraint causing issues
		URL:            "",
		Source:         "BACKEND",
		Argument:       argument,
		PublicArgument: publicArgument,
		UserProperties: json.RawMessage("{}"),
		FeatureFlags:   featureFlags,
		CohortID:       cohortID,
		DeviceID:       &deviceID,
		InsertID:       &insertIDFinal,
		EventID:        &eventID,
	})
}

// LogEvent logs an event.
func LogEvent(ctx context.Context, db database.DB, args Event) error {
	return LogEvents(ctx, db, []Event{args})
}

// LogEvents logs a batch of events.
func LogEvents(ctx context.Context, db database.DB, events []Event) error {
	if !conf.EventLoggingEnabled() {
		return nil
	}

	if envvar.SourcegraphDotComMode() {
		go func() {
			if err := publishSourcegraphDotComEvents(events); err != nil {
				log15.Error("publishSourcegraphDotComEvents failed", "err", err)
			}
		}()
	}

	if err := logLocalEvents(ctx, db, events); err != nil {
		return err
	}

	return nil
}

type bigQueryEvent struct {
	EventName       string  `json:"name"`
	URL             string  `json:"url"`
	AnonymousUserID string  `json:"anonymous_user_id"`
	FirstSourceURL  string  `json:"first_source_url"`
	LastSourceURL   string  `json:"last_source_url"`
	UserID          int     `json:"user_id"`
	Source          string  `json:"source"`
	Timestamp       string  `json:"timestamp"`
	Version         string  `json:"version"`
	FeatureFlags    string  `json:"feature_flags"`
	CohortID        *string `json:"cohort_id,omitempty"`
	Referrer        string  `json:"referrer,omitempty"`
	PublicArgument  string  `json:"public_argument"`
	DeviceID        *string `json:"device_id,omitempty"`
	InsertID        *string `json:"insert_id,omitempty"`
}

// publishSourcegraphDotComEvents publishes Sourcegraph.com events to BigQuery.
func publishSourcegraphDotComEvents(events []Event) error {
	if !envvar.SourcegraphDotComMode() {
		return nil
	}
	if pubSubDotComEventsTopicID == "" {
		return nil
	}

	pubsubEvents, err := serializePublishSourcegraphDotComEvents(events)
	if err != nil {
		return err
	}

	for _, event := range pubsubEvents {
		if err := pubsub.Publish(pubSubDotComEventsTopicID, event); err != nil {
			return err
		}
	}

	return nil
}

func serializePublishSourcegraphDotComEvents(events []Event) ([]string, error) {
	pubsubEvents := make([]string, 0, len(events))
	for _, event := range events {
		firstSourceURL := ""
		if event.FirstSourceURL != nil {
			firstSourceURL = *event.FirstSourceURL
		}
		lastSourceURL := ""
		if event.LastSourceURL != nil {
			lastSourceURL = *event.LastSourceURL
		}
		referrer := ""
		if event.Referrer != nil {
			referrer = *event.Referrer
		}
		featureFlagJSON, err := json.Marshal(event.FeatureFlags)
		if err != nil {
			return nil, err
		}

		url, err := redactSensitiveInfoFromCloudURL(event.URL)
		if err != nil {
			return nil, err
		}

		pubsubEvent, err := json.Marshal(bigQueryEvent{
			EventName:       event.EventName,
			UserID:          int(event.UserID),
			AnonymousUserID: event.UserCookieID,
			URL:             url,
			FirstSourceURL:  firstSourceURL,
			LastSourceURL:   lastSourceURL,
			Referrer:        referrer,
			Source:          event.Source,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			Version:         version.Version(),
			FeatureFlags:    string(featureFlagJSON),
			CohortID:        event.CohortID,
			PublicArgument:  string(event.PublicArgument),
			DeviceID:        event.DeviceID,
			InsertID:        event.InsertID,
		})
		if err != nil {
			return nil, err
		}

		pubsubEvents = append(pubsubEvents, string(pubsubEvent))
	}

	return pubsubEvents, nil
}

// logLocalEvents logs a batch of user events.
func logLocalEvents(ctx context.Context, db database.DB, events []Event) error {
	databaseEvents, err := serializeLocalEvents(events)
	if err != nil {
		return err
	}

	return db.EventLogs().BulkInsert(ctx, databaseEvents)
}

func serializeLocalEvents(events []Event) ([]*database.Event, error) {
	databaseEvents := make([]*database.Event, 0, len(events))
	for _, event := range events {
		if event.EventName == "SearchResultsQueried" {
			if err := logSiteSearchOccurred(); err != nil {
				return nil, err
			}
		}
		if event.EventName == "findReferences" {
			if err := logSiteFindRefsOccurred(); err != nil {
				return nil, err
			}
		}

		databaseEvents = append(databaseEvents, &database.Event{
			Name:            event.EventName,
			URL:             event.URL,
			UserID:          uint32(event.UserID),
			AnonymousUserID: event.UserCookieID,
			Source:          event.Source,
			Argument:        event.Argument,
			Timestamp:       timeNow().UTC(),
			FeatureFlags:    event.FeatureFlags,
			CohortID:        event.CohortID,
			PublicArgument:  event.PublicArgument,
		})
	}

	return databaseEvents, nil
}

// redactSensitiveInfoFromCloudURL redacts portions of URLs that
// may contain sensitive info on Sourcegraph Cloud. We replace all paths,
// and only maintain query parameters in a specified allowlist,
// which are known to be essential for marketing analytics on Sourcegraph Cloud.
func redactSensitiveInfoFromCloudURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Host != "sourcegraph.com" {
		return rawURL, nil
	}

	parsedURL.RawPath = "/redacted"
	parsedURL.Path = "/redacted"

	marketingQueryParameters := map[string]struct{}{
		"utm_source":   {},
		"utm_campaign": {},
		"utm_medium":   {},
		"utm_term":     {},
		"utm_content":  {},
		"utm_cid":      {},
		"obility_id":   {},
		"campaign_id":  {},
		"ad_id":        {},
		"offer":        {},
		"gclid":        {},
	}
	urlQueryParams, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return "", err
	}
	for key := range urlQueryParams {
		if _, ok := marketingQueryParameters[key]; !ok {
			urlQueryParams[key] = []string{"redacted"}
		}
	}

	parsedURL.RawQuery = urlQueryParams.Encode()

	return parsedURL.String(), nil
}
