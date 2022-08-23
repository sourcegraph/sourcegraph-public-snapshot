package aggregation

// This logic is pulled from the compute package, with slight modifications.
// The intention is to not take a dependency on the compute package itself.

import (
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func chunkContent(c result.ChunkMatch, r result.Range) string {
	// Set range relative to the start of the content.
	rr := r.Sub(c.ContentStart)
	return c.Content[rr.Start.Offset:rr.End.Offset]
}

func toTextResult(content string, matchPattern MatchPattern, outputPattern, separator, selector string) (string, error) {
	if selector != "" {
		// Don't run the search pattern over the search result content
		// when there's an explicit `select:` value.
		return outputPattern, nil
	}

	return output(content, matchPattern, outputPattern, separator)
}

func output(fragment string, matchPattern MatchPattern, replacePattern string, separator string) (string, error) {
	var newContent string
	switch match := matchPattern.(type) {
	case *Regexp:
		newContent = substituteRegexp(fragment, match.Value, replacePattern, separator)
	}

	return newContent, nil
}

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	var b strings.Builder
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		b.Write(match.ExpandString([]byte{}, replacePattern, content, submatches))
		b.WriteString(separator)
	}
	return b.String()
}

type MatchPattern interface {
	pattern()
	String() string
}

func (Regexp) pattern() {}
func (Comby) pattern()  {}

type Regexp struct {
	Value *regexp.Regexp
}

type Comby struct {
	Value string
}

func (p Regexp) String() string {
	return p.Value.String()
}

func (p Comby) String() string {
	return p.Value
}

func toRegexpPattern(value string) (MatchPattern, error) {
	rp, err := regexp.Compile(value)
	if err != nil {
		return nil, errors.Wrap(err, "compute endpoint")
	}
	return &Regexp{Value: rp}, nil
}
