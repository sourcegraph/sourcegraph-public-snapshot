package backend

import (
	"strings"

	srch "sourcegraph.com/sourcegraph/sourcegraph/pkg/search"
	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

var Search sourcegraph.SearchServer = &search{}

type search struct{}

var _ sourcegraph.SearchServer = (*search)(nil)

var tokenToKind = map[string]string{
	"func":    "func",
	"method":  "func",
	"type":    "type",
	"struct":  "type",
	"class":   "type",
	"var":     "var",
	"field":   "field",
	"package": "package",
	"const":   "const",
}

func (s *search) Search(ctx context.Context, op *sourcegraph.SearchOp) (*sourcegraph.SearchResultsList, error) {
	var unit, unitType string
	var kinds []string
	var descToks []string                            // "descriptor" tokens that don't have a special filter meaning.
	for _, token := range strings.Fields(op.Query) { // at first tokenize on spaces
		if strings.HasPrefix(token, "r:") {
			op.Opt.Repos = append(op.Opt.Repos, strings.TrimPrefix(token, "r:"))
			continue
		}
		if strings.HasPrefix(token, "u:") {
			unit = strings.TrimPrefix(token, "u:")
			continue
		}
		if strings.HasPrefix(token, "t:") {
			unit = strings.TrimPrefix(token, "t:")
			continue
		}
		if kind, exist := tokenToKind[token]; exist {
			kinds = append(kinds, kind)
			continue
		}

		// function shorthand, still include token as a descriptor token
		if strings.HasSuffix(token, "()") {
			kinds = append(kinds, "func")
		}

		if strings.HasSuffix(token, ".com") || strings.HasSuffix(token, ".org") {
			descToks = append(descToks, token)
		} else {
			descToks = append(descToks, srch.QueryTokens(token)...)
		}
	}

	results, err := store.GlobalDefsFromContext(ctx).Search(ctx, &store.GlobalDefSearchOp{
		UnitQuery:     unit,
		UnitTypeQuery: unitType,

		TokQuery: descToks,
		Opt:      op.Opt,
	})
	if err != nil {
		return nil, err
	}
	for _, r := range results.DefResults {
		populateDefFormatStrings(&r.Def)
	}

	if !op.Opt.IncludeRepos {
		return results, nil
	}

	results.RepoResults, err = store.ReposFromContext(ctx).Search(ctx, op.Query)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *search) RefreshIndex(ctx context.Context, op *sourcegraph.SearchRefreshIndexOp) (*pbtypes.Void, error) {
	// Currently, the only pre-computation we do is aggregating the global ref counts
	// for every def. This will pre-compute the ref counts based on the current state
	// of the GlobalRefs table for all defs in the given repos.
	var updateOp store.GlobalDefUpdateOp
	for _, r := range op.Repos {
		updateOp.RepoUnits = append(updateOp.RepoUnits, store.RepoUnit{Repo: sourcegraph.RepoSpec{r.URI}})
	}

	if op.RefreshSearch {
		if err := store.GlobalDefsFromContext(ctx).Update(ctx, updateOp); err != nil {
			return nil, err
		}
	}

	if op.RefreshCounts {
		if err := store.GlobalDefsFromContext(ctx).RefreshRefCounts(ctx, updateOp); err != nil {
			return nil, err
		}
	}
	return &pbtypes.Void{}, nil
}
