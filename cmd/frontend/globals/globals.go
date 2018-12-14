// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// ExternalURL is the fully-resolved, externally accessible frontend URL.
var ExternalURL = &url.URL{Scheme: "http", Host: "example.com"}

// ConfigurationServerFrontendOnly provides the contents of the site configuration
// to other services and manages modifications to it.
//
// Any another service that attempts to use this variable will panic.
var ConfigurationServerFrontendOnly *conf.Server
