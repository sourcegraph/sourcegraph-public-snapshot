package codygateway

type ActorSource string

const (
	ActorSourceProductSubscription ActorSource = "dotcom-product-subscriptions"
	ActorSourceDotcomUser          ActorSource = "dotcom-user"
)

const CompletionsEventFeatureMetadataField = "feature"
const EmbeddingsTokenUsageMetadataField = "tokens_used"

const CompletionsEventFeatureEmbeddings = "embeddings"

type EventName string

const (
	EventNameUnauthorized         EventName = "Unauthorized"
	EventNameAccessDenied         EventName = "AccessDenied"
	EventNameRateLimited          EventName = "RateLimited"
	EventNameCompletionsFinished  EventName = "CompletionsFinished"
	EventNameEmbeddingsFinished   EventName = "EmbeddingsFinished"
	EventNameRequestBlocked       EventName = "RequestBlocked"
	EventNameCodeCompletionLogged EventName = "CodeCompletionLogged"
)

const FeatureHeaderName = "X-Sourcegraph-Feature"

// GQLErrCodeDotcomUserNotFound is the GraphQL error code returned when
// attempting to look up a dotcom user failed.
const GQLErrCodeDotcomUserNotFound = "ErrDotcomUserNotFound"

// CodyGatewayUsageRedisKeyPrefix is used in a Sourcegraph instance for storing the
// usage in percent for the different features in redis. Worker ingests this data
// and frontend can read from it to render site alerts for admins when usage limits
// are about to be hit.s
const CodyGatewayUsageRedisKeyPrefix = "v1:cody_gateway_usage_percent"
