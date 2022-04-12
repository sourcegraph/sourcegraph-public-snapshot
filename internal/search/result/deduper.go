package result

// deduper deduplicates matches added to it with Add(). Matches are deduplicated by their key,
// and the return value of Results() is ordered in the same order results are added with Add().
type deduper struct {
	results Matches
	seen    map[Key]Match
}

func NewDeduper() deduper {
	return deduper{
		seen: make(map[Key]Match),
	}
}

func (d *deduper) Add(m Match) {
	prev, seen := d.seen[m.Key()]

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

	d.results = append(d.results, m)
	d.seen[m.Key()] = m
}

func (d *deduper) Seen(m Match) bool {
	_, ok := d.seen[m.Key()]
	return ok
}

func (d *deduper) Results() Matches {
	return d.results
}
