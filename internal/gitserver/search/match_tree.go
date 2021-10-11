package search

import (
	"bytes"
	"unicode/utf8"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// ToMatchTree converts a protocol.SearchQuery into its equivalent MatchTree.
// We don't send a match tree directly over the wire so using the protocol
// package doesn't pull in all the dependencies that the match tree needs.
func ToMatchTree(q protocol.Node) (MatchTree, error) {
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
	case *protocol.Boolean:
		return &Constant{v.Value}, nil
	case *protocol.Operator:
		operands := make([]MatchTree, 0, len(v.Operands))
		for _, operand := range v.Operands {
			sub, err := ToMatchTree(operand)
			if err != nil {
				return nil, err
			}
			operands = append(operands, sub)
		}
		return &Operator{Kind: v.Kind, Operands: operands}, nil
	default:
		return nil, errors.Errorf("unknown protocol query type %T", q)
	}
}

// MatchTree is an interface representing the queries we can run against a commit.
type MatchTree interface {
	// Match returns whether the given predicate matches a commit and, if it does,
	// the portions of the commit that match in the form of *CommitHighlights
	Match(*LazyCommit, MatchTree) (matched bool, highlights MatchedCommit, err error)

	// MatchFileDiff executes the query against the portion of a diff applying to a single file.
	// This method is necessary because DiffModifiesFile and DiffMatches are not independent when applied
	// to a full diff. When matching the full diff, you may get results where one file diff matches
	// the DiffModifiesFile node, and a different file diff matches the DiffMatches node. However,
	// when users search "type:diff file:a b", they're probably looking for diffs that contain the modification "b"
	// within the file "a", not diffs that contain the modification "b" and also happen to modify the file "a".
	MatchFileDiff(fileDiff *diff.FileDiff, lowerBuf *[]byte) (matched bool, highlights MatchedFileDiff, err error)
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	*casetransform.Regexp
}

func (a *AuthorMatches) Match(lc *LazyCommit, _ MatchTree) (bool, MatchedCommit, error) {
	return a.Regexp.Match(lc.AuthorName, &lc.LowerBuf) || a.Regexp.Match(lc.AuthorEmail, &lc.LowerBuf), MatchedCommit{}, nil
}

func (a *AuthorMatches) MatchFileDiff(_ *diff.FileDiff, _ *[]byte) (bool, MatchedFileDiff, error) {
	return true, MatchedFileDiff{}, nil
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	*casetransform.Regexp
}

func (c *CommitterMatches) Match(lc *LazyCommit, _ MatchTree) (bool, MatchedCommit, error) {
	return c.Regexp.Match(lc.CommitterName, &lc.LowerBuf) || c.Regexp.Match(lc.CommitterEmail, &lc.LowerBuf), MatchedCommit{}, nil
}

func (c *CommitterMatches) MatchFileDiff(_ *diff.FileDiff, _ *[]byte) (bool, MatchedFileDiff, error) {
	return true, MatchedFileDiff{}, nil
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	protocol.CommitBefore
}

func (c *CommitBefore) Match(lc *LazyCommit, _ MatchTree) (bool, MatchedCommit, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, MatchedCommit{}, err
	}
	return authorDate.Before(c.Time), MatchedCommit{}, nil
}

func (c *CommitBefore) MatchFileDiff(_ *diff.FileDiff, _ *[]byte) (bool, MatchedFileDiff, error) {
	return true, MatchedFileDiff{}, nil
}

// CommitAfter is a predicate that matches if the commit is after the given date
type CommitAfter struct {
	protocol.CommitAfter
}

func (c *CommitAfter) Match(lc *LazyCommit, _ MatchTree) (bool, MatchedCommit, error) {
	authorDate, err := lc.AuthorDate()
	if err != nil {
		return false, MatchedCommit{}, err
	}
	return authorDate.After(c.Time), MatchedCommit{}, nil
}

func (c *CommitAfter) MatchFileDiff(_ *diff.FileDiff, _ *[]byte) (bool, MatchedFileDiff, error) {
	return true, MatchedFileDiff{}, nil
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	*casetransform.Regexp
}

