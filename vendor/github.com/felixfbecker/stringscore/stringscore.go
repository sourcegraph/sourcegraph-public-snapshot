// Based on https://github.com/Microsoft/vscode/blob/321cec5618ce19067ebeb187a782329f028957aa/src/vs/base/common/scorer.ts

package stringscore

import (
	"strings"
	"unicode"
)

// Score computes a score for the given string and the given query.
//
// Rules:
// Character score: 1
// Same case bonus: 1
// Upper case bonus: 1
// End of string bonus: 1
// Consecutive match bonus: 5
// Start of word/path bonus: 7
// Start of string bonus: 8
func Score(target string, query string) int {
	if target == "" || query == "" {
		return 0 // return early if target or query are undefined
	}

	if len(query) > len(target) {
		return 0 // impossible for query to be a substring
	}

	var (
		err         error
		targetC     rune
		targetCPrev rune
		reader      = strings.NewReader(target)
		score       = 0
		firstMatch  = true
	)

	for _, queryC := range query {
		queryCLower := unicode.ToLower(queryC)
		consecutive := true
		for {
			targetCPrev = targetC
			targetC, _, err = reader.ReadRune()
			if err != nil {
				// EOF, so query is not contained in target
				return 0
			}
			if unicode.ToLower(targetC) == queryCLower {
				break
			}
			consecutive = false
		}

		// Character match bonus
		score++

		// Consecutive match bonus
		if consecutive {
			score += 5
		}

		// Same case bonus
		if targetC == queryC {
			score++
		}

		// Start of string bonus
		if firstMatch && consecutive {
			score += 8
		} else if isWordSeparator(targetCPrev) {
			// After separator bonus
			score += 7
		} else if unicode.IsUpper(targetC) {
			// Inside word upper case bonus
			score++
		}
		firstMatch = false
	}

	// End of string bonus
	_, _, err = reader.ReadRune()
	if err != nil {
		score++
	}

	return score
}

const wordPathBoundary = "-_ /\\."

func isWordSeparator(r rune) bool {
	return strings.IndexRune(wordPathBoundary, r) >= 0
}
