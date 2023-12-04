package usagestats

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/siteid"
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
	LastSourceURL    *string
	URL              string
	Source           string
	EvaluatedFlagSet featureflag.EvaluatedFlagSet
	CohortID         *string
	// Referrer is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	Referrer               *string
	OriginalReferrer       *string
	SessionReferrer        *string
	SessionFirstURL        *string
	Argument               json.RawMessage
	PublicArgument         json.RawMessage
	UserProperties         json.RawMessage
	DeviceID               *string
	InsertID               *string
	EventID                *int32
	DeviceSessionID        *string
	Client                 *string
	BillingProductCategory *string
	BillingEventID         *string
	// ConnectedSiteID is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	ConnectedSiteID *string
	// HashedLicenseKey is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	HashedLicenseKey *string
}

// LogBackendEvent is a convenience function for logging backend events.
//
// ❗ DEPRECATED: Use event recorders from internal/telemetryrecorder instead.
func LogBackendEvent(db database.DB, userID int32, deviceID, eventName string, argument, publicArgument json.RawMessage, evaluatedFlagSet featureflag.EvaluatedFlagSet, cohortID *string) error {
	insertID, _ := uuid.NewRandom()
	insertIDFinal := insertID.String()
	eventID := int32(rand.Int())

	client := "SERVER_BACKEND"
	if envvar.SourcegraphDotComMode() {
		client = "DOTCOM_BACKEND"
	}
	if deploy.IsApp() {
		client = "APP_BACKEND"
	}

	hashedLicenseKey := conf.HashedCurrentLicenseKeyForAnalytics()
	connectedSiteID := siteid.Get(db)

	return LogEvent(context.Background(), db, Event{
		EventName:        eventName,
		UserID:           userID,
		UserCookieID:     "backend", // Use a non-empty string here to avoid the event_logs table's user existence constraint causing issues
		URL:              "",
		Source:           "BACKEND",
		Argument:         argument,
		PublicArgument:   publicArgument,
		UserProperties:   json.RawMessage("{}"),
		EvaluatedFlagSet: evaluatedFlagSet,
		CohortID:         cohortID,
		DeviceID:         &deviceID,
		InsertID:         &insertIDFinal,
		EventID:          &eventID,
		Client:           &client,
		ConnectedSiteID:  &connectedSiteID,
		HashedLicenseKey: &hashedLicenseKey,
	})
}

// LogEvent logs an event.
//
// Deprecated: Use EventRecorder from internal/telemetryrecorder instead.
// Learn more: https://docs.sourcegraph.com/dev/background-information/telemetry
func LogEvent(ctx context.Context, db database.DB, args Event) error {
	return LogEvents(ctx, db, []Event{args})
}

// LogEvents logs a batch of events.
//
// Deprecated: Use EventRecorder from internal/telemetryrecorder instead.
// Learn more: https://docs.sourcegraph.com/dev/background-information/telemetry
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
	EventName              string  `json:"name"`
	URL                    string  `json:"url"`
	AnonymousUserID        string  `json:"anonymous_user_id"`
	FirstSourceURL         string  `json:"first_source_url"`
	LastSourceURL          string  `json:"last_source_url"`
	UserID                 int     `json:"user_id"`
	Source                 string  `json:"source"`
	Timestamp              string  `json:"timestamp"`
	Version                string  `json:"version"`
	FeatureFlags           string  `json:"feature_flags"`
	CohortID               *string `json:"cohort_id,omitempty"`
	Referrer               string  `json:"referrer,omitempty"`
	OriginalReferrer       string  `json:"original_referrer"`
	SessionReferrer        string  `json:"session_referrer"`
	SessionFirstURL        string  `json:"session_first_url"`
	PublicArgument         string  `json:"public_argument"`
	DeviceID               *string `json:"device_id,omitempty"`
	InsertID               *string `json:"insert_id,omitempty"`
	DeviceSessionID        *string `json:"device_session_id,omitempty"`
	Client                 *string `json:"client,omitempty"`
	BillingProductCategory *string `json:"billing_product_category,omitempty"`
	BillingEventID         *string `json:"billing_event_id,omitempty"`
	ConnectedSiteID        *string `json:"connected_site_id,omitempty"`
	HashedLicenseKey       *string `json:"hashed_license_key,omitempty"`
}

var (
	pubsubClient     pubsub.TopicClient
	pubsubClientOnce sync.Once
	pubsubClientErr  error
)

// publishSourcegraphDotComEvents publishes Sourcegraph.com events to BigQuery.
func publishSourcegraphDotComEvents(events []Event) error {
	if !envvar.SourcegraphDotComMode() || pubSubDotComEventsTopicID == "" {
		return nil
	}
	pubsubClientOnce.Do(func() {
		pubsubClient, pubsubClientErr = pubsub.NewDefaultTopicClient(pubSubDotComEventsTopicID)
	})
	if pubsubClientErr != nil {
		return pubsubClientErr
	}

	pubsubEvents, err := serializePublishSourcegraphDotComEvents(events)
	if err != nil {
		return err
	}
	return pubsubClient.Publish(context.Background(), pubsubEvents...)
}

