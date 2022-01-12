package compute

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type Replace struct {
	MatchPattern   MatchPattern
	ReplacePattern string
}

func (c *Replace) String() string {
	return fmt.Sprintf("Replace in place: (%s) -> (%s)", c.MatchPattern.String(), c.ReplacePattern)
}

func replace(ctx context.Context, content []byte, matchPattern MatchPattern, replacePattern string) (*Text, error) {
	var newContent string
	switch match := matchPattern.(type) {
	case *Regexp:
		newContent = match.Value.ReplaceAllString(string(content), replacePattern)
	case *Comby:
		replacements, err := comby.Replacements(ctx, comby.Args{
			Input:           comby.FileContent(content),
			MatchTemplate:   match.Value,
			RewriteTemplate: replacePattern,
			Matcher:         ".generic", // TODO(rvantonder): use language or file filter
			ResultKind:      comby.Replacement,
			NumWorkers:      0, // Just a single file's content.
		})
		if err != nil {
			return nil, err
		}
		// There is only one replacement value since we passed in comby.FileContent.
		newContent = replacements[0].Content
	default:
		return nil, errors.Errorf("unsupported replacement operation for match pattern %T", match)
	}
	return &Text{Value: newContent, Kind: "replace-in-place"}, nil
}

func (c *Replace) Run(ctx context.Context, r result.Match) (Result, error) {
	switch m := r.(type) {
	case *result.FileMatch:
		content, err := git.ReadFile(ctx, m.Repo.Name, m.CommitID, m.Path, 0, authz.DefaultSubRepoPermsChecker)
		if err != nil {
			return nil, err
		}
		return replace(ctx, content, c.MatchPattern, c.ReplacePattern)
	}
	return nil, nil
}
