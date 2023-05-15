package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/server/shared"

	_ "github.com/sourcegraph/sourcegraph/ui/assets/oss" // Select oss assets
)

func main() {
	shared.Main()
}
