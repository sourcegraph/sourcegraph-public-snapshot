package result

import "sort"

type deduper map[Key]Match

func NewDeduper() deduper {
	return make(map[Key]Match)
}

func (d deduper) Add(m Match) {
	prev, seen := d[m.Key()]

	if seen {
		switch prevMatch := prev.(type) {
		// key matches, so we know to convert to respective type
		case *FileMatch:
			prevMatch.AppendMatches(m.(*FileMatch))
		case *CommitMatch:
			prevMatch.AppendMatches(m.(*CommitMatch))
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
