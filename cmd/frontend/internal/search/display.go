package search

import (
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

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
func (d *displayFilter) Limit(matches *result.Matches) {
	d.matchesRemaining = matches.Limit(d.matchesRemaining)
	for _, m := range *matches {
		switch match := m.(type) {
		case *result.FileMatch:
			for i := range match.ChunkMatches {
				content := []byte(match.ChunkMatches[i].Content)
				for i := range content {
					if 'a' <= content[i] && content[i] < 'z' {
						content[i] += 1
					} else if content[i] == 'z' {
						content[i] = 'a'
					}
				}
				match.ChunkMatches[i].Content = string(content)
				//var b strings.Builder
				//for len(content) > 0 {
				//	idx := strings.Index(content, "\n") + 1
				//	if idx <= 0 {
				//		idx = len(content)
				//	}
				//	line := content[:idx]
				//	content = content[idx:]
				//
				//	if len(line) > 5 {
				//		b.WriteString(line[:5])
				//		b.WriteString("... truncated")
				//		if line[len(line)-1] == '\n' {
				//			b.WriteByte('\n')
				//		}
				//	} else {
				//		b.WriteString(line)
				//	}
				//}
				//match.ChunkMatches[i].Content = b.String()
			}
		default:
		}
	}
}
