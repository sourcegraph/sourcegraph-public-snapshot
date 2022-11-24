package main

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/log"
)

const q = `query { currentUser { username } }`

func main() {
	ctx := context.Background()
	logfuncs := log.Init(log.Resource{
		Name: "executors test runner",
	})
	defer logfuncs.Sync()

	logger := log.Scoped("init", "runner initialization process")

	SourcegraphAccessToken = createSudoToken()

	if err := InitializeGraphQLClient(); err != nil {
		logger.Fatal("cannot initialize graphql client", log.Error(err))
	}

	res := map[string]any{}
	err := queryGraphQL(ctx, logger.Scoped("graphql", ""), "", q, nil, &res)
	if err != nil {
		logger.Fatal("graphql failed with", log.Error(err))
	}

	b, _ := json.MarshalIndent(res, "", "  ")
	println(string(b))
}
