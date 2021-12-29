package compute

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type Output struct {
	MatchPattern  MatchPattern
	OutputPattern string
	Separator     string
}

func (c *Output) String() string {
	return fmt.Sprintf("Output with separator: (%s) -> (%s) separator: %s", c.MatchPattern.String(), c.OutputPattern, c.Separator)
}

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	var b strings.Builder
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		b.Write(match.ExpandString([]byte{}, replacePattern, content, submatches))
		b.WriteString(separator)
	}
	return b.String()
}

func output(ctx context.Context, fragment string, matchPattern MatchPattern, replacePattern string, separator string) (*Text, error) {
	var newContent string
	var err error
	switch match := matchPattern.(type) {
	case *Regexp:
		newContent = substituteRegexp(fragment, match.Value, replacePattern, separator)
	case *Comby:
		newContent, err = comby.Outputs(ctx, comby.Args{
			Input:           comby.FileContent(fragment),
			MatchTemplate:   match.Value,
			RewriteTemplate: replacePattern,
			Matcher:         ".generic", // TODO(rvantoner): use language or file filter
			ResultKind:      comby.NewlineSeparatedOutput,
			NumWorkers:      0,
		})
		if err != nil {
			return nil, err
		}

	}
	return &Text{Value: newContent, Kind: "output"}, nil
}

func resultContent(ctx context.Context, r result.Match) (string, bool, error) {
	switch m := r.(type) {
	case *result.FileMatch:
		contentBytes, err := git.ReadFile(ctx, m.Repo.Name, m.CommitID, m.Path, 0, authz.DefaultSubRepoPermsChecker)
		if err != nil {
			return "", false, err
		}
		return string(contentBytes), true, nil
	case *result.CommitMatch:
		var content string
		if m.DiffPreview != nil {
			content = m.DiffPreview.Value
		} else {
			content = string(m.Commit.Message)
		}
		return content, true, nil
	default:
		return "", false, nil
	}
}

func (c *Output) Run(ctx context.Context, r result.Match) (Result, error) {
	content, ok, err := resultContent(ctx, r)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	env := NewMetaEnvironment(r, content)
	outputPattern, err := substituteMetaVariables(c.OutputPattern, env)
	if err != nil {
		return nil, err
	}
	return output(ctx, content, c.MatchPattern, outputPattern, c.Separator)
}
