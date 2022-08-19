package querybuilder

import (
	"strings"

	"github.com/grafana/regexp"
)

// replaceCaptureGroupsWithString will replace the first capturing group in a regexp
// pattern with a replacement literal. This is somewhat an inverse
// operation of capture groups, with the goal being to produce a new regexp that
// can match a specific instance of a captured value. For example, given the
// pattern `(\w+)-(\w+)` and the replacement `cat` this would generate a
// new regexp `(?:cat)-(\w+)` The capture group that is replaced will be converted
// into a non-capturing group containing the literal replacement.
func replaceCaptureGroupsWithString(pattern string, groups []group, replacement string) string {
	if len(groups) < 1 {
		return pattern
	}
	var sb strings.Builder

	// extract the first capturing group by finding the capturing group with the smallest group number
	var firstCapturing *group
	for i := range groups {
		current := groups[i]
		if !current.capturing {
			continue
		}
		if firstCapturing == nil || current.number < firstCapturing.number {
			firstCapturing = &current
		}
	}
	if firstCapturing == nil {
		return pattern
	}

	offset := 0
	sb.WriteString(pattern[offset:firstCapturing.start])
	sb.WriteString("(?:")
	sb.WriteString(regexp.QuoteMeta(replacement))
	sb.WriteString(")")
	offset = firstCapturing.end + 1

	if firstCapturing.end+1 < len(pattern) {
		// this will copy the rest of the pattern if the last group isn't the end of the pattern string
		sb.WriteString(pattern[offset:])
	}
	return sb.String()
}

type group struct {
	start     int
	end       int
	capturing bool
	number    int
}

// findGroups will extract all capturing and non-capturing groups from a
// **valid** regexp string. If the provided string is not a valid regexp this
// function may panic or otherwise return undefined results.
// This will return all groups (including nested), but not necessarily in any interesting order.
func findGroups(pattern string) (groups []group) {
	var opens []group
	inCharClass := false
	groupNumber := 0
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' {
			i += 1
			continue
		}
		if pattern[i] == '[' {
			inCharClass = true
		} else if pattern[i] == ']' {
			inCharClass = false
		}

		if pattern[i] == '(' && !inCharClass {
			g := group{start: i, capturing: true}
			if peek(pattern, i, 1) == '?' {
				g.capturing = false
				g.number = 0
			} else {
				groupNumber += 1
				g.number = groupNumber
			}
			opens = append(opens, g)

		} else if pattern[i] == ')' && !inCharClass {
			if len(opens) == 0 {
				// this shouldn't happen if we are parsing a well formed regexp since it
				// effectively means we have encountered a closing parenthesis without a
				// corresponding open, but for completeness here this will no-op
				return nil
			}
			current := opens[len(opens)-1]
			current.end = i
			groups = append(groups, current)
			opens = opens[:len(opens)-1]
		}
	}
	return groups
}

func peek(pattern string, currentIndex, peekOffset int) byte {
	if peekOffset+currentIndex >= len(pattern) || peekOffset+currentIndex < 0 {
		return 0
	}
	return pattern[peekOffset+currentIndex]
}
