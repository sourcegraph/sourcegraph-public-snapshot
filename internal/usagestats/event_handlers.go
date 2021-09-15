package usagestats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/amplitude"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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
	// FirstSourceURL is only measured for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	FirstSourceURL *string
	URL            string
	Source         string
	FeatureFlags   featureflag.FlagSet
	CohortID       *string
	// Referrer is only measured for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	Referrer       *string
	Argument       json.RawMessage
	PublicArgument json.RawMessage
}

// LogBackendEvent is a convenience function for logging backend events.
func LogBackendEvent(db dbutil.DB, userID int32, eventName string, argument, publicArgument json.RawMessage, featureFlags featureflag.FlagSet, cohortID *string) error {
	return LogEvent(context.Background(), db, Event{
		EventName:      eventName,
		UserID:         userID,
		UserCookieID:   "backend", // Use a non-empty string here to avoid the event_logs table's user existence constraint causing issues
		URL:            "",
		Source:         "BACKEND",
		Argument:       argument,
		PublicArgument: publicArgument,
		FeatureFlags:   featureFlags,
		CohortID:       cohortID,
	})
}

// LogEvent logs an event.
func LogEvent(ctx context.Context, db dbutil.DB, args Event) error {
	if !conf.EventLoggingEnabled() {
		return nil
	}
	if envvar.SourcegraphDotComMode() {
		err := publishSourcegraphDotComEvent(args)
		if err != nil {
			return err
		}
		err = publishAmplitudeEvent(args)
		if err != nil {
			return err
		}
	}
	return logLocalEvent(ctx, db, args.EventName, args.URL, args.UserID, args.UserCookieID, args.Source, args.Argument, args.PublicArgument, args.FeatureFlags, args.CohortID)
}

type bigQueryEvent struct {
	EventName       string  `json:"name"`
	AnonymousUserID string  `json:"anonymous_user_id"`
	FirstSourceURL  string  `json:"first_source_url"`
	UserID          int     `json:"user_id"`
	Source          string  `json:"source"`
	Timestamp       string  `json:"timestamp"`
	Version         string  `json:"version"`
	FeatureFlags    string  `json:"feature_flags"`
	CohortID        *string `json:"cohort_id,omitempty"`
	Referrer        string  `json:"referrer,omitempty"`
	PublicArgument  string  `json:"public_argument"`
}

type amplitudeEventJson struct {
	APIKey string           `json:"api_key"`
	Events []amplitudeEvent `json:"events"`
}
type amplitudeEvent struct {
	UserID          string          `json:"user_id"`
	DeviceID        string          `json:"device_id"`
	EventType       string          `json:"event_type"`
	EventProperties json.RawMessage `json:"event_properties,omitempty"`
	UserProperties  json.RawMessage `json:"user_properties,omitempty"`
	Time            int64           `json:"time,omitempty"`
	// // Event Property
	// Version string `json:"version"`
	// // User Property
	// FeatureFlags string `json:"feature_flags"`
	// // User Property
	// CohortID *string `json:"cohort_id,omitempty"`
	// // User Property
	// Referrer string `json:"referrer,omitempty"`
	// // User Property
	// AnonymousUserID string `json:"anonymous_user_id"`
	// // User Property
	// FirstSourceURL string `json:"first_source_url"`
}

// publishSourcegraphDotComEvent publishes Sourcegraph.com events to BigQuery.
func publishSourcegraphDotComEvent(args Event) error {
	if !envvar.SourcegraphDotComMode() {
		return nil
	}
	if pubSubDotComEventsTopicID == "" {
		return nil
	}
	firstSourceURL := ""
	if args.FirstSourceURL != nil {
		firstSourceURL = *args.FirstSourceURL
	}
	referrer := ""
	if args.Referrer != nil {
		referrer = *args.Referrer
	}
	featureFlagJSON, err := json.Marshal(args.FeatureFlags)
	if err != nil {
		return err
	}

	pubsubEvent, err := json.Marshal(bigQueryEvent{
		EventName:       args.EventName,
		UserID:          int(args.UserID),
		AnonymousUserID: args.UserCookieID,
		FirstSourceURL:  firstSourceURL,
		Referrer:        referrer,
		Source:          args.Source,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Version:         version.Version(),
		FeatureFlags:    string(featureFlagJSON),
		CohortID:        args.CohortID,
		PublicArgument:  string(args.PublicArgument),
	})
	if err != nil {
		return err
	}

	return pubsub.Publish(pubSubDotComEventsTopicID, string(pubsubEvent))
}

