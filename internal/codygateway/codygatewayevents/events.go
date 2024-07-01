package codygatewayevents

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

const CompletionsEventFeatureMetadataField = "feature"
const EmbeddingsTokenUsageMetadataField = "tokens_used"

const CompletionsEventFeatureEmbeddings = "embeddings"
