package search

import (
	"regexp"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/libgit2/git2go/v31"
)

type CommitPredicate interface {
	Match(*Commit) (matched bool, highlights *CommitHighlights)
}

type AuthorMatches Regexp

func (a *AuthorMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	author := commit.Author()
	return a.MatchString(author.Name) || a.MatchString(author.Email), nil
}

type CommitterMatches Regexp

func (c *CommitterMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	committer := commit.Committer()
	return c.MatchString(committer.Name) || c.MatchString(committer.Email), nil
}

type CommitBefore struct {
	time.Time
}

func (c *CommitBefore) Match(commit *Commit) (bool, *CommitHighlights) {
	return commit.Author().When.Before(c.Time), nil
}

type CommitAfter struct {
	time.Time
}

func (c *CommitAfter) Match(commit *Commit) (bool, *CommitHighlights) {
	return commit.Author().When.After(c.Time), nil
}

type MessageMatches Regexp

func (m *MessageMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	results := m.FindAllStringIndex(commit.Message(), -1) // TODO limit?
	if results == nil {
		return false, nil
	}

	return true, &CommitHighlights{
		Message: matchesToRanges(results),
	}
}

var diffFoundSentinel = errors.New("found diff")

type DiffMatches Regexp

func (d *DiffMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	diff, err := commit.Diff()
	if err != nil {
		panic(err)
	}

	var (
		foundMatch = false
		fileNum    = -1
		hunkNum    = -1
		lineNum    = -1
		highlights = &CommitHighlights{}
	)

	lineCallback := func(line git.DiffLine) error {
		lineNum++
		if line.Origin == git.DiffLineContext {
			return nil
		}

		matches := d.FindAllStringIndex(line.Content, -1)
		if matches != nil {
			foundMatch = true
			highlights.AddDiffLineMatches(fileNum, hunkNum, lineNum, matchesToRanges(matches))
		}
		return nil
	}

	hunkCallback := func(git.DiffHunk) (git.DiffForEachLineCallback, error) {
		hunkNum++
		lineNum = -1
		return lineCallback, nil
	}

	fileCallback := func(git.DiffDelta, float64) (git.DiffForEachHunkCallback, error) {
		fileNum++
		hunkNum = -1
		return hunkCallback, nil
	}

	if err = diff.ForEach(fileCallback, git.DiffDetailLines); err != nil {
		panic(err)
	}

	// TODO again, ignoring errors?
	return foundMatch, highlights
}

type DiffModifiesFile Regexp

func (d *DiffModifiesFile) Match(commit *Commit) (bool, *CommitHighlights) {
	diff, err := commit.Diff()
	if err != nil {
		// TODO is ignoring okay, or should the Match() signature return an error?
		return false, nil
	}

	var (
		foundMatch = false
		highlights = &CommitHighlights{}
		fileNum    = 0
	)
	err = diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		defer func() { fileNum++ }()

		oldFileMatches := d.FindAllStringIndex(delta.OldFile.Path, -1)
		newFileMatches := d.FindAllStringIndex(delta.NewFile.Path, -1)
		if oldFileMatches != nil || newFileMatches != nil {
			foundMatch = true
			highlights.AddFileNameHighlights(fileNum, matchesToRanges(oldFileMatches), matchesToRanges(newFileMatches))
		}
		return nil, nil
	}, git.DiffDetailLines)

	// TODO again, ignoring errors?
	return foundMatch, highlights
}

type And []CommitPredicate

func (a And) Match(commit *Commit) (bool, *CommitHighlights) {
	highlights := &CommitHighlights{}
	for _, child := range a {
		childMatched, childHighlights := child.Match(commit)
		if !childMatched {
			// Since we don't care about the highlights if we don't match all children, we can short-circuit
			return false, nil
		}
		highlights.Merge(childHighlights)
	}
	return true, highlights
}

type Or []CommitPredicate

func (o Or) Match(commit *Commit) (bool, *CommitHighlights) {
	hasMatch := false
	mergedHighlights := &CommitHighlights{}
	for _, child := range o {
		if matched, highlights := child.Match(commit); matched {
			// Because we want to highlight every match, we can't short circuit
			hasMatch = true
			mergedHighlights.Merge(highlights)
		}
	}
	return hasMatch, mergedHighlights
}

type Not struct {
	CommitPredicate
}

func (n *Not) Match(commit *Commit) (bool, *CommitHighlights) {
	// Even if the child highlights, since we're negating, the match shouldn't be highlighted
	foundMatch, _ := n.CommitPredicate.Match(commit)
	return !foundMatch, nil
}

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

func matchesToRanges(matches [][]int) Ranges {
	res := make(Ranges, 0, len(matches))
	for _, match := range matches {
		res = append(res, Range{
			Start: Location{Offset: match[0]},
			End:   Location{Offset: match[1]},
		})
	}
	return res
}
