package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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

func output(ctx context.Context, logger log.Logger, fragment string, matchPattern MatchPattern, replacePattern string, separator string) (string, error) {
	var newContent string
	var err error
	switch match := matchPattern.(type) {
	case *Regexp:
		newContent = substituteRegexp(fragment, match.Value, replacePattern, separator)
	case *Comby:
		newContent, err = comby.Outputs(ctx, logger, comby.Args{
			Input:           comby.FileContent(fragment),
			MatchTemplate:   match.Value,
			RewriteTemplate: replacePattern,
			Matcher:         ".generic", // TODO(search): use language or file filter
			ResultKind:      comby.NewlineSeparatedOutput,
			NumWorkers:      0,
		})
		if err != nil {
			return "", err
		}

	}
	return newContent, nil
}

func resultChunks(r result.Match, kind string, onlyPath bool) []string {
	switch m := r.(type) {
	case *result.RepoMatch:
		return []string{string(m.Name)}
	case *result.FileMatch:
		if onlyPath {
			return []string{m.Path}
		}

		chunks := make([]string, 0, len(m.ChunkMatches))
		for _, cm := range m.ChunkMatches {
			for _, range_ := range cm.Ranges {
				chunks = append(chunks, chunkContent(cm, range_))
			}
		}

		if kind == "output.structural" {
			// concatenate all chunk matches into one string so we
			// don't invoke comby for every result.
			return []string{strings.Join(chunks, "")}
		}

		return chunks
	case *result.CommitDiffMatch:
		var sb strings.Builder
		for _, h := range m.Hunks {
			for _, l := range h.Lines {
				sb.WriteString(l)
			}
		}
		return []string{sb.String()}
	case *result.CommitMatch:
		var content string
		if m.DiffPreview != nil {
			content = m.DiffPreview.Content
		} else {
			content = string(m.Commit.Message)
		}
		return []string{content}
	case *result.OwnerMatch:
		return []string{m.ResolvedOwner.Identifier()}
	default:
		panic("unsupported result kind in compute output command")
	}
}

func toTextResult(ctx context.Context, logger log.Logger, content string, matchPattern MatchPattern, outputPattern, separator, selector string) (string, error) {
	if selector != "" {
		// Don't run the search pattern over the search result content
		// when there's an explicit `select:` value.
		return outputPattern, nil
	}
	return output(ctx, logger, content, matchPattern, outputPattern, separator)
}

func toTextExtraResult(content string, r result.Match) *TextExtra {
	return &TextExtra{
		Text:         Text{Value: content, Kind: "output"},
		RepositoryID: int32(r.RepoName().ID),
		Repository:   string(r.RepoName().Name),
	}
}

func (c *Output) Run(ctx context.Context, _ gitserver.Client, r result.Match) (Result, error) {
	onlyPath := c.TypeValue == "path" // don't read file contents for file matches when we only want type:path
	chunks := resultChunks(r, c.Kind, onlyPath)

	var sb strings.Builder
	for _, content := range chunks {
		env := NewMetaEnvironment(r, content)
		outputPattern, err := substituteMetaVariables(c.OutputPattern, env)
		if err != nil {
			return nil, err
		}

		textResult, err := toTextResult(ctx, log.Scoped("compute"), content, c.SearchPattern, outputPattern, c.Separator, c.Selector)
		if err != nil {
			return nil, err
		}
		sb.WriteString(textResult)
	}

	switch c.Kind {
	case "output.extra":
		return toTextExtraResult(sb.String(), r), nil
	default:
		return &Text{Value: sb.String(), Kind: "output"}, nil
	}
}
