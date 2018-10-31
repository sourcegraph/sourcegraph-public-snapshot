package extsvc

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type CodeHost interface {
	ServiceID() string
	ServiceType() string
	IsHostOf(repo *api.ExternalRepoSpec) bool
}

// NormalizeBaseURL modifies the input and returns a normalized form of the a base URL with insignificant
// differences (such as in presence of a trailing slash, or hostname case) eliminated. Its return value should be
// used for the (ExternalRepoSpec).ServiceID field (and passed to XyzExternalRepoSpec) instead of a non-normalized
// base URL.
func NormalizeBaseURL(baseURL *url.URL) *url.URL {
	baseURL.Host = strings.ToLower(baseURL.Host)
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return baseURL
}
