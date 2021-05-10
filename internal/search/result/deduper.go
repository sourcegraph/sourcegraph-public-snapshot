package result

import "sort"

type deduper map[Key]Match

func NewDeduper() deduper {
	return make(map[Key]Match)
}

func (d deduper) Add(m Match) {
	prev, seen := d[m.Key()]

	if seen {
		if prevFileMatch, isFileMatch := prev.(*FileMatch); isFileMatch {
			// Merge file match lines
			prevFileMatch.AppendMatches(m.(*FileMatch)) // key matches, so we know it's a file match
		}
		return
	}

	d[m.Key()] = m
}

func (d deduper) Seen(m Match) bool {
	_, ok := d[m.Key()]
	return ok
}

func (d deduper) Results() []Match {
	matches := make([]Match, 0, len(d))
	for _, match := range d {
		matches = append(matches, match)
	}
	sort.Sort(Matches(matches))
	return matches
}
