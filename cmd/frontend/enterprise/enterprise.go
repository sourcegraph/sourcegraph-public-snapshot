package enterprise

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Services is a bag of HTTP handlers and factory functions that are registered by the
// enterprise frontend setup hook.
type Services struct {
	GitHubWebhook             http.Handler
	GitLabWebhook             http.Handler
	BitbucketServerWebhook    http.Handler
	NewCodeIntelUploadHandler NewCodeIntelUploadHandler
	NewExecutorProxyHandler   NewExecutorProxyHandler
	AuthzResolver             graphqlbackend.AuthzResolver
	CampaignsResolver         graphqlbackend.CampaignsResolver
	CodeIntelResolver         graphqlbackend.CodeIntelResolver
	CodeMonitorsResolver      graphqlbackend.CodeMonitorsResolver
}

// NewCodeIntelUploadHandler creates a new handler for the LSIF upload endpoint. The
// resulting handler skips auth checks when the internal flag is true.
type NewCodeIntelUploadHandler func(internal bool) http.Handler

// NewExecutorProxyHandler creates a new proxy handler for routes accessible to the
// executor services deployed separately from the k8s cluster. This handler is protected
// via a shared username and password.
type NewExecutorProxyHandler func() http.Handler

// DefaultServices creates a new Services value that has default implementations for all services.
func DefaultServices() Services {
	return Services{
		GitHubWebhook:             makeNotFoundHandler("github webhook"),
		GitLabWebhook:             makeNotFoundHandler("gitlab webhook"),
		BitbucketServerWebhook:    makeNotFoundHandler("bitbucket server webhook"),
		NewCodeIntelUploadHandler: func(_ bool) http.Handler { return makeNotFoundHandler("code intel upload") },
		NewExecutorProxyHandler:   func() http.Handler { return makeNotFoundHandler("executor proxy") },
		AuthzResolver:             graphqlbackend.DefaultAuthzResolver,
		CampaignsResolver:         graphqlbackend.DefaultCampaignsResolver,
		CodeMonitorsResolver:      graphqlbackend.DefaultCodeMonitorsResolver,
	}
}

// makeNotFoundHandler returns an HTTP handler that respond 404 for all requests.
func makeNotFoundHandler(handlerName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("%s is only available in enterprise", handlerName)))
	})
}
