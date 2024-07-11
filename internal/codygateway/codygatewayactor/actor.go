package codygatewayactor

import "time"

type ActorSource string

const (
	// We retain legacy naming just in case there are hard dependencies on this
	// name. Today, these are Enterprise Subscriptions sourced from the Enterprise
	// Portal service.
	ActorSourceEnterpriseSubscription ActorSource = "dotcom-product-subscriptions"
	// Sourcegraph.com user actors.
	ActorSourceDotcomUser ActorSource = "dotcom-user"
)

// ActorConcurrencyLimitConfig is the configuration for the concurrent requests
// limit of an actor.
type ActorConcurrencyLimitConfig struct {
	// Percentage is the percentage of the daily rate limit to be used to compute the
	// concurrency limit.
	Percentage float32
	// Interval is the time interval of the limit bucket.
	Interval time.Duration
}

// ActorRateLimitNotifyConfig is the configuration for the rate limit
// notifications of an actor.
type ActorRateLimitNotifyConfig struct {
	// SlackWebhookURL is the URL of the Slack webhook to send the alerts to.
	SlackWebhookURL string
}
