package search

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func ToMatchTree(q protocol.SearchQuery) MatchTree {
	switch v := q.(type) {
	case *protocol.AuthorMatches:
		return &AuthorMatches{*v}
	case *protocol.CommitterMatches:
		return &CommitterMatches{*v}
	case *protocol.CommitBefore:
		return &CommitBefore{*v}
	case *protocol.CommitAfter:
		return &CommitAfter{*v}
	case *protocol.MessageMatches:
		return &MessageMatches{*v}
	case *protocol.DiffMatches:
		return &DiffMatches{*v}
	case *protocol.DiffModifiesFile:
		return &DiffModifiesFile{*v}
	case *protocol.And:
		children := make([]MatchTree, 0, len(v.Children))
		for _, child := range v.Children {
			children = append(children, ToMatchTree(child))
		}
		return &And{Children: children}
	case *protocol.Or:
		children := make([]MatchTree, 0, len(v.Children))
		for _, child := range v.Children {
			children = append(children, ToMatchTree(child))
		}
		return &Or{Children: children}
	case *protocol.Not:
		return &Not{Child: ToMatchTree(v.Child)}
	default:
		panic(fmt.Sprintf("unknown protocol query type %T", q))
	}
}

// MatchTree is an interface representing the queries we can run against a commit.
type MatchTree interface {
	// Match returns whether the given predicate matches a commit and, if it does,
	// the portions of the commit that match in the form of *CommitHighlights
	Match(*LazyCommit) (matched bool, highlights *protocol.CommitHighlights, err error)
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	protocol.AuthorMatches
}

func (a *AuthorMatches) Match(cv *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	return a.Regexp.Match(cv.AuthorName) || a.Regexp.Match(cv.AuthorEmail), nil, nil
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	protocol.CommitterMatches
}

func (c *CommitterMatches) Match(cv *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	return c.Regexp.Match(cv.CommitterName) || c.Regexp.Match(cv.CommitterEmail), nil, nil
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	protocol.CommitBefore
}

func (c *CommitBefore) Match(lc *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, nil, err
	}
	return authorDate.Before(c.Time), nil, nil
}

// CommitAfter is a predicate that matches if the commit is after the given date
type CommitAfter struct {
	protocol.CommitAfter
}

func (c *CommitAfter) Match(lc *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, nil, err
	}
	return authorDate.After(c.Time), nil, nil
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	protocol.MessageMatches
}

func (m *MessageMatches) Match(commit *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	results := m.FindAllIndex(commit.Message, -1)
	if results == nil {
		return false, nil, nil
	}

	return true, &protocol.CommitHighlights{
		Message: matchesToRanges(commit.Message, results),
	}, nil
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	protocol.DiffMatches
}

func (dm *DiffMatches) Match(commit *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	diff, err := commit.Diff()
	if err != nil {
		return false, nil, err
	}

	foundMatch := false

	var fileDiffHighlights map[int]protocol.FileDiffHighlight
	for fileIdx, fileDiff := range diff {
		var hunkHighlights map[int]protocol.HunkHighlight
		for hunkIdx, hunk := range fileDiff.Hunks {
			var lineHighlights map[int]protocol.Ranges
			for lineIdx, line := range bytes.Split(hunk.Body, []byte("\n")) {
				if len(line) == 0 {
					continue
				}

				origin, lineWithoutPrefix := line[0], line[1:]
				switch origin {
				case '+', '-':
				default:
					continue
				}

				matches := dm.FindAllIndex(lineWithoutPrefix, -1)
				if matches != nil {
					foundMatch = true
					if lineHighlights == nil {
						lineHighlights = make(map[int]protocol.Ranges, 1)
					}
					lineHighlights[lineIdx] = matchesToRanges(lineWithoutPrefix, matches)
				}
			}

			if len(lineHighlights) > 0 {
				if hunkHighlights == nil {
					hunkHighlights = make(map[int]protocol.HunkHighlight, 1)
				}
				hunkHighlights[hunkIdx] = protocol.HunkHighlight{lineHighlights}
			}
		}
		if len(hunkHighlights) > 0 {
			if fileDiffHighlights == nil {
				fileDiffHighlights = make(map[int]protocol.FileDiffHighlight)
			}
			fileDiffHighlights[fileIdx] = protocol.FileDiffHighlight{HunkHighlights: hunkHighlights}
		}
	}

	return foundMatch, &protocol.CommitHighlights{
		Diff: fileDiffHighlights,
	}, nil
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	protocol.DiffModifiesFile
}

func (dmf *DiffModifiesFile) Match(commit *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	diff, err := commit.Diff()
	if err != nil {
		return false, nil, err
	}

	foundMatch := false
	var fileDiffHighlights map[int]protocol.FileDiffHighlight
	for fileIdx, fileDiff := range diff {
		oldFileMatches := dmf.FindAllStringIndex(fileDiff.OrigName, -1)
		newFileMatches := dmf.FindAllStringIndex(fileDiff.NewName, -1)
		if oldFileMatches != nil || newFileMatches != nil {
			if fileDiffHighlights == nil {
				fileDiffHighlights = make(map[int]protocol.FileDiffHighlight)
			}
			foundMatch = true
			fileDiffHighlights[fileIdx] = protocol.FileDiffHighlight{
				OldFile: matchesToRanges([]byte(fileDiff.OrigName), oldFileMatches),
				NewFile: matchesToRanges([]byte(fileDiff.NewName), newFileMatches),
			}
		}
	}

	return foundMatch, &protocol.CommitHighlights{
		Diff: fileDiffHighlights,
	}, nil
}

// And is a predicate that matches if all of its children predicates match
type And struct {
	Children []MatchTree
}

func (a *And) Match(commit *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	highlights := &protocol.CommitHighlights{}
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
	Children []MatchTree
}

func (o *Or) Match(commit *LazyCommit) (bool, *protocol.CommitHighlights, error) {
	hasMatch := false
	mergedHighlights := &protocol.CommitHighlights{}
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
	Child MatchTree
}

func (n *Not) Match(commit *LazyCommit) (bool, *protocol.CommitHighlights, error) {
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
func matchesToRanges(content []byte, matches [][]int) protocol.Ranges {
	// Incrementally search newlines to avoid counting newlines over the
	// entire string for every match.
	var (
		lastNewlineOffset int
		newlineCount      int
		searchEnd         int
	)
	lineAndColumn := func(offset int) (line, column int) {
		newlines, index := newlineCountAndLastIndex(content[searchEnd:offset])
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

	res := make(protocol.Ranges, 0, len(matches))
	for _, match := range matches {
		startLine, startColumn := lineAndColumn(match[0])
		endLine, endColumn := lineAndColumn(match[1])
		res = append(res, protocol.Range{
			Start: protocol.Location{Offset: match[0], Line: startLine, Column: startColumn},
			End:   protocol.Location{Offset: match[1], Line: endLine, Column: endColumn},
		})
	}
	return res
}

func newlineCountAndLastIndex(content []byte) (count int, lastIndex int) {
	lastIndex = bytes.LastIndexByte(content, '\n')
	if lastIndex == -1 {
		return 0, -1
	}

	return bytes.Count(content[:lastIndex], []byte("\n")) + 1, lastIndex
}
