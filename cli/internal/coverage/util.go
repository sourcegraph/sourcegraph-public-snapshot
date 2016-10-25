package coverage

import (
	"sort"
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/coverage/tokenizer"
)

func AtLeastOne(x int) int {
	if x <= 0 {
		return 1
	}
	return x
}

func AvgInt(values []int) int {
	sum := 0
	for _, i := range values {
		sum += i
	}
	return sum / AtLeastOne(len(values))
}

func AvgDuration(values []time.Duration) time.Duration {
	var sum time.Duration
	for _, i := range values {
		sum += i
	}
	return sum / time.Duration(AtLeastOne(len(values)))
}

func MinDuration(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return time.Duration(0)
	}
	min := values[0]
	for _, i := range values[1:] {
		if i < min {
			min = i
		}
	}
	return min
}

func MaxDuration(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return time.Duration(0)
	}
	max := values[0]
	for _, i := range values[1:] {
		if i > max {
			max = i
		}
	}
	return max
}

func Percentile(k int, values []time.Duration) time.Duration {
	// Copy the slice.
	if len(values) == 0 {
		return 0
	}
	cpy := make([]time.Duration, len(values))
	copy(cpy, values)
	values = cpy

	// Sort the values from smallest to largest.
	sort.Sort(smallestFirst(values))

	index := (float64(k) / 100.0) * float64(len(values))
	return values[int(index)]
}

type smallestFirst []time.Duration

func (s smallestFirst) Len() int      { return len(s) }
func (s smallestFirst) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s smallestFirst) Less(i, j int) bool {
	return s[i] < s[j]
}

// TokRange returns an LSP equivilent range for the given token.
func TokRange(t tokenizer.Token) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      t.Line - 1,
			Character: t.Column - 1,
		},
		End: lsp.Position{
			Line:      t.Line - 1,
			Character: (t.Column - 1) + len(t.Text),
		},
	}
}

// Dist calculates the distance between the two ranges. A distance of 0 implies
// the two ranges are equal, a distance of 1 implies there may be an off-by-one
// error, etc. Aside from "larger distances mean the two are less equal" the
// definition of distance is not strictly defined (except by the source).
func Dist(x, y lsp.Range) int {
	// dist calculates the distance between the two integers.
	dist := func(x, y int) int {
		v := x - y
		if v < 0 {
			return -v
		}
		return v
	}

	// We could return start character, start line, end character, and end line
	// distances all independently for better introspection, but this is good
	// enough until proven otherwise.
	d := dist(x.Start.Character, y.Start.Character)
	d += dist(x.Start.Line, y.Start.Line)
	d += dist(x.End.Character, y.End.Character)
	d += dist(x.End.Line, y.End.Line)
	return d
}
