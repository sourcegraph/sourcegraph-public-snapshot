package amplitude

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

// EventPayload represents the structure of the payloads we send to Amplitude
// when we log an event. We can send more than one event in Events, but we currently
// send events in individual payloads.
type EventPayload struct {
	APIKey string           `json:"api_key"`
	Events []AmplitudeEvent `json:"events"`
}

// AmplitudeEvent represents the fields that make up an event
// in Amplitude.
type AmplitudeEvent struct {
	UserID          string          `json:"user_id"`
	DeviceID        string          `json:"device_id"`
	EventID         int32           `json:"event_id"`
	InsertID        string          `json:"insert_id"`
	EventType       string          `json:"event_type"`
	EventProperties json.RawMessage `json:"event_properties,omitempty"`
	UserProperties  json.RawMessage `json:"user_properties,omitempty"`
	Time            int64           `json:"time,omitempty"`
}

// UserProperties contains the list of user properties we collect and send to Amplitude.
type UserProperties struct {
	AnonymousUserID         string              `json:"anonymous_user_id"`
	FirstSourceURL          string              `json:"first_source_url"`
	FeatureFlags            featureflag.FlagSet `json:"feature_flags"`
	CohortID                *string             `json:"cohort_id,omitempty"`
	Referrer                string              `json:"referrer,omitempty"`
	HasCloudAccount         bool                `json:"has_cloud_account"`
	HasAddedRepos           bool                `json:"has_added_repos"`
	NumberOfReposAdded      int                 `json:"number_repos_added"`
	NumberPublicReposAdded  int                 `json:"number_public_repos_added"`
	NumberPrivateReposAdded int                 `json:"number_private_repos_added"`
	HasActiveCodeHost       bool                `json:"has_active_code_host"`
	IsSourcegraphTeammate   bool                `json:"is_sourcegraph_teammate"`
}

// FrontendUserProperties contains the subset of user properties that are stored
// in localStorage in the web app, and passed in the userProperties field of
// Events in the EventLogger.
type FrontendUserProperties struct {
	HasAddedRepos           bool `json:"hasAddedRepositories"`
	NumberOfReposAdded      int  `json:"numberOfRepositoriesAdded"`
	NumberPublicReposAdded  int  `json:"numberOfPublicRepos"`
	NumberPrivateReposAdded int  `json:"numberOfPrivateRepos"`
	HasActiveCodeHost       bool `json:"hasActiveCodeHost"`
	IsSourcegraphTeammate   bool `json:"isSourcegraphTeammate"`
}
