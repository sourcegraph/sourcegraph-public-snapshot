package proto

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func (f *File) Match(path string) []*Owner {
	for _, r := range f.GetRule() {
		m, err := compile(r.GetPattern())
		if err != nil {
			continue
		}
		if m.match(path) {
			return r.GetOwner()
		}
	}
	return nil
}

const separator = "/"

type globPattern []patternPart

func compile(pattern string) (globPattern, error) {
	var m []patternPart
	if !strings.HasPrefix(pattern, separator) {
		// No leading `/` is equivalent to prefixing with `/**/`.
		m = append(m, anySubPath{})
	}
	for _, part := range strings.Split(strings.Trim(pattern, separator), separator) {
		switch {
		case part == "":
			return nil, errors.New("two consecutive forward slashes")
		case part == "*":
			m = append(m, anyMatch{})
		case part == "**":
			m = append(m, anySubPath{})
		default:
			m = append(m, exactMatch(part))
		}
	}
	if strings.HasSuffix(pattern, separator) {
		// Trailing `/` is equivalent with trailing `/**/*`
		m = append(m, anySubPath{}, anyMatch{})
	}
	return m, nil
}

// match iterates over `filePath` separated by `/`. It keeps track of which prefixes
// of the `globPattern` would be matching prefix of `filePath` up to current `part`.
// It keeps the state in `*big.Int` which implements a bit vector. Example:
// / ** / src / java / test / ** / *Test.java   | Pattern
// 0    1     2      3      4    5            6 | Matching state bit index of prefix:
// *    *     *                                 | /src
// *    *            *                          | /src/java
// *    *                   *    *              | /src/java/test
// *    *                   *    *            * | /src/java/test/UnitTest.java
// The match is successful if after iterating throght the whole path,
// full pattern matches.
func (p globPattern) match(filePath string) bool {
	currentState := big.NewInt(0)
	p.markEmptyMatches(currentState)
	parts := strings.Split(strings.Trim(filePath, separator), separator)
	nextState := big.NewInt(0)
	for _, part := range parts {
		nextState.SetInt64(0)
		p.consume(part, currentState, nextState)
		currentState, nextState = nextState, currentState
	}
	return p.matchesWhole(currentState)
}

// markEmptyMathes initializes a matching state with positions that are
// matches for an empty input. This is most often just bit 0, but in case
// there are sub-path wild-card **, it is expanded to all indices past the
// wild-cards, since they match empty path.
func (p globPattern) markEmptyMatches(state *big.Int) {
	state.SetBit(state, 0, 1)
	for i, p := range p {
		if _, ok := p.(anySubPath); !ok {
			break
		}
		state.SetBit(state, i+1, 1)
	}
}

// matchesWhole returns true if given state for matching this `globPattern`
// matches the whole pattern as opposed to just a prefix
func (p globPattern) matchesWhole(state *big.Int) bool {
	return state.Bit(len(p)) == 1
}

// consume takes the next `part` of the tested path, and the `current` state
// of which prefixes of the `globPattern` are matched and advances matching
// by the step of consuming given part. That is, all currently matching
// prefixes are considered, and for each the following pattern part is tested
// to match the given path part. The result is written to `next` which is assumed
// to be zeroed.
func (p globPattern) consume(part string, current, next *big.Int) {
	// ** invariant is that the position after ** is set if the position before ** is set.
	for i := 0; i < len(p); i++ {
		if current.Bit(i) == 0 {
			continue
		}
		// Advance to the i+1-th state depending on whether
		// the i-th pattern matches
		bit := uint(0)
		if p[i].Match(part) {
			bit = uint(1)
		}
		next.SetBit(next, i+1, bit)
		// Set the bit after next **
		if i+1 < len(p) {
			if _, ok := p[i+1].(anySubPath); ok {
				next.SetBit(next, i+2, bit)
			}
		}
		// Leave the bit set before **
		if _, ok := p[i].(anySubPath); ok {
			next.SetBit(next, i, 1)
		}
	}
}

func (p globPattern) debugString(state *big.Int) string {
	var s strings.Builder
	for i, p := range p {
		if state.Bit(i) != 0 {
			s.WriteByte('X')
		} else {
			s.WriteByte('_')
		}
		fmt.Fprint(&s, p.String())
	}
	if state.Bit(len(p)) != 0 {
		s.WriteByte('X')
	} else {
		s.WriteByte('_')
	}
	return s.String()
}

// patternPart implements matching for a single chunk of a glob pattern
// when separated by `/`.
type patternPart interface {
	String() string
	// Match is true if given file or directory name on the path matches
	// this part of the glob pattern.
	Match(string) bool
}

// anySubPath is indicated by ** in glob patterns, and matches arbitrary
// number of parts.
type anySubPath struct{}

func (p anySubPath) String() string      { return "**" }
func (p anySubPath) Match(_ string) bool { return true }

// exactMatch is indicated by an exact name of directory or a file within
// the glob pattern, and matches that exact part of the path only.
type exactMatch string

func (p exactMatch) String() string         { return string(p) }
func (p exactMatch) Match(part string) bool { return string(p) == part }

// anyMatch is indicated by * in a glob pattern, and matches any single file
// or directory on the path.
type anyMatch struct{}

func (p anyMatch) String() string         { return "*" }
func (p anyMatch) Match(part string) bool { return true }
