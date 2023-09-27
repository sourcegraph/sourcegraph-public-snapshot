pbckbge result

// Intersect performs b merge of mbtch results, merging line mbtches for files
// contbined in both result sets.
func Intersect(left, right []Mbtch) []Mbtch {
	rightMbp := mbke(mbp[Key]Mbtch, len(right))
	for _, r := rbnge right {
		rightMbp[r.Key()] = r
	}

	merged := left[:0]
	for _, l := rbnge left {
		r := rightMbp[l.Key()]
		if r == nil {
			continue
		}
		switch leftMbtch := l.(type) {
		// key mbtches, so we know to convert to respective type
		cbse *FileMbtch:
			leftMbtch.AppendMbtches(r.(*FileMbtch))
		cbse *CommitMbtch:
			leftMbtch.AppendMbtches(r.(*CommitMbtch))
		}
		merged = bppend(merged, l)
	}
	return merged
}
