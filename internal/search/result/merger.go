package result

import (
	"sort"
	"sync"

	"github.com/bits-and-blooms/bitset"
)

// NewMerger creates a type that will perform a merge on any number of result sets.
//
// The merger can be used for either intersections or unions. The matches returned
// from AddMatches are the result of an n-way intersection. The matches returned by
// UnsentTracked are the result of an n-way union minus the matches returned from
// AddMatches. Stated differently, to get the full intersection, collect the matches
// from AddMatches as they are returned, then add the matches returned by UnsentTracked.
//
// numSources is the number of result sets that we are merging. Sources
// are numbered [0,numSources).
func NewMerger(numSources int) *merger {
	return &merger{
		numSources: numSources,
		matches:    make(map[Key]mergeVal, 100),
	}
}

type merger struct {
	mu         sync.Mutex
	numSources int
	matches    map[Key]mergeVal
}

type mergeVal struct {
	match Match
	seen  *bitset.BitSet
	sent  bool
}

type byPopCount []mergeVal

func (s byPopCount) Len() int           { return len(s) }
func (s byPopCount) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byPopCount) Less(i, j int) bool { return s[i].seen.Count() < s[j].seen.Count() }

// AddMatches adds a set of Matches from the given source to the merger.
// For each of these matches, if that match has been seen by every source, we
// consider it "streamable" and it will be returned as an element of the return
// value. AddMatches is safe to call from multiple goroutines.
func (lm *merger) AddMatches(matches Matches, source int) Matches {
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

func (lm *merger) addMatch(m Match, source int) Match {
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
			match: m,
			seen:  bitset.New(uint(lm.numSources)).Set(uint(source)),
		}
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
	case *RepoMatch:
		prev.match.(*RepoMatch).AppendMatches(v)
	}

	// Mark the key as seen by this source
	prev.seen.Set(uint(source))

	// Check if the match has been added by all sources
	if prev.seen.All() {
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
func (lm *merger) UnsentTracked() Matches {
	// We possibly allocate more than is needed. However, assuming that most matches
	// will not been seen by all sources, the limit should be fine.
	matches := make([]mergeVal, 0, len(lm.matches))
	for _, val := range lm.matches {
		if !val.sent {
			matches = append(matches, val)
		}
	}

	// Prioritize matches that were found by multiple sources.
	sort.Sort(sort.Reverse(byPopCount(matches)))

	res := make(Matches, 0, len(matches))
	for _, val := range matches {
		res = append(res, val.match)
	}
	return res
}
