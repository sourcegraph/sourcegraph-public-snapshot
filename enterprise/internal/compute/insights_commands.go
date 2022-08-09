package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type InsightsCount struct {
	SearchPattern string
	OutputPattern string
}

func (c *InsightsCount) ToSearchPattern() string {
	return c.SearchPattern
}

func (c *InsightsCount) String() string {
	return fmt.Sprintf("insightsCount search pattern: %s", c.SearchPattern)
}

func (c *InsightsCount) Run(ctx context.Context, db database.DB, r result.Match) (Result, error) {
	return insightsCount(ctx, c.OutputPattern, r)
}

func insightsCount(ctx context.Context, outputPattern string, r result.Match) (Result, error) {
	countFunc, ok := insightCommandOutputRegistry[strings.ToUpper(outputPattern)]
	if !ok {
		return nil, errors.New("unknown ouput pattern for insightsCount command")
	}

	return countFunc(r)
}

type insightsCountFunc func(result.Match) (Result, error)

// Map of functions that resolves a count for all available output options
var insightCommandOutputRegistry map[string]insightsCountFunc = map[string]insightsCountFunc{
	"$REPO":   countRepo,
	"$LANG":   countLang,
	"$PATH":   countPath,
	"$AUTHOR": countAuthor,
}

func countRepo(r result.Match) (Result, error) {
	if r.RepoName().Name != "" {
		return &InsightsCountResult{Value: string(r.RepoName().Name), Count: r.ResultCount()}, nil
	}
	return nil, nil
}

func countLang(r result.Match) (Result, error) {
	env := NewMetaEnvironment(r, "")
	if env.Lang != "" {
		return &InsightsCountResult{Value: env.Lang, Count: r.ResultCount()}, nil
	}
	return nil, nil
}

func countPath(r result.Match) (Result, error) {
	env := NewMetaEnvironment(r, "")
	if env.Path != "" {
		return &InsightsCountResult{Value: env.Path, Count: r.ResultCount()}, nil
	}
	return nil, nil
}

func countAuthor(r result.Match) (Result, error) {
	env := NewMetaEnvironment(r, "")
	if env.Author != "" {
		return &InsightsCountResult{Value: env.Author, Count: r.ResultCount()}, nil
	}
	return nil, nil
}
