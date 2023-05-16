// Package hooks allow hooking into the frontend.
package hooks

import (
	"net/http"
)

// PostAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is authenticated when PostAuthMiddleware is called.
var PostAuthMiddleware func(http.Handler) http.Handler

// FeatureBatchChanges describes if and how the Batch Changes feature is available on
// the given license plan. It mirrors the type licensing.FeatureBatchChanges.
type FeatureBatchChanges struct {
	// If true, there is no limit to the number of changesets that can be created.
	Unrestricted bool `json:"unrestricted"`
	// Maximum number of changesets that can be created per batch change.
	// If Unrestricted is true, this is ignored.
	MaxNumChangesets int `json:"maxNumChangesets"`
}

// LicenseInfo contains information about the legitimate usage of the current
// license on the instance.
type LicenseInfo struct {
	CurrentPlan string `json:"currentPlan"`

	CodeScaleLimit         string               `json:"codeScaleLimit"`
	CodeScaleCloseToLimit  bool                 `json:"codeScaleCloseToLimit"`
	CodeScaleExceededLimit bool                 `json:"codeScaleExceededLimit"`
	KnownLicenseTags       []string             `json:"knownLicenseTags"`
	BatchChanges           *FeatureBatchChanges `json:"batchChanges"`
}

var GetLicenseInfo = func(isSiteAdmin bool) *LicenseInfo { return nil }
