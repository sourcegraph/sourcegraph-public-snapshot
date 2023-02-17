package apiclient

const (
	// HeaderActorUID is the HTTP header key for the actor UID.
	HeaderActorUID = "X-Sourcegraph-Actor-UID"
	// HeaderAuthorization is the HTTP header key for the authorization token.
	HeaderAuthorization = "Authorization"
	// HeaderExecutorName is the HTTP header key for the executor name.
	HeaderExecutorName = "X-Sourcegraph-Executor-Name"
	// HeaderJobID is the HTTP header key for the job ID.
	HeaderJobID = "X-Sourcegraph-Job-ID"
)

const (
	// AuthenticationSchemeBearer is the authentication scheme for Job tokens.
	AuthenticationSchemeBearer = "Bearer"
	// AuthenticationSchemeExecutorToken is the authentication scheme for the general executor access token.
	AuthenticationSchemeExecutorToken = "token-executor"
)
