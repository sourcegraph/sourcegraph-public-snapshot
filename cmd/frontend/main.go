package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
)

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

func main() {
	go authz.SetProviders(true, []authz.Provider{})
	shared.Main(nil, nil)
}