func (m *MessageMatches) Match(lc *LazyCommit, _ MatchTree) (bool, MatchedCommit, error) {
	results := m.FindAllIndex(lc.Message, -1, &lc.LowerBuf)
	if results == nil {
		return false, MatchedCommit{}, nil
	}

	return true, MatchedCommit{
		Message: matchesToRanges(lc.Message, results),
	}, nil
}

func (m *MessageMatches) MatchFileDiff(_ *diff.FileDiff, _ *[]byte) (bool, MatchedFileDiff, error) {
	return true, MatchedFileDiff{}, nil
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	*casetransform.Regexp
}

func (dm *DiffMatches) Match(lc *LazyCommit, q MatchTree) (bool, MatchedCommit, error) {
	diff, err := lc.Diff()
	if err != nil {
		return false, MatchedCommit{}, err
	}

	var fileDiffHighlights map[int]MatchedFileDiff
	for fileIdx, fileDiff := range diff {
		matched, matches, err := q.MatchFileDiff(fileDiff, &lc.LowerBuf)
		if err != nil {
			return false, MatchedCommit{}, err
		}
		if !matched {
			continue
		}
		if fileDiffHighlights == nil {
			fileDiffHighlights = make(map[int]MatchedFileDiff, 1)
		}
		fileDiffHighlights[fileIdx] = matches
	}

	return len(fileDiffHighlights) > 0, MatchedCommit{Diff: fileDiffHighlights}, nil
}

func (dm *DiffMatches) MatchFileDiff(fileDiff *diff.FileDiff, lowerBuf *[]byte) (bool, MatchedFileDiff, error) {
	var hunkHighlights map[int]MatchedHunk
	for hunkIdx, hunk := range fileDiff.Hunks {
		var lineHighlights map[int]result.Ranges
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

			matches := dm.FindAllIndex(lineWithoutPrefix, -1, lowerBuf)
			if matches != nil {
				if lineHighlights == nil {
					lineHighlights = make(map[int]result.Ranges, 1)
				}
				lineHighlights[lineIdx] = matchesToRanges(lineWithoutPrefix, matches)
			}
		}

		if len(lineHighlights) > 0 {
			if hunkHighlights == nil {
				hunkHighlights = make(map[int]MatchedHunk, 1)
			}
			hunkHighlights[hunkIdx] = MatchedHunk{lineHighlights}
		}
	}
	if len(hunkHighlights) > 0 {
		return true, MatchedFileDiff{MatchedHunks: hunkHighlights}, nil
	}
	return false, MatchedFileDiff{}, nil
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	*casetransform.Regexp
}

func (dmf *DiffModifiesFile) Match(lc *LazyCommit, mt MatchTree) (bool, MatchedCommit, error) {
	diff, err := lc.Diff()
	if err != nil {
		return false, MatchedCommit{}, err
	}

	var fileDiffHighlights map[int]MatchedFileDiff
	for fileIdx, fileDiff := range diff {
		matched, matchedFileDiff, err := mt.MatchFileDiff(fileDiff, &lc.LowerBuf)
		if err != nil {
			return false, MatchedCommit{}, err
		}
		if !matched {
			continue
		}
		if fileDiffHighlights == nil {
			fileDiffHighlights = make(map[int]MatchedFileDiff, 1)
		}
		fileDiffHighlights[fileIdx] = matchedFileDiff
	}

	return len(fileDiffHighlights) > 0, MatchedCommit{Diff: fileDiffHighlights}, nil
}

func (dmf *DiffModifiesFile) MatchFileDiff(fileDiff *diff.FileDiff, lowerBuf *[]byte) (bool, MatchedFileDiff, error) {
	oldFileMatches := dmf.FindAllIndex([]byte(fileDiff.OrigName), -1, lowerBuf)
	newFileMatches := dmf.FindAllIndex([]byte(fileDiff.NewName), -1, lowerBuf)
	if oldFileMatches != nil || newFileMatches != nil {
		return true, MatchedFileDiff{
			OldFile: matchesToRanges([]byte(fileDiff.OrigName), oldFileMatches),
			NewFile: matchesToRanges([]byte(fileDiff.NewName), newFileMatches),
		}, nil
	}
	return false, MatchedFileDiff{}, nil
}

