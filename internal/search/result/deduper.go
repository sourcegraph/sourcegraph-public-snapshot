package result

// deduper deduplicates matches added to it with Add(). Matches are deduplicated by their key,
// and the return value of Results() is ordered in the same order results are added with Add().
type deduper struct {
	results []Match
	seen    map[Key]Match
	count   int
	limit   int
}

func NewDeduper(limit int) deduper {
	return deduper{
		seen:  make(map[Key]Match),
		limit: limit,
	}
}

func (d *deduper) Add(m Match) int {
	if d.limit > 0 {
		if d.count < d.limit {
			m.Limit(d.limit - d.count)
		} else {
			return d.count
		}
	}

	prev, seen := d.seen[m.Key()]

	if seen {
		switch prevMatch := prev.(type) {
		// key matches, so we know to convert to respective type
		case *FileMatch:
			prevMatch.AppendMatches(m.(*FileMatch))
		case *CommitMatch:
			prevMatch.AppendMatches(m.(*CommitMatch))
		}
	} else {
		d.results = append(d.results, m)
		d.seen[m.Key()] = m
	}

	d.count += m.ResultCount()

	return d.count
}

func (d *deduper) Seen(m Match) bool {
	_, ok := d.seen[m.Key()]
	return ok
}

func (d *deduper) Results() []Match {
	return d.results
}

func (d *deduper) ResultCount() int {
	return d.count
}
