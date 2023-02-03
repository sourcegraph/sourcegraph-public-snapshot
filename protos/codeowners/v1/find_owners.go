package v1

import (
	"math/big"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FindOwners returns the Owners associated with given path as per this CODEOWNERS file.
// Rules are evaluated in order: Returned owners come from the rule which pattern matches
// given path, that is the furthest down the file.
func (x *File) FindOwners(path string) []*Owner {
	var owners []*Owner
	for _, rule := range x.GetRule() {
		glob, err := compile(rule.GetPattern())
		if err != nil {
			continue
		}
		if glob.match(path) {
			owners = rule.GetOwner()
		}
	}
	return owners
}

const separator = "/"

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

func (p anyMatch) String() string      { return "*" }
func (p anyMatch) Match(_ string) bool { return true }

// asteriskPattern is a pattern that may contain * glob wildcard.
type asteriskPattern struct {
	glob     string
	compiled *regexp.Regexp
}

func makeAsteriskPattern(pattern string) (asteriskPattern, error) {
	quoted := regexp.QuoteMeta(pattern)
	regular := strings.ReplaceAll(quoted, `\*`, `.*`)
	compiled, err := regexp.Compile("^" + regular + "$")
	if err != nil {
		return asteriskPattern{}, err
	}
	return asteriskPattern{glob: pattern, compiled: compiled}, nil
}
func (p asteriskPattern) String() string { return p.glob }
func (p asteriskPattern) Match(part string) bool {
	if p.compiled == nil {
		return false
	}
	return p.compiled.FindString(part) != ""
}

// compile translates a text representation of a glob pattern
// to an executable one that can `match` file paths.
func compile(pattern string) (globPattern, error) {
	var glob globPattern
	// No leading `/` is equivalent to prefixing with `/**/`.
	// The pattern matches arbitrarily down the directory tree.
	if !strings.HasPrefix(pattern, separator) {
		glob = append(glob, anySubPath{})
	}
	for _, part := range strings.Split(strings.Trim(pattern, separator), separator) {
		switch part {
		case "":
			return nil, errors.New("two consecutive forward slashes")
		case "**":
			glob = append(glob, anySubPath{})
		case "*":
			glob = append(glob, anyMatch{})
		default:
			if strings.Contains(part, "*") {
				p, err := makeAsteriskPattern(part)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to interpret %q within %q", part, pattern)
				}
				glob = append(glob, p)
			} else {
				glob = append(glob, exactMatch(part))
			}
		}
	}
	// Trailing `/` is equivalent with ending the pattern with `/**` instead.
	if strings.HasSuffix(pattern, separator) {
		glob = append(glob, anySubPath{})
	}
	// Trailing `/**` (explicitly or implicitly like above) is necessarily
	// translated to `/**/*.
	// This is because, trailing `/**` should not match if the path finishes
	// with the part that matches up to and excluding final `**` wildcard.
	// Example: Neither `/foo/bar/**` nor `/foo/bar/` should match file `/foo/bar`.
	if len(glob) > 0 {
		if _, ok := glob[len(glob)-1].(anySubPath); ok {
			glob = append(glob, anyMatch{})
		}
	}
	return glob, nil
}

// globPattern implements a pattern for matching file paths,
// which can use directory/file names, * and ** wildcards,
// and may or may not be anchored to the root directory.
type globPattern []patternPart

// match iterates over `filePath` separated by `/`. It uses a bit vector
// to track which prefixes of glob pattern match the file path prefix so far.
// Bit vector indices correspond to separators between pattern parts.
//
// Visualized matching of `/src/java/test/UnitTest.java`
// against `src/java/test/**/*Test.java`:
// / ** / src / java / test / ** / *Test.java   | Glob pattern
// 0    1     2      3      4    5            6 | Bit vector index
// X    X     -      -      -    -            - | / (starting state)
// X    X     X      -      -    -            - | /src
// X    X     -      X      -    -            - | /src/java
// X    X     -      -      X    X            - | /src/java/test
// X    X     -      -      X    X            X | /src/java/test/UnitTest.java
//
// Another example of matching `/src/app/components/Label.tsx`
// against `/src/app/components/*.tsx`:
// / src / app / components / *.tsx   | Glob pattern
// 0     1     2            3       4 | Bit vector index
// X     -     -            -       - | / (starting state)
// -     X     -            -       - | /src
// -     -     X            -       - | /src/app
// -     -     -            X       - | /src/app/components
// -     -     -            -       X | /src/app/components/Label.tsx
//
// The match is successful if after iterating through the whole file path,
// full pattern matches, that is, there is a bit at the end of the glob.
func (glob globPattern) match(filePath string) bool {
	currentState := big.NewInt(0)
	glob.markEmptyMatches(currentState)
	filePathParts := strings.Split(strings.Trim(filePath, separator), separator)
	nextState := big.NewInt(0)
	for _, part := range filePathParts {
		nextState.SetInt64(0)
		glob.consume(part, currentState, nextState)
		currentState, nextState = nextState, currentState
	}
	return glob.matchesWhole(currentState)
}

// markEmptyMatches initializes a matching state with positions that are
// matches for an empty input (`/`). This is most often just bit 0, but in case
// there are subpath wildcard **, it is expanded to all indices past the
// wildcards, since they match empty path.
func (glob globPattern) markEmptyMatches(state *big.Int) {
	state.SetBit(state, 0, 1)
	for i, globPart := range glob {
		if _, ok := globPart.(anySubPath); !ok {
			break
		}
		state.SetBit(state, i+1, 1)
	}
}

// matchesWhole returns true if given state indicates whole glob being matched.
func (glob globPattern) matchesWhole(state *big.Int) bool {
	return state.Bit(len(glob)) == 1
}

// consume advances matching algorithm by a single part of a file path.
// The `current` bit vector is the matching state for up until, but excluding
// given `part` of the file path. The result - next set of states - is written
// to bit vector `next`, which is assumed to be zero when passed in.
func (glob globPattern) consume(part string, current, next *big.Int) {
	// Since `**` or `anySubPath` can match any number of times, we hold
	// an invariant: If a bit vector has 1 at the state preceding `**`,
	// then that bit vector also has 1 at the state following `**`.
	for i := 0; i < len(glob); i++ {
		if current.Bit(i) == 0 {
			continue
		}
		// Case 1: `current` matches before i-th part of the pattern,
		// so set the i+1-th position of the `next` state to whether
		// the i-th pattern matches (consumes) `part`.
		bit := uint(0)
		if glob[i].Match(part) {
			bit = uint(1)
		}
		next.SetBit(next, i+1, bit)
		// Keep the invariant: if there is `**` afterwards, set it
		// to the same bit. This will not be overridden in the next
		// loop turns as `**` always matches.
		if i+1 < len(glob) {
			if _, ok := glob[i+1].(anySubPath); ok {
				next.SetBit(next, i+2, bit)
			}
		}
		// Case 2: To allow `**` to consume subsequent parts of the file path,
		// we keep the i-th bit - which precedes `**` - set.
		if _, ok := glob[i].(anySubPath); ok {
			next.SetBit(next, i, 1)
		}
	}
}
