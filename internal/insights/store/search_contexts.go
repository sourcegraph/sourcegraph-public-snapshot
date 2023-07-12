package store

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	sctypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SearchContextLoader loads search contexts just from the full name of the
// context. This will not verify that the calling context owns the context, it
// will load regardless of the current user.
type SearchContextLoader interface {
	GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error)
}

type scLoader struct {
	primary database.DB
}

func (l *scLoader) GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error) {
	return searchcontexts.ResolveSearchContextSpec(ctx, l.primary, name)
}

type SearchContextHandler struct {
	loader SearchContextLoader
}

func NewSearchContextHandler(db database.DB) *SearchContextHandler {
	return &SearchContextHandler{loader: &scLoader{db}}
}

func (h *SearchContextHandler) UnwrapSearchContexts(ctx context.Context, rawContexts []string) ([]string, []string, error) {
	var include []string
	var exclude []string

	for _, rawContext := range rawContexts {
		searchContext, err := h.loader.GetByName(ctx, rawContext)
		if err != nil {
			return nil, nil, err
		}
		if searchContext.Query != "" {
			var plan searchquery.Plan
			plan, err := searchquery.Pipeline(
				searchquery.Init(searchContext.Query, searchquery.SearchTypeRegex),
			)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to parse search query for search context: %s", rawContext)
			}
			inc, exc := plan.ToQ().Repositories()
			for _, repoFilter := range inc {
				if len(repoFilter.Revs) > 0 {
					return nil, nil, errors.Errorf("search context filters cannot include repo revisions: %s", rawContext)
				}
				include = append(include, repoFilter.Repo)
			}
			exclude = append(exclude, exc...)
		}
	}
	return include, exclude, nil
}
