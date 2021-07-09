package result

// Union performs a merge of results, merging line matches when they occur in
// the same file.
func Union(left, right []Match) []Match {
	dedup := NewDeduper()
	// Add results to maps for deduping
	for _, result := range left {
		dedup.Add(result)
	}
	for _, result := range right {
		dedup.Add(result)
	}
	return dedup.Results()
}

// Intersect performs a merge of file match results, merging line matches
// for files contained in both result sets.
func Intersect(left, right []Match) []Match {
	rightFileMatches := make(map[Key]*FileMatch)
	for _, m := range right {
		if fileMatch, ok := m.(*FileMatch); ok {
			rightFileMatches[fileMatch.Key()] = fileMatch
		}
	}

	var merged []Match
	for _, m := range left {
		leftFileMatch, ok := m.(*FileMatch)
		if !ok {
			continue
		}

		rightFileMatch := rightFileMatches[leftFileMatch.Key()]
		if rightFileMatch == nil {
			continue
		}

		leftFileMatch.AppendMatches(rightFileMatch)
		merged = append(merged, m)
	}
	return merged
}
