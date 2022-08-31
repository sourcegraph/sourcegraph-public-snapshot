// Package hooks allow hooking into the frontend.
package hooks

import (
	"net/http"
)

// PostAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is authenticated when PostAuthMiddleware is called.
var PostAuthMiddleware func(http.Handler) http.Handler

// LicenseInfo contains information about the legitimate usage of the current
// license on the instance.
type LicenseInfo struct {
	CurrentPlan string `json:"currentPlan"`

	CodeScaleLimit         string `json:"codeScaleLimit"`
	CodeScaleCloseToLimit  bool   `json:"codeScaleCloseToLimit"`
	CodeScaleExceededLimit bool   `json:"codeScaleExceededLimit"`
}

var GetLicenseInfo = func(isSiteAdmin bool) *LicenseInfo { return nil }
