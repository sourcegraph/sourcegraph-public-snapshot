package upload

import (
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Client interface {
	// Do runs an http.Request against the Sourcegraph API.
	Do(req *http.Request) (*http.Response, error)
}

type UploadOptions struct {
	SourcegraphInstanceOptions
	OutputOptions
	UploadRecordOptions
}

type SourcegraphInstanceOptions struct {
	SourcegraphURL      string            // The URL (including scheme) of the target Sourcegraph instance
	AccessToken         string            // The user access token
	AdditionalHeaders   map[string]string // Additional request headers on each request
	Path                string            // Custom path on the Sourcegraph instance (used internally)
	MaxRetries          int               // The maximum number of retries per request
	RetryInterval       time.Duration     // Sleep duration between retries
	MaxPayloadSizeBytes int64             // The maximum number of bytes sent in a single request
	MaxConcurrency      int               // The maximum number of concurrent uploads. Only relevant for multipart uploads
	GitHubToken         string            // GitHub token used for auth when lsif.enforceAuth is true (optional)
	GitLabToken         string            // GitLab token used for auth when lsif.enforceAuth is true (optional)
	HTTPClient          Client
}

type OutputOptions struct {
	Logger RequestLogger  // Logger of all HTTP request/responses (optional)
	Output *output.Output // Output instance used for fancy output (optional)
}

type UploadRecordOptions struct {
	Repo              string
	Commit            string
	Root              string
	Indexer           string
	IndexerVersion    string
	AssociatedIndexID *int
}
