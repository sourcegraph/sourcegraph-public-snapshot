package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// TelemetryRootResolver provides TelemetryResolver via field 'telemetry' as
// defined in telemetry.graphql
type TelemetryRootResolver struct{ Resolver TelemetryResolver }

func (t *TelemetryRootResolver) Telemetry() TelemetryResolver { return t.Resolver }

type TelemetryResolver interface {
	// Queries
	ExportedEvents(context.Context, *ExportedEventsArgs) (ExportedEventsConnectionResolver, error)

	// Mutations
	RecordEvents(context.Context, *RecordEventsArgs) (*EmptyResponse, error)
}

type ExportedEventsArgs struct {
	First int32
	After *string
}

type ExportedEventResolver interface {
	ID() graphql.ID
	ExportedAt() gqlutil.DateTime
	Payload() (JSONValue, error)
}

type ExportedEventsConnectionResolver interface {
	Nodes() []ExportedEventResolver
	TotalCount() (int32, error)
	PageInfo() *graphqlutil.PageInfo
}

type RecordEventArgs struct{ Event TelemetryEventInput }
type RecordEventsArgs struct{ Events []TelemetryEventInput }

type TelemetryEventInput struct {
	Timestamp         *gqlutil.DateTime                     `json:"timestamp"`
	Feature           string                                `json:"feature"`
	Action            string                                `json:"action"`
	Source            TelemetryEventSourceInput             `json:"source"`
	Parameters        TelemetryEventParametersInput         `json:"parameters"`
	MarketingTracking *TelemetryEventMarketingTrackingInput `json:"marketingTracking,omitempty"`
}

type TelemetryEventSourceInput struct {
	Client        string  `json:"client"`
	ClientVersion *string `json:"clientVersion,omitempty"`
}

type TelemetryEventParametersInput struct {
	Version         int32                               `json:"version"`
	Metadata        *[]TelemetryEventMetadataInput      `json:"metadata,omitempty"`
	PrivateMetadata *JSONValue                          `json:"privateMetadata,omitempty"`
	BillingMetadata *TelemetryEventBillingMetadataInput `json:"billingMetadata,omitempty"`
	InteractionID   *string                             `json:"interactionID,omitempty"`
}

type TelemetryEventMetadataInput struct {
	Key   string    `json:"key"`
	Value JSONValue `json:"value"`
}

type TelemetryEventBillingMetadataInput struct {
	Product  string `json:"product"`
	Category string `json:"category"`
}

type TelemetryEventMarketingTrackingInput struct {
	Url             *string `json:"url,omitempty"`
	FirstSourceURL  *string `json:"firstSourceURL,omitempty"`
	CohortID        *string `json:"cohortID,omitempty"`
	Referrer        *string `json:"referrer,omitempty"`
	LastSourceURL   *string `json:"lastSourceURL,omitempty"`
	DeviceSessionID *string `json:"deviceSessionID,omitempty"`
	SessionReferrer *string `json:"sessionReferrer,omitempty"`
	SessionFirstURL *string `json:"sessionFirstURL,omitempty"`
}
