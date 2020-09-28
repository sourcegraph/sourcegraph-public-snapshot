// Package app exports symbols from frontend/internal/app. See the parent
// package godoc for more information.
package app

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
)

func SetBillingPublishableKey(value string) {
	jscontext.BillingPublishableKey = value
}