type Constant struct {
	Value bool
}

func (c *Constant) Match(_ *LazyCommit, _ MatchTree) (bool, MatchedCommit, error) {
	return c.Value, MatchedCommit{}, nil
}

func (c *Constant) MatchFileDiff(_ *diff.FileDiff, _ *[]byte) (bool, MatchedFileDiff, error) {
	return c.Value, MatchedFileDiff{}, nil
}

type Operator struct {
	Kind     protocol.OperatorKind
	Operands []MatchTree
}

func (o *Operator) Match(commit *LazyCommit, mt MatchTree) (bool, MatchedCommit, error) {
	switch o.Kind {
	case protocol.Not:
		matched, _, err := o.Operands[0].Match(commit, mt)
		if err != nil {
			return false, MatchedCommit{}, err
		}
		return !matched, MatchedCommit{}, nil
	case protocol.And:
		resultMatches := MatchedCommit{}
		for _, operand := range o.Operands {
			matched, matches, err := operand.Match(commit, mt)
			if err != nil {
				return false, MatchedCommit{}, err
			}
			if !matched {
				return false, MatchedCommit{}, nil
			}
			resultMatches = resultMatches.Merge(matches)
		}
		return true, resultMatches, nil
	case protocol.Or:
		resultMatches := MatchedCommit{}
		hasMatch := false
		for _, operand := range o.Operands {
			matched, matches, err := operand.Match(commit, mt)
			if err != nil {
				return false, MatchedCommit{}, err
			}
			if matched {
				hasMatch = true
				resultMatches = resultMatches.Merge(matches)
			}
		}
		return hasMatch, resultMatches, nil
	default:
		panic("invalid operator kind")
	}
}

func (o *Operator) MatchFileDiff(fileDiff *diff.FileDiff, lowerBuf *[]byte) (bool, MatchedFileDiff, error) {
	switch o.Kind {
	case protocol.Not:
		matched, _, err := o.Operands[0].MatchFileDiff(fileDiff, lowerBuf)
		if err != nil {
			return false, MatchedFileDiff{}, err
		}
		return !matched, MatchedFileDiff{}, nil
	case protocol.And:
		resultMatches := MatchedFileDiff{}
		for _, operand := range o.Operands {
			matched, matches, err := operand.MatchFileDiff(fileDiff, lowerBuf)
			if err != nil {
				return false, MatchedFileDiff{}, err
			}
			if !matched {
				return false, MatchedFileDiff{}, nil
			}
			resultMatches = resultMatches.Merge(matches)
		}
		return true, resultMatches, nil
	case protocol.Or:
		resultMatches := MatchedFileDiff{}
		hasMatch := false
		for _, operand := range o.Operands {
			matched, matches, err := operand.MatchFileDiff(fileDiff, lowerBuf)
			if err != nil {
				return false, MatchedFileDiff{}, err
			}
			if matched {
				hasMatch = true
				resultMatches = resultMatches.Merge(matches)
			}
		}
		return hasMatch, resultMatches, nil
	default:
		panic("invalid operator kind")
	}

}

// matchesToRanges is a helper that takes the return value of regexp.FindAllStringIndex()
// and converts it to Ranges.
// INVARIANT: matches must be ordered and non-overlapping,
// which is guaranteed by regexp.FindAllIndex()
func matchesToRanges(content []byte, matches [][]int) result.Ranges {
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

	res := make(result.Ranges, 0, len(matches))
	for _, match := range matches {
		startLine, startColumn, startOffset := lineColumnOffset(match[0])
		endLine, endColumn, endOffset := lineColumnOffset(match[1])
		res = append(res, result.Range{
			Start: result.Location{Line: startLine, Column: startColumn, Offset: startOffset},
			End:   result.Location{Line: endLine, Column: endColumn, Offset: endOffset},
		})
	}
	return res
}
