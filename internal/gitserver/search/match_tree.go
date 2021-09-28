package search

import (
	"bytes"
	"unicode/utf8"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
)

// ToMatchTree converts a protocol.SearchQuery into its equivalent MatchTree.
// We don't send a match tree directly over the wire so using the protocol
// package doesn't pull in all the dependencies that the match tree needs.
func ToMatchTree(q protocol.SearchQuery) (MatchTree, error) {
	switch v := q.(type) {
	case *protocol.CommitBefore:
		return &CommitBefore{*v}, nil
	case *protocol.CommitAfter:
		return &CommitAfter{*v}, nil
	case *protocol.AuthorMatches:
		re, err := casetransform.CompileRegexp(v.Expr, v.IgnoreCase)
		return &AuthorMatches{re}, err
	case *protocol.CommitterMatches:
		re, err := casetransform.CompileRegexp(v.Expr, v.IgnoreCase)
		return &CommitterMatches{re}, err
	case *protocol.MessageMatches:
		re, err := casetransform.CompileRegexp(v.Expr, v.IgnoreCase)
		return &MessageMatches{re}, err
	case *protocol.DiffMatches:
		re, err := casetransform.CompileRegexp(v.Expr, v.IgnoreCase)
		return &DiffMatches{re}, err
	case *protocol.DiffModifiesFile:
		re, err := casetransform.CompileRegexp(v.Expr, v.IgnoreCase)
		return &DiffModifiesFile{re}, err
	case *protocol.And:
		children := make([]MatchTree, 0, len(v.Children))
		for _, child := range v.Children {
			sub, err := ToMatchTree(child)
			if err != nil {
				return nil, err
			}
			children = append(children, sub)
		}
		return &And{Children: children}, nil
	case *protocol.Or:
		children := make([]MatchTree, 0, len(v.Children))
		for _, child := range v.Children {
			sub, err := ToMatchTree(child)
			if err != nil {
				return nil, err
			}
			children = append(children, sub)
		}
		return &Or{Children: children}, nil
	case *protocol.Not:
		sub, err := ToMatchTree(v.Child)
		if err != nil {
			return nil, err
		}
		return &Not{Child: sub}, nil
	default:
		return nil, errors.Errorf("unknown protocol query type %T", q)
	}
}

// MatchTree is an interface representing the queries we can run against a commit.
type MatchTree interface {
	// Match returns whether the given predicate matches a commit and, if it does,
	// the portions of the commit that match in the form of *CommitHighlights
	Match(*LazyCommit) (matched bool, highlights *CommitHighlights, err error)
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	*casetransform.Regexp
}

func (a *AuthorMatches) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
	return a.Regexp.Match(lc.AuthorName, &lc.LowerBuf) || a.Regexp.Match(lc.AuthorEmail, &lc.LowerBuf), nil, nil
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	*casetransform.Regexp
}

func (c *CommitterMatches) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
	return c.Regexp.Match(lc.CommitterName, &lc.LowerBuf) || c.Regexp.Match(lc.CommitterEmail, &lc.LowerBuf), nil, nil
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	protocol.CommitBefore
}

func (c *CommitBefore) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
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

func (c *CommitAfter) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, nil, err
	}
	return authorDate.After(c.Time), nil, nil
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	*casetransform.Regexp
}

func (m *MessageMatches) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
	results := m.FindAllIndex(lc.Message, -1, &lc.LowerBuf)
	if results == nil {
		return false, nil, nil
	}

	return true, &CommitHighlights{
		Message: matchesToRanges(lc.Message, results),
	}, nil
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	*casetransform.Regexp
}

