package codygateway

const ProductSubscriptionActorSourceName = "dotcom-product-subscriptions"

const CompletionsEventFeatureMetadataField = "feature"
const EmbeddingsTokenUsageMetadataField = "tokens_used"

const CompletionsEventFeatureEmbeddings = "embeddings"

type EventName string

const (
	EventNameUnauthorized        EventName = "Unauthorized"
	EventNameAccessDenied        EventName = "AccessDenied"
	EventNameRateLimited         EventName = "RateLimited"
	EventNameCompletionsFinished EventName = "CompletionsFinished"
	EventNameEmbeddingsFinished  EventName = "EmbeddingsFinished"
)

const FeatureHeaderName = "X-Sourcegraph-Feature"

// GQLErrCodeProductSubscriptionNotFound is the GraphQL error code returned when
// attempting to look up a product subscription failed by any means.
const GQLErrCodeProductSubscriptionNotFound = "ErrProductSubscriptionNotFound"
