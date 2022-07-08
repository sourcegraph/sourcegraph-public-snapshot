package result

// Deduper deduplicates matches added to it with Add(). Matches are deduplicated by their key,
// and the return value of Results() is ordered in the same order results are added with Add().
type Deduper struct {
	results Matches
	seen    map[Key]Match
}

func NewDeduper() Deduper {
	return Deduper{
		seen: make(map[Key]Match),
	}
}

func (d *Deduper) Add(m Match) {
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

func (d *Deduper) Seen(m Match) bool {
	_, ok := d.seen[m.Key()]
	return ok
}

func (d *Deduper) Results() Matches {
	return d.results
}
