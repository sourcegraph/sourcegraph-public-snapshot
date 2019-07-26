package main

import (
	_ "sourcegraph.com/cmd/frontend/internal/app/assets"
	_ "sourcegraph.com/cmd/frontend/registry"
	"sourcegraph.com/cmd/frontend/shared"
)

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

func main() {
	shared.Main()
}
