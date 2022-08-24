package aggregation

// This logic is pulled from the compute package, with slight modifications.
// The intention is to not take a dependency on the compute package itself.

import (
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
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

func extractPattern(basic *query.Basic) (*query.Pattern, error) {
	if basic.Pattern == nil {
		return nil, errors.New("compute endpoint expects nonempty pattern")
	}
	var err error
	var pattern *query.Pattern
	seen := false
	query.VisitPattern([]query.Node{basic.Pattern}, func(value string, negated bool, annotation query.Annotation) {
		if err != nil {
			return
		}
		if negated {
			err = errors.New("compute endpoint expects a nonnegated pattern")
			return
		}
		if seen {
			err = errors.New("compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)")
			return
		}
		pattern = &query.Pattern{Value: value, Annotation: annotation}
		seen = true
	})
	if err != nil {
		return nil, err
	}
	return pattern, nil
}

func fromRegexpMatches(submatches []int, namedGroups []string, content string, range_ result.Range) map[string]int {
	counts := map[string]int{}

	// iterate over pairs of offsets. Cf. FindAllStringSubmatchIndex
	// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatchIndex.
	for j := 0; j < len(submatches); j += 2 {
		start := submatches[j]
		end := submatches[j+1]
		if start == -1 || end == -1 {
			// The entire regexp matched, but a capture
			// group inside it did not. Ignore this entry.
			continue
		}
		value := content[start:end]

		if j == 0 {
			// The first submatch is the overall match
			// value. Don't add this to the Environment

			continue
		}

		current, _ := counts[value]
		counts[value] = current + 1

	}
	return counts
}

func newRange(startOffset, endOffset int) result.Range {
	return result.Range{
		Start: newLocation(-1, -1, startOffset),
		End:   newLocation(-1, -1, endOffset),
	}
}

func newLocation(line, column, offset int) result.Location {
	return result.Location{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}
