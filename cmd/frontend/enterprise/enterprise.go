package enterprise

import "net/http"

// Services is a bag of HTTP handlers and factory functions that are registered by the
// enterprise frontend setup hook.
type Services struct {
	GithubWebhook             http.Handler
	BitbucketServerWebhook    http.Handler
	NewCodeIntelUploadHandler NewCodeIntelUploadHandler
}

// NewCodeIntelUploadHandler creates a new handler for the LSIF upload endpoint. The
// resulting handler skips auth checks when the internal flag is true.
type NewCodeIntelUploadHandler func(internal bool) http.Handler

// defaultNewCodeIntelUploadHandler creates a new canned handler that responds 404 to
// all requests.
var defaultNewCodeIntelUploadHandler = func(_ bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("codeintel upload is only available in enterprise"))
	})
}

// DefaultServices creates a new Services value that has default implementations for the
// values which are expected to be non-nil (e.g. factory functions).
func DefaultServices() Services {
	return Services{
		GithubWebhook:             nil,
		BitbucketServerWebhook:    nil,
		NewCodeIntelUploadHandler: defaultNewCodeIntelUploadHandler,
	}
}
