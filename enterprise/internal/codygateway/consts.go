package codygateway

const ProductSubscriptionActorSourceName = "dotcom-product-subscriptions"

const CompletionsEventFeatureMetadataField = "feature"

type EventName string

const (
	EventNameUnauthorized        EventName = "Unauthorized"
	EventNameAccessDenied        EventName = "AccessDenied"
	EventNameRateLimited         EventName = "RateLimited"
	EventNameCompletionsStarted  EventName = "CompletionsStarted"
	EventNameCompletionsFinished EventName = "CompletionsFinished"
)

const FeatureHeaderName = "X-Sourcegraph-Feature"