func serializePublishSourcegraphDotComEvents(events []Event) ([][]byte, error) {
	pubsubEvents := make([][]byte, 0, len(events))
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

		originalReferrer := ""
		if event.OriginalReferrer != nil {
			originalReferrer = *event.OriginalReferrer
		}

		sessionReferrer := ""
		if event.SessionReferrer != nil {
			sessionReferrer = *event.SessionReferrer
		}

		sessionFirstURL := ""
		if event.SessionFirstURL != nil {
			sessionFirstURL = *event.SessionFirstURL
		}

		featureFlagJSON, err := json.Marshal(event.EvaluatedFlagSet)
		if err != nil {
			return nil, err
		}
		saferUrl, err := redactSensitiveInfoFromCloudURL(event.URL)
		if err != nil {
			return nil, err
		}

		pubsubEvent, err := json.Marshal(bigQueryEvent{
			EventName:              event.EventName,
			UserID:                 int(event.UserID),
			AnonymousUserID:        event.UserCookieID,
			URL:                    saferUrl,
			FirstSourceURL:         firstSourceURL,
			LastSourceURL:          lastSourceURL,
			Referrer:               referrer,
			OriginalReferrer:       originalReferrer,
			SessionReferrer:        sessionReferrer,
			SessionFirstURL:        sessionFirstURL,
			Source:                 event.Source,
			Timestamp:              time.Now().UTC().Format(time.RFC3339),
			Version:                version.Version(),
			FeatureFlags:           string(featureFlagJSON),
			CohortID:               event.CohortID,
			PublicArgument:         string(event.PublicArgument),
			DeviceID:               event.DeviceID,
			InsertID:               event.InsertID,
			DeviceSessionID:        event.DeviceSessionID,
			Client:                 event.Client,
			BillingProductCategory: event.BillingProductCategory,
			BillingEventID:         event.BillingEventID,
			ConnectedSiteID:        event.ConnectedSiteID,
			HashedLicenseKey:       event.HashedLicenseKey,
		})
		if err != nil {
			return nil, err
		}

		pubsubEvents = append(pubsubEvents, pubsubEvent)
	}

	return pubsubEvents, nil
}

// logLocalEvents logs a batch of user events.
func logLocalEvents(ctx context.Context, db database.DB, events []Event) error {
	databaseEvents, err := serializeLocalEvents(events)
	if err != nil {
		return err
	}

	// Use EventRecorder from internal/telemetryrecorder instead - logLocalEvents
	// should eventually be removed entirely.
	//lint:ignore SA1019 existing usage of deprecated functionality.
	return db.EventLogs().BulkInsert(ctx, databaseEvents)
}

func serializeLocalEvents(events []Event) ([]*database.Event, error) {
	databaseEvents := make([]*database.Event, 0, len(events))
	for _, event := range events {
		// If this event should only be logged to our remote data warehouse, simply exclude it
		// from the serialized events for the local database.
		for _, eventToOnlyLogRemotely := range eventlogger.OnlyLogRemotelyEvents {
			if event.EventName == eventToOnlyLogRemotely {
				continue
			}
		}

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
			Name:                   event.EventName,
			URL:                    event.URL,
			UserID:                 uint32(event.UserID),
			AnonymousUserID:        event.UserCookieID,
			Source:                 event.Source,
			Argument:               event.Argument,
			Timestamp:              timeNow().UTC(),
			EvaluatedFlagSet:       event.EvaluatedFlagSet,
			CohortID:               event.CohortID,
			PublicArgument:         event.PublicArgument,
			FirstSourceURL:         event.FirstSourceURL,
			LastSourceURL:          event.LastSourceURL,
			Referrer:               event.Referrer,
			DeviceID:               event.DeviceID,
			InsertID:               event.InsertID,
			Client:                 event.Client,
			BillingProductCategory: event.BillingProductCategory,
			BillingEventID:         event.BillingEventID,
		})
	}

	return databaseEvents, nil
}

// redactSensitiveInfoFromCloudURL redacts portions of URLs that
// may contain sensitive info on Sourcegraph Cloud. We replace all paths,
// and only maintain query parameters in a specified allowlist,
// which are known to be essential for marketing analytics on Sourcegraph Cloud.

func redactSensitiveInfoFromCloudURL(rawURL string) (string, error) {
	// Because Sourcegraph.com only contains public code, URLs do not contain sensitive information.
	// Redaction is only used for URLs from cloud and self-hosted instance telemetry.
	if envvar.SourcegraphDotComMode() {
		return rawURL, nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

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
	// Redact non-marketing query parameter values, while retaining keys for analytics.
	// Allowlisted parameters remain unchanged to protect marketing data integrity.
	for key := range urlQueryParams {
		if _, ok := marketingQueryParameters[key]; !ok {
			urlQueryParams[key] = []string{"redacted"}
		}
	}

	// Retain only first part of the URL's path segment for security(avoid leaking sensitive path info)
	pathParts := strings.Split(parsedURL.Path, "/")

	// Check length to avoid index out of range error
	if len(pathParts) > 1 {
		parsedURL.Path = pathParts[1]

		// Add '/redacted' if we removed parts of the original path
		if len(pathParts) > 2 {
			parsedURL.Path += "/redacted"
		}
	}

	parsedURL.RawQuery = urlQueryParams.Encode()

	return parsedURL.String(), nil
}
