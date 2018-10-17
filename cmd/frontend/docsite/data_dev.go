// +build !dist

package docsite

import (
	"net/http"
)

// content contains the Sourcegraph documentation content.
var content = http.Dir("doc") // top-level doc dir
