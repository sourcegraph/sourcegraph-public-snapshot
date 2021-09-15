package search

import (
	"encoding/gob"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CommitPredicate is an interface representing the queries we can run against a commit.
type CommitPredicate interface {
	// Match returns whether the given predicate matches a commit and, if it does,
	// the portions of the commit that match in the form of *CommitHighlights
	Match(*LazyCommit) (matched bool, highlights *HighlightedCommit, err error)
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	Regexp
}

func (a *AuthorMatches) Match(cv *LazyCommit) (bool, *HighlightedCommit, error) {
	return a.Regexp.Match(cv.AuthorName) || a.Regexp.Match(cv.AuthorEmail), nil, nil
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	Regexp
}

func (c *CommitterMatches) Match(cv *LazyCommit) (bool, *HighlightedCommit, error) {
	return c.Regexp.Match(cv.CommitterName) || c.Regexp.Match(cv.CommitterEmail), nil, nil
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	time.Time
}

func (c *CommitBefore) Match(lc *LazyCommit) (bool, *HighlightedCommit, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, nil, err
	}
	return authorDate.Before(c.Time), nil, nil
}

// CommitAfter is a predicate that matches if the commit is after the given date
type CommitAfter struct {
	time.Time
}

func (c *CommitAfter) Match(lc *LazyCommit) (bool, *HighlightedCommit, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, nil, err
	}
	return authorDate.After(c.Time), nil, nil
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	Regexp
}

func (m *MessageMatches) Match(commit *LazyCommit) (bool, *HighlightedCommit, error) {
	results := m.FindAllIndex(commit.Message, -1) // TODO limit?
	if results == nil {
		return false, nil, nil
	}

	messageString := string(commit.Message)
	return true, &HighlightedCommit{
		Message: HighlightedString{
			Content:    messageString,
			Highlights: matchesToRanges(messageString, results),
		},
	}, nil
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	Regexp
}

func (dm *DiffMatches) Match(commit *LazyCommit) (bool, *HighlightedCommit, error) {
	diff, err := commit.Diff()
	if err != nil {
		return false, nil, err
	}

	foundMatch := false
	var highlights Ranges

	diff.ForEachDelta(func(d Delta) bool {
		d.ForEachHunk(func(h Hunk) bool {
			h.ForEachLine(func(l Line) bool {
				switch l.Origin() {
				case '+', '-':
				default:
					return true
				}

				content, loc := l.Content()
				matches := dm.FindAllStringIndex(content, -1)
				if matches != nil {
					foundMatch = true
					highlights = append(highlights, matchesToRanges(content, matches).Shift(loc)...)
				}
				return true
			})
			return true
		})
		return true
	})

	return foundMatch, &HighlightedCommit{Diff: HighlightedString{Content: string(diff), Highlights: highlights}}, nil
}

// // DiffModifiesFile is a predicate that matches if the commit modifies any files
// // that match the given regex pattern.
// type DiffModifiesFile struct {
// 	Regexp
// }

// func (dmf *DiffModifiesFile) Match(commit *LazyCommit) (bool, *HighlightedCommit) {
// 	diff, err := commit.Diff()
// 	if err != nil {
// 		// TODO is ignoring okay, or should the Match() signature return an error?
// 		return false, nil
// 	}

// 	foundMatch := false
// 	var highlights Ranges

// 	diff.ForEachDelta(func(d Delta) bool {
// 		oldFile, oldLoc := d.OldFile()
// 		oldFileMatches := dmf.FindAllStringIndex(oldFile, -1)
// 		if oldFileMatches != nil {
// 			foundMatch = true
// 			highlights = append(highlights, matchesToRanges(oldFile, oldFileMatches).Shift(oldLoc)...)
// 		}

// 		newFile, newLoc := d.NewFile()
// 		newFileMatches := dmf.FindAllStringIndex(newFile, -1)
// 		if newFileMatches != nil {
// 			foundMatch = true
// 			highlights = append(highlights, matchesToRanges(newFile, newFileMatches).Shift(newLoc)...)
// 		}

// 		return true
// 	})

