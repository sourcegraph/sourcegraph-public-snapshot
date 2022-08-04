package querybuilder

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"
)

func main() {
	pattern := `name:\((.*)\)(.*) [(] asdf`

	text := `name:(test1) ( asdf`

	reg, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	matches := reg.FindStringSubmatch(text)
	// println(len(matches))
	// for _, match := range matches {
	// 	fmt.Println(match)
	// }

	groups := findGroups(pattern)

	for _, g := range groups {
		if !g.capturing {
			continue
		}
		g.value = matches[g.number]
	}

	fmt.Println(groups)
	// fmt.Println(pairs)
	// fmt.Println(fmt.Sprintf("old_pattern: %s", pattern))
	// fmt.Println(fmt.Sprintf("document: %s", text))
	// fmt.Println(fmt.Sprintf("new_pattern: %s", replaceRange(pattern, matches[1], pairs)))
}

// func replaceCaptureGroups(pattern, replace string) (string, error) {
//
// }

type pair struct {
	x, y int
}

func replaceRange(pattern string, groups []group) string {
	if len(groups) < 1 {
		return pattern
	}
	var sb strings.Builder

	offset := 0
	for _, group := range groups {
		sb.WriteString(pattern[offset:group.start])
		sb.WriteString(group.value)
		offset = group.end
	}
	if offset < len(pattern) {
		sb.WriteString(pattern[offset:])
	}
	return sb.String()
}

type group struct {
	start     int
	end       int
	capturing bool
	number    int
	value     string
}

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
