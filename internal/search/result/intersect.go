package result

import (
	"sync"
)

// NewLiveMerger creates a type that will perform a "live" merge on
// any number of result sets. numSources is the number of result sets that we
// are merging. Sources are numbered [0,numSources).
func NewLiveMerger(numSources int) *liveMerger {
	return &liveMerger{
		numSources: numSources,
		matches:    make(map[Key]mergeVal, 100),
	}
}

type liveMerger struct {
	mu         sync.Mutex
	numSources int
	matches    map[Key]mergeVal
}

type mergeVal struct {
	match       Match
	sourceMarks []bool
	sent        bool
}

// AddMatches adds a set of Matches from the given source to the merger.
// For each of these matches, if that match has been seen by every source, we
// consider it "streamable" and it will be returned as an element of the return
// value. AddMatches is safe to call from multiple goroutines.
func (lm *liveMerger) AddMatches(matches Matches, source int) Matches {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	var streamableMatches Matches
	for _, match := range matches {
		streamableMatch := lm.addMatch(match, source)
		if streamableMatch != nil {
			streamableMatches = append(streamableMatches, streamableMatch)
		}
	}
	return streamableMatches
}

func (lm *liveMerger) addMatch(m Match, source int) Match {
	// Check if we've seen the match before
	key := m.Key()
	prev, ok := lm.matches[key]
	if prev.sent {
		// If we already returned this match as "streamable",
		// ignore any additional matches with the same key. This
		// is unlikely to happen, but it depends on there being no
		// source that returns two matches with the same key, which
		// we can't guarantee.
		return nil
	}
	if !ok {
		// If we've not seen the match before, track it and continue
		newVal := mergeVal{
			match:       m,
			sourceMarks: make([]bool, lm.numSources), // all false
		}
		newVal.sourceMarks[source] = true
		lm.matches[key] = newVal
		return nil
	}

	// If we have seen it, merge any mergeable types.
	// The type assertions should never fail because we know the keys match.
	switch v := m.(type) {
	case *FileMatch:
		prev.match.(*FileMatch).AppendMatches(v)
	case *CommitMatch:
		prev.match.(*CommitMatch).AppendMatches(v)
	}

	// Mark the key as seen by this source
	prev.sourceMarks[source] = true

	// Check if the match has been added by all sources
	if all(prev.sourceMarks) {
		// Mark that we've returned this match as "streamable"
		// so we don't return it again.
		prev.sent = true
		lm.matches[key] = prev
		return prev.match
	}

	lm.matches[key] = prev
	return nil
}

// UnsentTracked returns the matches that we have added to the merger with
// AddMatches, but weren't seen by every source so weren't returned. This
// returns the union of the sources minus the intersection of the sources.
// Stated differently, when added to the matches that were already returned
// by AddMatches, you get the union of sources.
func (lm *liveMerger) UnsentTracked() Matches {
	var res Matches
	for _, val := range lm.matches {
		if !val.sent {
			res = append(res, val.match)
		}
	}
	return res
}

func all(b []bool) bool {
	for _, val := range b {
		if !val {
			return false
		}
	}
	return true
}
