package backend

import (
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	srch "sourcegraph.com/sourcegraph/sourcegraph/pkg/search"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
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
	"pkg":     "package",
	"const":   "const",
}

var tokenToLanguage = map[string]string{
	"golang": "go",
	"java":   "java",
	"python": "python",
}

func (s *search) Search(ctx context.Context, op *sourcegraph.SearchOp) (*sourcegraph.SearchResultsList, error) {
	var unit, unitType string
	var kinds []string
	var descToks []string                            // "descriptor" tokens that don't have a special filter meaning.
	for _, token := range strings.Fields(op.Query) { // at first tokenize on spaces
		if strings.HasPrefix(token, "r:") {
			repoPath := strings.TrimPrefix(token, "r:")
			res, err := svc.Repos(ctx).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoPath})
			if err == nil {
				op.Opt.Repos = append(op.Opt.Repos, res.Repo)
			} else {
				log15.Warn("Search.Search: failed to resolve repo in query; ignoring.", "repo", repoPath, "err", err)
			}
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
		if kind, exist := tokenToKind[strings.ToLower(token)]; exist {
			op.Opt.Kinds = append(op.Opt.Kinds, kind)
			unit = ""
			continue
		}
		if lang, exist := tokenToLanguage[strings.ToLower(token)]; exist {
			op.Opt.Languages = append(op.Opt.Languages, lang)
			unit = ""
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

	// For global search analytics purposes
	results.SearchQueryOptions = []*sourcegraph.SearchOptions{op.Opt}

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
	for _, repo := range op.Repos {
		updateOp.RepoUnits = append(updateOp.RepoUnits, store.RepoUnit{Repo: repo})
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
