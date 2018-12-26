package graphqlbackend

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
)

var mockSearchRepositories func(args *search.Args) ([]*searchResultResolver, *searchResultsCommon, error)
var repoIcon = "data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgNjQgNjQiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDY0IDY0OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+Cjx0aXRsZT5JY29ucyA0MDA8L3RpdGxlPgo8Zz4KCTxwYXRoIGQ9Ik0yMywyMi40YzEuMywwLDIuNC0xLjEsMi40LTIuNHMtMS4xLTIuNC0yLjQtMi40Yy0xLjMsMC0yLjQsMS4xLTIuNCwyLjRTMjEuNywyMi40LDIzLDIyLjR6Ii8+Cgk8cGF0aCBkPSJNMzUsMjYuNGMxLjMsMCwyLjQtMS4xLDIuNC0yLjRzLTEuMS0yLjQtMi40LTIuNHMtMi40LDEuMS0yLjQsMi40UzMzLjcsMjYuNCwzNSwyNi40eiIvPgoJPHBhdGggZD0iTTIzLDQyLjRjMS4zLDAsMi40LTEuMSwyLjQtMi40cy0xLjEtMi40LTIuNC0yLjRzLTIuNCwxLjEtMi40LDIuNFMyMS43LDQyLjQsMjMsNDIuNHoiLz4KCTxwYXRoIGQ9Ik01MCwxNmgtMS41Yy0wLjMsMC0wLjUsMC4yLTAuNSwwLjV2MzVjMCwwLjMtMC4yLDAuNS0wLjUsMC41aC0yN2MtMC41LDAtMS0wLjItMS40LTAuNmwtMC42LTAuNmMtMC4xLTAuMS0wLjEtMC4yLTAuMS0wLjQKCQljMC0wLjMsMC4yLTAuNSwwLjUtMC41SDQ0YzEuMSwwLDItMC45LDItMlYxMmMwLTEuMS0wLjktMi0yLTJIMTRjLTEuMSwwLTIsMC45LTIsMnYzNi4zYzAsMS4xLDAuNCwyLjEsMS4yLDIuOGwzLjEsMy4xCgkJYzEuMSwxLjEsMi43LDEuOCw0LjIsMS44SDUwYzEuMSwwLDItMC45LDItMlYxOEM1MiwxNi45LDUxLjEsMTYsNTAsMTZ6IE0xOSwyMGMwLTIuMiwxLjgtNCw0LTRjMS40LDAsMi44LDAuOCwzLjUsMgoJCWMxLjEsMS45LDAuNCw0LjMtMS41LDUuNFYzM2MxLTAuNiwyLjMtMC45LDQtMC45YzEsMCwyLTAuNSwyLjgtMS4zQzMyLjUsMzAsMzMsMjkuMSwzMywyOHYtMC42Yy0xLjItMC43LTItMi0yLTMuNQoJCWMwLTIuMiwxLjgtNCw0LTRjMi4yLDAsNCwxLjgsNCw0YzAsMS41LTAuOCwyLjctMiwzLjVoMGMtMC4xLDIuMS0wLjksNC40LTIuNSw2Yy0xLjYsMS42LTMuNCwyLjQtNS41LDIuNWMtMC44LDAtMS40LDAuMS0xLjksMC4zCgkJYy0wLjIsMC4xLTEsMC44LTEuMiwwLjlDMjYuNiwzOCwyNywzOC45LDI3LDQwYzAsMi4yLTEuOCw0LTQsNHMtNC0xLjgtNC00YzAtMS41LDAuOC0yLjcsMi0zLjRWMjMuNEMxOS44LDIyLjcsMTksMjEuNCwxOSwyMHoiLz4KPC9nPgo8L3N2Zz4K"

// searchRepositories searches for repositories by name.
//
// For a repository to match a query, the repository's name must match all of the repo: patterns AND the
// default patterns (i.e., the patterns that are not prefixed with any search field).
func searchRepositories(ctx context.Context, args *search.Args, limit int32) (res []*searchResultResolver, common *searchResultsCommon, err error) {
	if mockSearchRepositories != nil {
		return mockSearchRepositories(args)
	}

	fieldWhitelist := map[string]struct{}{
		query.FieldRepo:      {},
		query.FieldRepoGroup: {},
		query.FieldType:      {},
		query.FieldDefault:   {},
		query.FieldIndex:     {},
		query.FieldCount:     {},
		query.FieldMax:       {},
		query.FieldTimeout:   {},
		query.FieldFork:      {},
		query.FieldArchived:  {},
	}
	// Don't return repo results if the search contains fields that aren't on the whitelist.
	// Matching repositories based whether they contain files at a certain path (etc.) is not yet implemented.
	for field := range args.Query.Fields {
		if _, ok := fieldWhitelist[field]; !ok {
			return nil, nil, nil
		}
	}

	pattern, err := regexp.Compile(args.Pattern.Pattern)
	if err != nil {
		return nil, nil, err
	}

	common = &searchResultsCommon{}
	var results []*searchResultResolver
	for _, repo := range args.Repos {
		if len(results) == int(limit) {
			common.limitHit = true
			break
		}
		if pattern.MatchString(string(repo.Repo.Name)) {
			results = append(results, &searchResultResolver{repo: &repositoryResolver{repo: repo.Repo, icon: repoIcon}})
		}
	}
	return results, common, nil
}
