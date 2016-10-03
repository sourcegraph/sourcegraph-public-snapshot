package httpapi

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

type RepoSearchResult struct {
	*sourcegraph.Repo
}

type DefSearchResult struct {
	sourcegraph.Def
	RefCount int32
	Score    float32
}

type SearchOptions struct {
	Kinds     []string
	Languages []string
}

func serveGlobalSearch(w http.ResponseWriter, r *http.Request) error {
	var params struct {
		Query        string
		Repos        []string
		NotRepos     []string
		Limit        int32
		IncludeRepos bool
		Fast         bool
	}
	if err := schemaDecoder.Decode(&params, r.URL.Query()); err != nil {
		return err
	}

	if params.Limit == 0 {
		params.Limit = 100
	}

	paramsRepos, err := resolveLocalRepos(r.Context(), params.Repos, true)
	if err != nil {
		return err
	}

	paramsNotRepos, err := resolveLocalRepos(r.Context(), params.NotRepos, true)
	if err != nil {
		return err
	}

	op := &sourcegraph.SearchOp{
		Query: params.Query,
		Opt: &sourcegraph.SearchOptions{
			Repos:        paramsRepos,
			NotRepos:     paramsNotRepos,
			ListOptions:  sourcegraph.ListOptions{PerPage: params.Limit},
			IncludeRepos: params.IncludeRepos,
			Fast:         params.Fast,
		},
	}

	results, err := backend.Search.Search(r.Context(), op)
	if err != nil {
		return err
	}
	repos := make([]*RepoSearchResult, 0, len(results.RepoResults))
	for _, r := range results.RepoResults {
		repos = append(repos, &RepoSearchResult{
			Repo: r.Repo,
		})
	}

	var defs []*DefSearchResult
	for _, r := range results.DefResults {
		if r.Def.CommitID == "" {
			r.Def.CommitID = "master" // HACK
		}
		defs = append(defs, &DefSearchResult{
			Def:      r.Def,
			RefCount: r.RefCount,
			Score:    r.Score,
		})
	}

	var options []*SearchOptions
	for _, r := range results.SearchQueryOptions {
		options = append(options, &SearchOptions{
			Kinds:     r.Kinds,
			Languages: r.Languages,
		})
	}

	return json.NewEncoder(w).Encode(struct {
		Repos   []*RepoSearchResult
		Defs    []*DefSearchResult
		Options []*SearchOptions
	}{
		Repos:   repos,
		Defs:    defs,
		Options: options,
	})
}
