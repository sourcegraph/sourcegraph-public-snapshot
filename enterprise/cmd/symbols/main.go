package main

import (
	"log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func main() {
	shared.Main(setup)
}

func setup(config *shared.Config, observationContext *observation.Context) (types.SearchFunc, []goroutine.BackgroundRoutine) {
	if !config.UseRockskip {
		return nil, nil
	}

	searchFunc, err := api.MakeRockskipSearchFunc(observationContext, config.Ctags, config.MaxTotalPathsLength, config.MaxRepos)
	if err != nil {
		log.Fatalf("Failed to create rockskip search function: %s", err)
	}

	return searchFunc, nil
}
