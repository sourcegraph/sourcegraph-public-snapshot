package proto

import (
	"fmt"
	"math/big"
	"strings"
)

func (f *File) Match(path string) []*Owner {
	for _, r := range f.GetRule() {
		if compile(r.GetPattern()).match(path) {
			return r.GetOwner()
		}
	}
	return nil
}

const separator = "/"

type globMatcher struct {
	parts      []patternPart
	skipStates *big.Int
}

func compile(pattern string) globMatcher {
	m := globMatcher{
		// TODO this assumes not / in front
		parts:      []patternPart{anySubPath{}},
		skipStates: big.NewInt(0),
	}
	skipVector := big.NewInt(0)
	for i, part := range strings.Split(pattern, separator) {
		switch {
		case part == "**":
			m.parts = append(m.parts, anySubPath{})
			skipVector.SetBit(skipVector, i, 1)
		default:
			m.parts = append(m.parts, exactMatch(part))
		}
	}
	return m
}

func (m globMatcher) initialState() *big.Int {
	state := big.NewInt(1)
	for i, p := range m.parts {
		if _, ok := p.(anySubPath); !ok {
			break
		}
		state.SetBit(state, i+1, 1)
	}
	return state
}

func (m globMatcher) match(filePath string) bool {
	state := m.initialState()
	parts := strings.Split(filePath, separator)
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	format := fmt.Sprintf("STATE %%0%db\n", len(m.parts)+1)
	fmt.Printf(format, state)
	moves := big.NewInt(0)
	for _, part := range parts {
		couldSkip := big.NewInt(0).And(state, m.skipStates) // which states could stay while parsing part
		matchMoves(moves, m.parts, part)                    // which states can advance +1
		moved := big.NewInt(0).And(state, moves)            // intersect with current states
		state = moved.Lsh(moved, 1)                         // move everything +1
		state.Or(state, couldSkip)
		fmt.Printf(format, state)
	}
	return state.Bit(len(m.parts)) == 1
}

func matchMoves(moves *big.Int, patterns []patternPart, part string) {
	for i, p := range patterns {
		bit := uint(0)
		if p.Match(part) {
			bit = uint(1)
		}
		moves.SetBit(moves, i, bit)
	}
}

func advance(state *big.Int, moves *big.Int) *big.Int {
	moved := big.NewInt(0).And(state, moves)
	return moved.Lsh(moved, 1)
}

type patternPart interface {
	Match(string) bool
}

type anySubPath struct{}

func (p anySubPath) Match(_ string) bool { return true }

type exactMatch string

func (p exactMatch) Match(part string) bool { return string(p) == part }

// func match(pattern, path string) bool {
// 	// left anchored
// 	if !strings.ContainsAny(pattern, `*?\`) && pattern[0] == os.PathSeparator {
// 		prefix := pattern

// 		// Strip the leading slash as we're anchored to the root already
// 		if prefix[0] == os.PathSeparator {
// 			prefix = prefix[1:]
// 		}

// 		// If the pattern ends with a slash we can do a simple prefix match
// 		if prefix[len(prefix)-1] == os.PathSeparator {
// 			return strings.HasPrefix(path, prefix)
// 		}

// 		// If the strings are the same length, check for an exact match
// 		if len(path) == len(prefix) {
// 			return path == prefix
// 		}

// 		// Otherwise check if the test path is a subdirectory of the pattern
// 		if len(path) > len(prefix) && path[len(prefix)] == os.PathSeparator {
// 			return path[:len(prefix)] == prefix
// 		}
// 		return false
// 	}
// 	re, err := regex(pattern)
// 	if err != nil {
// 		return false
// 	}
// 	return re.MatchString(path)
// }

// func regex(pattern string) (*regexp.Regexp, error) {
// 	// Handle specific edge cases first
// 	switch {
// 	case strings.Contains(pattern, "***"):
// 		return nil, errors.Errorf("pattern cannot contain three consecutive asterisks")
// 	case pattern == "":
// 		return nil, errors.Errorf("empty pattern")
// 	case pattern == "/":
// 		// "/" doesn't match anything
// 		return regexp.Compile(`\A\z`)
// 	}

// 	segs := strings.Split(pattern, "/")

// 	if segs[0] == "" {
// 		// Leading slash: match is relative to root
// 		segs = segs[1:]
// 	} else {
// 		// No leading slash - check for a single segment pattern, which matches
// 		// relative to any descendent path (equivalent to a leading **/)
// 		if len(segs) == 1 || (len(segs) == 2 && segs[1] == "") {
// 			if segs[0] != "**" {
// 				segs = append([]string{"**"}, segs...)
// 			}
// 		}
// 	}

// 	if len(segs) > 1 && segs[len(segs)-1] == "" {
// 		// Trailing slash is equivalent to "/**"
// 		segs[len(segs)-1] = "**"
// 	}

// 	sep := string(os.PathSeparator)

// 	lastSegIndex := len(segs) - 1
// 	needSlash := false
// 	var re strings.Builder
// 	re.WriteString(`\A`)
// 	for i, seg := range segs {
// 		switch seg {
// 		case "**":
// 			switch {
// 			case i == 0 && i == lastSegIndex:
// 				// If the pattern is just "**" we match everything
// 				re.WriteString(`.+`)
// 			case i == 0:
// 				// If the pattern starts with "**" we match any leading path segment
// 				re.WriteString(`(?:.+` + sep + `)?`)
// 				needSlash = false
// 			case i == lastSegIndex:
// 				// If the pattern ends with "**" we match any trailing path segment
// 				re.WriteString(sep + `.*`)
// 			default:
// 				// If the pattern contains "**" we match zero or more path segments
// 				re.WriteString(`(?:` + sep + `.+)?`)
// 				needSlash = true
// 			}

// 		case "*":
// 			if needSlash {
// 				re.WriteString(sep)
// 			}

// 			// Regular wildcard - match any characters except the separator
// 			re.WriteString(`[^` + sep + `]+`)
// 			needSlash = true

// 		default:
// 			if needSlash {
// 				re.WriteString(sep)
// 			}

// 			escape := false
// 			for _, ch := range seg {
// 				if escape {
// 					escape = false
// 					re.WriteString(regexp.QuoteMeta(string(ch)))
// 					continue
// 				}

// 				// Other pathspec implementations handle character classes here (e.g.
// 				// [AaBb]), but CODEOWNERS doesn't support that so we don't need to
// 				switch ch {
// 				case '\\':
// 					escape = true
// 				case '*':
// 					// Multi-character wildcard
// 					re.WriteString(`[^` + sep + `]*`)
// 				case '?':
// 					// Single-character wildcard
// 					re.WriteString(`[^` + sep + `]`)
// 				default:
// 					// Regular character
// 					re.WriteString(regexp.QuoteMeta(string(ch)))
// 				}
// 			}

// 			if i == lastSegIndex {
// 				// As there's no trailing slash (that'd hit the '**' case), we
// 				// need to match descendent paths
// 				re.WriteString(`(?:` + sep + `.*)?`)
// 			}

// 			needSlash = true
// 		}
// 	}
// 	re.WriteString(`\z`)
// 	return regexp.Compile(re.String())
// }
