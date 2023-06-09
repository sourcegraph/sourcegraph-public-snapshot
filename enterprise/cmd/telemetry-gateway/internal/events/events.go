package events

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type TopicConfig struct {
	ProjectName string
	TopicName   string
}

type ClientInfo struct {
	SiteId string
}

type TelemetryGatewayProxyRequest struct {
	Client ClientInfo
	Events TelemetryEvent
}

type TelemetryEvent struct {
	SiteID            string  `json:"site_id"`
	LicenseKey        string  `json:"license_key"`
	InitialAdminEmail string  `json:"initial_admin_email"`
	DeployType        string  `json:"deploy_type"`
	EventName         string  `json:"name"`
	URL               string  `json:"url"`
	AnonymousUserID   string  `json:"anonymous_user_id"`
	FirstSourceURL    string  `json:"first_source_url"`
	LastSourceURL     string  `json:"last_source_url"`
	UserID            int     `json:"user_id"`
	Source            string  `json:"source"`
	Timestamp         string  `json:"timestamp"`
	Version           string  `json:"Version"`
	FeatureFlags      string  `json:"feature_flags"`
	CohortID          *string `json:"cohort_id,omitempty"`
	Referrer          string  `json:"referrer,omitempty"`
	PublicArgument    string  `json:"public_argument"`
	DeviceID          *string `json:"device_id,omitempty"`
	InsertID          *string `json:"insert_id,omitempty"`
}

func SendEvents(ctx context.Context, request TelemetryGatewayProxyRequest, config TopicConfig) error {
	client, err := pubsub.NewClient(ctx, config.ProjectName)
	if err != nil {
		return errors.Wrap(err, "pubsub.NewClient")
	}
	defer client.Close()

	marshal, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	topic := client.Topic(config.TopicName)
	defer topic.Stop()
	masg := &pubsub.Message{
		Data: marshal,
	}
	result := topic.Publish(ctx, masg)
	_, err = result.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "result.Get")
	}

	return nil
}
