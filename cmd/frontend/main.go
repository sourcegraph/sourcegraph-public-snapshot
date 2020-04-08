package main

import (
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
)

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

func main() {
	shared.Main(nil, nil)
}
