// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"net/url"
)

// AppURL is the fully-resolved frontend app URL.
var AppURL = &url.URL{Scheme: "http", Host: "example.com"}
