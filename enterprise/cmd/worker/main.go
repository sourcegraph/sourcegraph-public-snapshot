package main

import (
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/indexing"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/janitor"
)

var setupHooks = map[string]shared.SetupHook{
	"codeintel-commitgraph": commitgraph.NewInitializer(),
	"codeintel-janitor":     janitor.NewInitializer(),
	"codeintel-indexing":    indexing.NewInitializer(),
}

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	shared.Main(setupHooks)
}
