package llmproxy

const ProductSubscriptionActorSourceName = "dotcom-product-subscriptions"

type EventName string

const (
	EventNameUnauthorized        EventName = "Unauthorized"
	EventNameAccessDenied        EventName = "AccessDenied"
	EventNameRateLimited         EventName = "RateLimited"
	EventNameCompletionsStarted  EventName = "CompletionsStarted"
	EventNameCompletionsFinished EventName = "CompletionsFinished"
)
