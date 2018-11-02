// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// AppURL is the fully-resolved frontend app URL.
var AppURL = &url.URL{Scheme: "http", Host: "example.com"}

// ConfigurationServerFrontendOnly provides the contents of the site configuration
// to other services and manages modifications to it.
//
// Any another service that attempts to use this variable will panic.
var ConfigurationServerFrontendOnly = conf.InitConfigurationServerFrontendOnly()