func (dm *DiffMatches) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
	diff, err := lc.Diff()
	if err != nil {
		return false, nil, err
	}

	foundMatch := false

	var fileDiffHighlights map[int]FileDiffHighlight
	for fileIdx, fileDiff := range diff {
		var hunkHighlights map[int]HunkHighlight
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

				matches := dm.FindAllIndex(lineWithoutPrefix, -1, &lc.LowerBuf)
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
					hunkHighlights = make(map[int]HunkHighlight, 1)
				}
				hunkHighlights[hunkIdx] = HunkHighlight{lineHighlights}
			}
		}
		if len(hunkHighlights) > 0 {
			if fileDiffHighlights == nil {
				fileDiffHighlights = make(map[int]FileDiffHighlight)
			}
			fileDiffHighlights[fileIdx] = FileDiffHighlight{HunkHighlights: hunkHighlights}
		}
	}

	return foundMatch, &CommitHighlights{
		Diff: fileDiffHighlights,
	}, nil
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	*casetransform.Regexp
}

func (dmf *DiffModifiesFile) Match(lc *LazyCommit) (bool, *CommitHighlights, error) {
	diff, err := lc.Diff()
	if err != nil {
		return false, nil, err
	}

	foundMatch := false
	var fileDiffHighlights map[int]FileDiffHighlight
	for fileIdx, fileDiff := range diff {
		oldFileMatches := dmf.FindAllIndex([]byte(fileDiff.OrigName), -1, &lc.LowerBuf)
		newFileMatches := dmf.FindAllIndex([]byte(fileDiff.NewName), -1, &lc.LowerBuf)
		if oldFileMatches != nil || newFileMatches != nil {
			if fileDiffHighlights == nil {
				fileDiffHighlights = make(map[int]FileDiffHighlight)
			}
			foundMatch = true
			fileDiffHighlights[fileIdx] = FileDiffHighlight{
				OldFile: matchesToRanges([]byte(fileDiff.OrigName), oldFileMatches),
				NewFile: matchesToRanges([]byte(fileDiff.NewName), newFileMatches),
			}
		}
	}

	return foundMatch, &CommitHighlights{
		Diff: fileDiffHighlights,
	}, nil
}

// And is a predicate that matches if all of its children predicates match
type And struct {
	Children []MatchTree
}

func (a *And) Match(commit *LazyCommit) (bool, *CommitHighlights, error) {
	highlights := &CommitHighlights{}
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

func (o *Or) Match(commit *LazyCommit) (bool, *CommitHighlights, error) {
	hasMatch := false
	mergedHighlights := &CommitHighlights{}
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

func (n *Not) Match(commit *LazyCommit) (bool, *CommitHighlights, error) {
	// Even if the child highlights, since we're negating, the match shouldn't be highlighted
	foundMatch, _, err := n.Child.Match(commit)
	return !foundMatch, nil, err
}

// matchesToRanges is a helper that takes the return value of regexp.FindAllStringIndex()
// and converts it to Ranges.
// INVARIANT: matches must be ordered and non-overlapping,
// which is guaranteed by regexp.FindAllIndex()
func matchesToRanges(content []byte, matches [][]int) protocol.Ranges {
	var (
		unscannedOffset          = 0
		scannedNewlines          = 0
		scannedRunes             = 0
		lastScannedNewlineOffset = -1
	)

	lineColumnOffset := func(byteOffset int) (line, column, offset int) {
		unscanned := content[unscannedOffset:byteOffset]
		scannedRunes += utf8.RuneCount(unscanned)
		lastUnscannedNewlineOffset := bytes.LastIndexByte(unscanned, '\n')
		if lastUnscannedNewlineOffset != -1 {
			lastScannedNewlineOffset = unscannedOffset + lastUnscannedNewlineOffset
			scannedNewlines += bytes.Count(unscanned, []byte("\n"))
		}
		column = utf8.RuneCount(content[lastScannedNewlineOffset+1 : byteOffset])
		unscannedOffset = byteOffset
		return scannedNewlines, column, scannedRunes
	}

	res := make(protocol.Ranges, 0, len(matches))
	for _, match := range matches {
		startLine, startColumn, startOffset := lineColumnOffset(match[0])
		endLine, endColumn, endOffset := lineColumnOffset(match[1])
		res = append(res, protocol.Range{
			Start: protocol.Location{Line: startLine, Column: startColumn, Offset: startOffset},
			End:   protocol.Location{Line: endLine, Column: endColumn, Offset: endOffset},
		})
	}
	return res
}