// 	return foundMatch, &HighlightedCommit{Diff: HighlightedString{Content: string(diff), Highlights: highlights}}
// }

// And is a predicate that matches if all of its children predicates match
type And struct {
	Children []CommitPredicate
}

func (a *And) Match(commit *LazyCommit) (bool, *HighlightedCommit, error) {
	highlights := &HighlightedCommit{}
	for _, child := range a.Children {
		childMatched, childHighlights, err := child.Match(commit)
		if err != nil {
			return false, nil, err
		}

		if !childMatched {
			// Since we don't care about the highlights if we don't match all children, we can short-circuit
			return false, nil, nil
		}
		highlights.Merge(childHighlights)
	}
	return true, highlights, nil
}

// Or is a predicate that matches if any of its children predicates match
type Or struct {
	Children []CommitPredicate
}

func (o *Or) Match(commit *LazyCommit) (bool, *HighlightedCommit, error) {
	hasMatch := false
	mergedHighlights := &HighlightedCommit{}
	for _, child := range o.Children {
		matched, highlights, err := child.Match(commit)
		if err != nil {
			return false, nil, err
		}
		if matched {
			// Because we want to highlight every match, we can't short circuit
			hasMatch = true
			mergedHighlights.Merge(highlights)
		}
	}
	return hasMatch, mergedHighlights, nil
}

// Not is a predicate that matches if its child predicate does not match
type Not struct {
	Child CommitPredicate
}

func (n *Not) Match(commit *LazyCommit) (bool, *HighlightedCommit, error) {
	// Even if the child highlights, since we're negating, the match shouldn't be highlighted
	foundMatch, _, err := n.Child.Match(commit)
	return !foundMatch, nil, err
}

// Regexp is a thin wrapper around the stdlib Regexp type that enables gob encoding
type Regexp struct {
	*regexp.Regexp
}

func (r Regexp) GobEncode() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Regexp) GobDecode(data []byte) (err error) {
	r.Regexp, err = regexp.Compile(string(data))
	return err
}

// matchesToRanges is a helper that takes the return value of regexp.FindAllStringIndex()
// and converts it to Ranges.
// INVARIANT: matches must be ordered and non-overlapping,
// which is guaranteed by regexp.FindAllStringIndex()
func matchesToRanges(s string, matches [][]int) Ranges {
	// Incrementally search newlines to avoid counting newlines over the
	// entire string for every match.
	var (
		lastNewlineOffset int
		newlineCount      int
		searchEnd         int
	)
	lineAndColumn := func(offset int) (line, column int) {
		newlines, index := newlineCountAndLastIndex(s[searchEnd:offset])
		newlineCount += newlines
		if index >= 0 {
			lastNewlineOffset = searchEnd + index
		}
		searchEnd = offset

		if newlineCount > 0 {
			return newlineCount, offset - (lastNewlineOffset + 1)
		}
		return 0, offset
	}

	res := make(Ranges, 0, len(matches))
	for _, match := range matches {
		startLine, startColumn := lineAndColumn(match[0])
		endLine, endColumn := lineAndColumn(match[1])
		res = append(res, Range{
			Start: Location{Offset: match[0], Line: startLine, Column: startColumn},
			End:   Location{Offset: match[1], Line: endLine, Column: endColumn},
		})
	}
	return res
}

func newlineCountAndLastIndex(s string) (count int, lastIndex int) {
	lastIndex = strings.LastIndexByte(s, '\n')
	if lastIndex == -1 {
		return 0, -1
	}

	return strings.Count(s[:lastIndex], "\n") + 1, lastIndex
}

var registerOnce sync.Once

func RegisterGob() {
	registerOnce.Do(func() {
		gob.Register(&AuthorMatches{})
		gob.Register(&CommitterMatches{})
		gob.Register(&CommitBefore{})
		gob.Register(&CommitAfter{})
		gob.Register(&MessageMatches{})
		// gob.Register(&DiffMatches{})
		// gob.Register(&DiffModifiesFile{})
		gob.Register(&And{})
		gob.Register(&Or{})
		gob.Register(&Not{})
	})
}
