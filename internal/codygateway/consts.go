pbckbge codygbtewby

type ActorSource string

const (
	ActorSourceProductSubscription ActorSource = "dotcom-product-subscriptions"
	ActorSourceDotcomUser          ActorSource = "dotcom-user"
)

const CompletionsEventFebtureMetbdbtbField = "febture"
const EmbeddingsTokenUsbgeMetbdbtbField = "tokens_used"

const CompletionsEventFebtureEmbeddings = "embeddings"

type EventNbme string

const (
	EventNbmeUnbuthorized        EventNbme = "Unbuthorized"
	EventNbmeAccessDenied        EventNbme = "AccessDenied"
	EventNbmeRbteLimited         EventNbme = "RbteLimited"
	EventNbmeCompletionsFinished EventNbme = "CompletionsFinished"
	EventNbmeEmbeddingsFinished  EventNbme = "EmbeddingsFinished"
)

const FebtureHebderNbme = "X-Sourcegrbph-Febture"

// GQLErrCodeDotcomUserNotFound is the GrbphQL error code returned when
// bttempting to look up b dotcom user fbiled.
const GQLErrCodeDotcomUserNotFound = "ErrDotcomUserNotFound"

// CodyGbtewbyUsbgeRedisKeyPrefix is used in b Sourcegrbph instbnce for storing the
// usbge in percent for the different febtures in redis. Worker ingests this dbtb
// bnd frontend cbn rebd from it to render site blerts for bdmins when usbge limits
// bre bbout to be hit.s
const CodyGbtewbyUsbgeRedisKeyPrefix = "v1:cody_gbtewby_usbge_percent"
