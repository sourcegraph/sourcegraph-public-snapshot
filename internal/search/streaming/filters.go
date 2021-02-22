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
	filterSlice := make([]*Filter, 0, len(m))
	repoFilterSlice := make([]*Filter, 0, len(m)/2) // heuristic - half of all filters are repo filters.
	for _, f := range m {
		if f.Kind == "repo" {
			repoFilterSlice = append(repoFilterSlice, f)
		} else {
			filterSlice = append(filterSlice, f)
		}
	}
	sort.Slice(filterSlice, func(i, j int) bool {
		if filterSlice[i].Important == filterSlice[j].Important {
			return filterSlice[i].Count > filterSlice[j].Count
		}
		return filterSlice[i].Important
	})
	// limit amount of non-repo filters to be rendered arbitrarily to 12
	if len(filterSlice) > 12 {
		filterSlice = filterSlice[:12]
	}

	allFilters := append(filterSlice, repoFilterSlice...)
	sort.Slice(allFilters, func(i, j int) bool {
		left := allFilters[i]
		right := allFilters[j]
		if left.Important == right.Important {
			// Order alphabetically for equal scores.
			return strings.Compare(left.Value, right.Value) < 0
		}
		return left.Important
	})

	return allFilters
}
