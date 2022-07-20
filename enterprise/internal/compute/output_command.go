package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Output struct {
	SearchPattern MatchPattern
	OutputPattern string
	Separator     string
	Selector      string
	TypeValue     string
	Kind          string
}

func (c *Output) ToSearchPattern() string {
	return c.SearchPattern.String()
}

func (c *Output) String() string {
	return fmt.Sprintf("Output with separator: (%s) -> (%s) separator: %s", c.SearchPattern.String(), c.OutputPattern, c.Separator)
}

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	var b strings.Builder
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		b.Write(match.ExpandString([]byte{}, replacePattern, content, submatches))
		b.WriteString(separator)
	}
	return b.String()
}

func output(ctx context.Context, fragment string, matchPattern MatchPattern, replacePattern string, separator string) (string, error) {
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
			return "", err
		}

	}
	return newContent, nil
}

func resultContent(r result.Match, onlyPath bool) string {
	switch m := r.(type) {
	case *result.RepoMatch:
		return string(m.Name)
	case *result.FileMatch:
		if onlyPath {
			return m.Path
		}
		var sb strings.Builder
		for _, cm := range m.ChunkMatches {
			for _, range_ := range cm.Ranges {
				sb.WriteString(chunkContent(cm, range_))
			}
		}
		return sb.String()
	case *result.CommitDiffMatch:
		var sb strings.Builder
		for _, h := range m.Hunks {
			for _, l := range h.Lines {
				sb.WriteString(l)
			}
		}
		return sb.String()
	case *result.CommitMatch:
		var content string
		if m.DiffPreview != nil {
			content = m.DiffPreview.Content
		} else {
			content = string(m.Commit.Message)
		}
		return content
	default:
		panic("unsupported result kind in compute output command")
	}
}

func toTextResult(ctx context.Context, content string, matchPattern MatchPattern, outputPattern, separator, selector string) (string, error) {
	if selector != "" {
		// Don't run the search pattern over the search result content
		// when there's an explicit `select:` value.
		return outputPattern, nil
	}

	return output(ctx, content, matchPattern, outputPattern, separator)
}

func toTextExtraResult(content string, r result.Match) *TextExtra {
	return &TextExtra{
		Text:         Text{Value: content, Kind: "output"},
		RepositoryID: int32(r.RepoName().ID),
		Repository:   string(r.RepoName().Name),
	}
}

func (c *Output) Run(ctx context.Context, _ database.DB, r result.Match) (Result, error) {
	onlyPath := c.TypeValue == "path" // don't read file contents for file matches when we only want type:path
	content := resultContent(r, onlyPath)
	env := NewMetaEnvironment(r, content)
	outputPattern, err := substituteMetaVariables(c.OutputPattern, env)
	if err != nil {
		return nil, err
	}

	result, err := toTextResult(ctx, content, c.SearchPattern, outputPattern, c.Separator, c.Selector)
	if err != nil {
		return nil, err
	}

	switch c.Kind {
	case "output.extra":
		return toTextExtraResult(result, r), nil
	default:
		return &Text{Value: result, Kind: "output"}, nil
	}
}
