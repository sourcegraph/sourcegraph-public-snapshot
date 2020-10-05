package graphs

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
)

func (r *Resolver) RepositoriesForGraph(ctx context.Context, id graphql.ID) ([]string, error) {
	if err := graphsEnabled(); err != nil {
		return nil, err
	}

	graphID, err := unmarshalGraphID(id)
	if err != nil {
		return nil, err
	}

	if graphID == 0 {
		return nil, nil
	}

	graph, err := r.store.GetGraph(ctx, GetGraphOpts{ID: graphID})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(graph.Spec, "\n")
	var repos []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		const plusDepsSuffix = " + deps"
		plusDeps := false
		if strings.HasSuffix(line, plusDepsSuffix) {
			plusDeps = true
			line = strings.TrimSuffix(line, plusDepsSuffix)
		}

		repoName := line
		repos = append(repos, repoName)

		if plusDeps {
			switch repoName {
			case "github.com/hashicorp/go-multierror":
				repos = append(repos, "github.com/hashicorp/errwrap")
			}
		}
	}

	return repos, nil
}
