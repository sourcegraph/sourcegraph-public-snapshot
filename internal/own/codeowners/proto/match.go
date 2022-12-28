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
			fmt.Println(err)
			continue
		}
		if m.match(path) {
			return r.GetOwner()
		}
	}
	return nil
}

const separator = "/"

type globMatcher []patternPart

func compile(pattern string) (globMatcher, error) {
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
		// Trailing `/` is equivalent with trailing `/**`
		m = append(m, anySubPath{})
	}
	return m, nil
}

func (m globMatcher) initialState() *big.Int {
	state := big.NewInt(1)
	for i, p := range m {
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
	fmt.Println(parts)
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	moves := big.NewInt(0)
	fmt.Println(m.debugString(state))
	for _, part := range parts {
		m.move(part, state, moves)
		state, moves = moves, state
		fmt.Println(part)
		fmt.Println(m.debugString(state))
	}
	return state.Bit(len(m)) == 1
}

func (m globMatcher) move(part string, state, moves *big.Int) {
	moves.SetInt64(0)
	for i, p := range m {
		if state.Bit(i) == 0 {
			continue
		}
		// Advance to the i+1-th state depending on whether
		// the i-th pattern matches
		bit := uint(0)
		if p.Match(part) {
			bit = uint(1)
		}
		moves.SetBit(moves, i+1, bit)
		// If the i-th pattern part is **, then the current path part
		// may exhaust ** (and therefore i+1-th state is set) or it may
		// be matched and consumed by ** (and therefore i-th state must be set).
		if _, ok := p.(anySubPath); ok {
			moves.SetBit(moves, i, 1)
		}
	}
}

func (m globMatcher) debugString(state *big.Int) string {
	var s strings.Builder
	for i, p := range m {
		if state.Bit(i) != 0 {
			s.WriteByte('X')
		} else {
			s.WriteByte('_')
		}
		fmt.Fprint(&s, p.String())
	}
	if state.Bit(len(m)) != 0 {
		s.WriteByte('X')
	} else {
		s.WriteByte('_')
	}
	return s.String()
}

type patternPart interface {
	fmt.Stringer
	Match(string) bool
}

type anySubPath struct{}

func (p anySubPath) String() string      { return "**" }
func (p anySubPath) Match(_ string) bool { return true }

type exactMatch string

func (p exactMatch) String() string         { return string(p) }
func (p exactMatch) Match(part string) bool { return string(p) == part }

type anyMatch struct{}

func (p anyMatch) String() string         { return "*" }
func (p anyMatch) Match(part string) bool { return true }