func publishAmplitudeEvent(args Event) error {
	if !envvar.SourcegraphDotComMode() {
		return nil
	}
	if amplitudeAPIToken == "" {
		return nil
	}
	// For anonymous users, do not assign a user ID.
	// Amplitude does not want User IDs for anonymous users
	// so it can perform merging for users who sign up based on device ID.
	var userID string
	if args.UserID != 0 {
		// Minimum length for an Amplitude user ID is 5 characters.
		userID = fmt.Sprintf("%06d", args.UserID)
	}

	userProperties, err := getAmplitudeUserProperties(args)
	if err != nil {
		return err
	}

	amplitudeEvent, err := json.Marshal(amplitudeEventJson{
		APIKey: amplitudeAPIToken,
		Events: []amplitudeEvent{{
			UserID: userID,
			// TODO(FARHAN): Need to store this in context or cookie and pull it each time. Perhaps send with event from
			// the client side.
			DeviceID:        fmt.Sprintf(`%s-%d`, string(args.UserID), time.Now().Unix()),
			EventType:       args.EventName,
			EventProperties: args.PublicArgument,
			UserProperties:  userProperties,
			Time:            time.Now().Unix(),
		}},
	})
	if err != nil {
		return err
	}

	return amplitude.Publish(amplitudeAPIToken, amplitudeEvent)

}

// AnonymousUserID: args.UserCookieID,
// FirstSourceURL:  firstSourceURL,
// Referrer:        referrer,
// Timestamp:       time.Now().UTC().Format(time.RFC3339),
// FeatureFlags:    string(featureFlagJSON),
// CohortID:        args.CohortID,
// NumberOfReposAdded
// HasAddedRepos
// NumberPublicReposAdded
// NumberPrivateReposAdded
// ActiveCodeHost
// CohortWeek
// IsSourcegraphTeammate
// Feature Flag - w0SignupOptimization, w1SignupOptimization, SearchNotebookOnboarding
type AmplitudeUserProperties struct {
	AnonymousUserID string              `json:"anonymous_user_id"`
	FirstSourceURL  string              `json:"first_source_url"`
	FeatureFlags    featureflag.FlagSet `json:"feature_flags"`
	CohortID        *string             `json:"cohort_id,omitempty"`
	Referrer        string              `json:"referrer,omitempty"`
}

func getAmplitudeUserProperties(args Event) (json.RawMessage, error) {
	firstSourceURL := ""
	if args.FirstSourceURL != nil {
		firstSourceURL = *args.FirstSourceURL
	}
	referrer := ""
	if args.Referrer != nil {
		referrer = *args.Referrer
	}

	userProperties, err := json.Marshal(AmplitudeUserProperties{
		AnonymousUserID: args.UserCookieID,
		FirstSourceURL:  firstSourceURL,
		Referrer:        referrer,
		CohortID:        args.CohortID,
		FeatureFlags:    args.FeatureFlags,
	})
	if err != nil {
		return nil, err
	}

	return userProperties, nil
}

// logLocalEvent logs users events.
func logLocalEvent(ctx context.Context, db dbutil.DB, name, url string, userID int32, userCookieID, source string, argument, publicArgument json.RawMessage, featureFlags featureflag.FlagSet, cohortID *string) error {
	if name == "SearchResultsQueried" {
		err := logSiteSearchOccurred()
		if err != nil {
			return err
		}
	}
	if name == "findReferences" {
		err := logSiteFindRefsOccurred()
		if err != nil {
			return err
		}
	}

	info := &database.Event{
		Name:            name,
		URL:             url,
		UserID:          uint32(userID),
		AnonymousUserID: userCookieID,
		Source:          source,
		Argument:        argument,
		Timestamp:       timeNow().UTC(),
		FeatureFlags:    featureFlags,
		CohortID:        cohortID,
		PublicArgument:  publicArgument,
	}
	return database.EventLogs(db).Insert(ctx, info)
}
