package search

import "github.com/sourcegraph/sourcegraph/internal/search/result"

// displayFilter mutates event matches before sending them to the end user
// (after all statistics/etc have been calculated).
//
// In particular displayFilter is responsible for ensuring we only send
// display limit results (?display URL parameter).
//
// Note: this is different to the search limit (MaxResults). Search limit
// decides when to stop searching, this decides when to stop sending the
// actual matches to the user.
type displayFilter struct {
	// MatchLimit is the computed display limit (?display URL parameter).
	// After MatchLimit results have been sent to the user we stop streaming
	// more matches.
	MatchLimit int

	matchesRemaining int
}

// newDisplayFilter returns the displayFilter for args. maxResults is the
// maximum results we are allowed to search. This is used to bound the display
// limit.
func newDisplayFilter(args *args, maxResults int) *displayFilter {
	// Display is the number of results we send down. If display is < 0 we
	// want to send everything we find before hitting a limit. Otherwise we
	// can only send up to limit results.
	displayLimit := args.Display
	if displayLimit < 0 || displayLimit > maxResults {
		displayLimit = maxResults
	}

	return &displayFilter{
		MatchLimit:       displayLimit,
		matchesRemaining: displayLimit,
	}
}

// Limit will limit m to enforce the display limits.
//
// Note: this should only be called after aggregating statistics.
func (d *displayFilter) Limit(m *result.Matches) {
	d.matchesRemaining = m.Limit(d.matchesRemaining)
}
