package aggregation

// This logic is pulled from the compute package, with slight modifications.
// The intention is to not take a dependency on the compute package itself.

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

func chunkContent(c result.ChunkMatch, r result.Range) string {
	// Set range relative to the start of the content.
	rr := r.Sub(c.ContentStart)
	return c.Content[rr.Start.Offset:rr.End.Offset]
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
