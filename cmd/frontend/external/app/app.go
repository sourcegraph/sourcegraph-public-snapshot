package app

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
)

type (
	SignOutURL = app.SignOutURL
)

var (
	RegisterSSOSignOutHandler = app.RegisterSSOSignOutHandler
)

func SetBillingPublishableKey(value string) {
	jscontext.BillingPublishableKey = value
}
