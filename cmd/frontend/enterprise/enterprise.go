package enterprise

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Services is a bag of HTTP handlers and factory functions that are registered by the
// enterprise frontend setup hook.
type Services struct {
	GithubWebhook             http.Handler
	BitbucketServerWebhook    http.Handler
	NewCodeIntelUploadHandler NewCodeIntelUploadHandler
	AuthzResolver             graphqlbackend.AuthzResolver
	CampaignsResolver         graphqlbackend.CampaignsResolver
	CodeIntelResolver         graphqlbackend.CodeIntelResolver
}

// NewCodeIntelUploadHandler creates a new handler for the LSIF upload endpoint. The
// resulting handler skips auth checks when the internal flag is true.
type NewCodeIntelUploadHandler func(internal bool) http.Handler

// DefaultServices creates a new Services value that has default implementations for all services.
func DefaultServices() Services {
	return Services{
		GithubWebhook:             makeHandler("github webhook"),
		BitbucketServerWebhook:    makeHandler("bitbucket server webhook"),
		NewCodeIntelUploadHandler: func(_ bool) http.Handler { return makeHandler("code intel upload") },
		AuthzResolver:             graphqlbackend.DefaultAuthzResolver,
		CampaignsResolver:         graphqlbackend.DefaultCampaignsResolver,
		CodeIntelResolver:         graphqlbackend.DefaultCodeIntelResolver,
	}
}

// makeHandler returns an HTTP handler that respond 404 for all requests.
func makeHandler(handlerName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("%s is only available in enterprise", handlerName)))
	})
}
