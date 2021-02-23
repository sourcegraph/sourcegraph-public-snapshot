package streaming

import (
	"sort"
	"strings"
)

type Filter struct {
	Value string

	// Label is the string to be displayed in the UI.
	Label string

	// Count is the number of matches in a particular repository. Only used
	// for `repo:` filters.
	Count int

	// IsLimitHit is true if the results returned for a repository are
	// incomplete.
	IsLimitHit bool

	// Kind of filter. Should be "repo", "file", or "lang".
	Kind string

	// Important is used to prioritize the order that filters appear in.
	Important bool
}

// Filters is a map of filter values to the Filter.
type Filters map[string]*Filter

// Add the count to the filter with value.
func (m Filters) Add(value string, label string, count int32, limitHit bool, kind string) {
	sf, ok := m[value]
	if !ok {
		sf = &Filter{
			Value:      value,
			Label:      label,
			Count:      int(count),
			IsLimitHit: limitHit,
			Kind:       kind,
		}
		m[value] = sf
	} else {
		sf.Count += int(count)
	}
}

// MarkImportant sets the filter with value as important. Can only be called
// after Add.
func (m Filters) MarkImportant(value string) {
	m[value].Important = true
}

// Compute returns an ordered slice of Filter to present to the user.
func (m Filters) Compute() []*Filter {
	all := make([]*Filter, 0, len(m))
	repos := make([]*Filter, 0, len(m)/2) // heuristic - half of all filters are repo filters.
	for _, f := range m {
		if f.Kind == "repo" {
			repos = append(repos, f)
		} else {
			all = append(all, f)
		}
	}
	sort.Sort(filterSlice(all))
	sort.Sort(filterSlice(repos))

	// limit amount of filters to be rendered arbitrarily to 24, half each.
	if len(all) > 12 {
		all = all[:12]
	}
	if len(repos) > 12 {
		repos = repos[:12]
	}

	all = append(all, repos...)
	sort.Sort(filterSlice(all))

	return all
}

type filterSlice []*Filter

func (fs filterSlice) Len() int {
	return len(fs)
}

func (fs filterSlice) Less(i, j int) bool {
	left := fs[i]
	right := fs[j]
	if left.Important != right.Important {
		// Prefer more important
		return left.Important
	}
	if left.Count != right.Count {
		// Prefer higher count
		return left.Count > right.Count
	}
	// Order alphabetically for equal scores.
	return strings.Compare(left.Value, right.Value) < 0
}

func (fs filterSlice) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}
