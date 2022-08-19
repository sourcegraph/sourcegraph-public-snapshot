package querybuilder

import (
	"sort"
	"strings"

	"github.com/grafana/regexp"
)

// replaceCaptureGroupsWithString will replace capturing groups in a regexp
// pattern with their respective literal matches. This is somewhat an inverse
// operation of capture groups, with the goal being to produce a new regexp that
// can match a specific instance of a captured value. For example, given the
// pattern `(\w+)-(\w+)` and the text `cat-cow dog-pig` this would generate a
// new regexp `(?:cat|dog)-(?:cow|pig)` The maxGroup argument allows
// control over how many capture groups are replaced (up to and including capture group N, indexed at 1). If the value is
// negative all capture groups will be replaced. Each capture group that is replaced will be converted
// into a non-capturing group containing the literal matches.
func replaceCaptureGroupsWithString(pattern string, groups []group, matches [][]string, maxGroup int) string {
	if len(groups) < 1 {
		return pattern
	} else if len(matches) == 0 || len(matches[0]) == 0 {
		return pattern
	}
	var sb strings.Builder

	capturing := make([]group, 0, len(groups))
	for _, g := range groups {
		if g.capturing {
			capturing = append(capturing, g)
		}
	}

	// groups need to be in stable order
	sort.Slice(capturing, func(i, j int) bool {
		return capturing[i].start < capturing[j].start
	})
	// todo handle nested groups by generating a set of non-overlapping groups

	// pivot the matches from [match][group_number] to [group_number][match] to more easily reference the set of literals
	pivotMatches := make([][]string, len(matches[0]))
	for i := range pivotMatches {
		pivotMatches[i] = make([]string, 0, len(matches))
	}
	for _, match := range matches {
		for inner, literal := range match {
			pivotMatches[inner] = append(pivotMatches[inner], regexp.QuoteMeta(literal))
		}
	}

	if maxGroup < 0 {
		// even though the length of groups isn't necessarily the number of matched
		// groups we can still use that as the max index here. The iteration will
		// effectively become the minimum of this and the actual length. We
		maxGroup = len(capturing)
	}
	offset := 0
	for groupIndex, group := range capturing {
		if !group.capturing {
			continue
		} else if groupIndex > (maxGroup - 1) {
			// the -1 offset is because we reference regexp groups with submatches starting at 1, whereas the array offset is 0
			break
		}
		sb.WriteString(pattern[offset:group.start])
		sb.WriteString("(?:")
		sb.WriteString(strings.Join(pivotMatches[group.number], "|"))
		sb.WriteString(")")
		offset = group.end + 1
	}
	if offset < len(pattern) {
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
