package main

import (
	"log"

	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
)

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

func main() {
	log.Println("XXXXXXXXXXXXXXXXXXXXX")

	shared.Main()
}
