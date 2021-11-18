package usagestats

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/amplitude"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// pubSubDotComEventsTopicID is the topic ID of the topic that forwards messages to Sourcegraph.com events' pub/sub subscribers.
var pubSubDotComEventsTopicID = env.Get("PUBSUB_DOTCOM_EVENTS_TOPIC_ID", "", "Pub/sub dotcom events topic ID is the pub/sub topic id where Sourcegraph.com events are published.")
var amplitudeAPIToken = env.Get("AMPLITUDE_API_TOKEN", "", "The API token for the Amplitude project to send data to.")

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
			if err := publishAmplitudeEvents(events); err != nil {
				log15.Error("publishAmplitudeEvents failed", "err", err)
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

		pubsubEvent, err := json.Marshal(bigQueryEvent{
			EventName:       event.EventName,
			UserID:          int(event.UserID),
			AnonymousUserID: event.UserCookieID,
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

func publishAmplitudeEvents(events []Event) error {
	if !envvar.SourcegraphDotComMode() {
		return nil
	}
	if amplitudeAPIToken == "" {
		return nil
	}

	amplitudeEvents, err := createAmplitudeEvents(events)
	if err != nil {
		return err
	}

	amplitudeEvent, err := json.Marshal(amplitude.EventPayload{
		APIKey: amplitudeAPIToken,
		Events: amplitudeEvents,
	})
	if err != nil {
		return err
	}

	return amplitude.Publish(amplitudeEvent)
}

func createAmplitudeEvents(events []Event) ([]amplitude.AmplitudeEvent, error) {
	amplitudeEvents := make([]amplitude.AmplitudeEvent, 0, len(events))
	for _, event := range events {
		if _, ok := amplitude.DenyList[event.EventName]; ok {
			continue
		}

		// For anonymous users, do not assign a user ID.
		// Amplitude does not want User IDs for anonymous users
		// so it can perform merging for users who sign up based on device ID.
		var userID string
		if event.UserID != 0 {
			// Minimum length for an Amplitude user ID is 5 characters.
			userID = fmt.Sprintf("%06d", event.UserID)
		}

		if event.DeviceID == nil {
			return nil, errors.New("amplitude: Missing device ID")
		}
		if event.EventID == nil {
			return nil, errors.New("amplitude: Missing event ID")
		}
		if event.InsertID == nil {
			return nil, errors.New("amplitude: Missing insert ID")
		}

		userProperties, err := json.Marshal(amplitude.UserProperties{
			AnonymousUserID: event.UserCookieID,
			FeatureFlags:    event.FeatureFlags,
		})
		if err != nil {
			return nil, err
		}

		amplitudeEvents = append(amplitudeEvents, amplitude.AmplitudeEvent{
			UserID:          userID,
			DeviceID:        *event.DeviceID,
			InsertID:        *event.InsertID,
			EventID:         *event.EventID,
			EventType:       event.EventName,
			EventProperties: event.PublicArgument,
			UserProperties:  userProperties,
			Time:            time.Now().Unix(),
		})
	}

	return amplitudeEvents, nil
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
