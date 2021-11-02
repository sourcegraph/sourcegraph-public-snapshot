package compute

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
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
	var newFragment string
	switch match := matchPattern.(type) {
	case *Regexp:
		newFragment = substituteRegexp(fragment, match.Value, replacePattern, separator)
	case *Comby:
		return nil, nil
	}
	return &Text{Value: newFragment, Kind: "output"}, nil
}

func (c *Output) Run(ctx context.Context, fm *result.FileMatch) (Result, error) {
	lines := make([]string, 0, len(fm.LineMatches))
	for _, line := range fm.LineMatches {
		lines = append(lines, line.Preview)
	}
	fragment := strings.Join(lines, "\n")
	return output(ctx, fragment, c.MatchPattern, c.OutputPattern, c.Separator)
}
