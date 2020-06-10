package enterprise

import "net/http"

// GithubWebhook is re-assigned by the enterprise frontend.
var GithubWebhook http.Handler = nil

// BitbucketServerWebhook is re-assigned by the enterprise frontend.
var BitbucketServerWebhook http.Handler = nil
