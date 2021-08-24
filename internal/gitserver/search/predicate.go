package search

import (
	"regexp"
	"time"
)

// CommitPredicate is an interface representing the queries we can run against a commit.
type CommitPredicate interface {
	// Match returns whether the given predicate matches a commit and, if it does,
	// the portions of the commit that match in the form of *CommitHighlights
	Match(*Commit) (matched bool, highlights *CommitHighlights)
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	Regexp
}

func (a *AuthorMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	author := commit.Author()
	return a.MatchString(author.Name) || a.MatchString(author.Email), nil
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	Regexp
}

func (c *CommitterMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	committer := commit.Committer()
	return c.MatchString(committer.Name) || c.MatchString(committer.Email), nil
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	time.Time
}

func (c *CommitBefore) Match(commit *Commit) (bool, *CommitHighlights) {
	return commit.Author().When.Before(c.Time), nil
}

// CommitAfter is a predicate that matches if the commit is after the given date
type CommitAfter struct {
	time.Time
}

func (c *CommitAfter) Match(commit *Commit) (bool, *CommitHighlights) {
	return commit.Author().When.After(c.Time), nil
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	Regexp
}

func (m *MessageMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	results := m.FindAllStringIndex(commit.Message(), -1) // TODO limit?
	if results == nil {
		return false, nil
	}

	return true, &CommitHighlights{
		Message: matchesToRanges(results),
	}
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	Regexp
}

func (d *DiffMatches) Match(commit *Commit) (bool, *CommitHighlights) {
	diff, err := commit.Diff()
	if err != nil {
		// TODO(camdencheek) don't ignore error
		return false, nil
	}

	var deltaHighlights DeltaHighlights
	var foundMatches bool
	for i, delta := range diff {
		var hunkHighlights HunkHighlights
		for j, hunk := range delta.Hunks {
			var lineHighlights LineHighlights
			for k, line := range hunk.Lines {
				matches := d.FindAllStringIndex(line.Content, -1)
				if matches != nil {
					foundMatches = true
					lineHighlights = append(lineHighlights, LineHighlight{
						Index:      k,
						Highlights: matchesToRanges(matches),
					})
				}
			}
			if len(lineHighlights) > 0 {
				hunkHighlights = append(hunkHighlights, HunkHighlight{
					Index: j,
					Lines: lineHighlights,
				})
			}
		}
		if len(hunkHighlights) > 0 {
			deltaHighlights = append(deltaHighlights, DeltaHighlight{
				Index: i,
				Hunks: hunkHighlights,
			})
		}
	}

	return foundMatches, &CommitHighlights{
		Diff: deltaHighlights,
	}
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	Regexp
}

func (d *DiffModifiesFile) Match(commit *Commit) (bool, *CommitHighlights) {
	diff, err := commit.Diff()
	if err != nil {
		// TODO is ignoring okay, or should the Match() signature return an error?
		return false, nil
	}

	foundMatch := false
	var deltaHighlights DeltaHighlights
	for i, delta := range diff {
		oldFileMatches := d.FindAllStringIndex(delta.OldFile.Path, -1)
		newFileMatches := d.FindAllStringIndex(delta.NewFile.Path, -1)
		if oldFileMatches != nil || newFileMatches != nil {
			foundMatch = true
			deltaHighlights = append(deltaHighlights, DeltaHighlight{
				Index:             i,
				OldFileHighlights: matchesToRanges(oldFileMatches),
				NewFileHighlights: matchesToRanges(oldFileMatches),
			})
		}
	}
	return foundMatch, &CommitHighlights{Diff: deltaHighlights}
}

// And is a predicate that matches if all of its children predicates match
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

// Or is a predicate that matches if any of its children predicates match
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

// Not is a predicate that matches if its child predicate does not match
type Not struct {
	CommitPredicate
}

func (n *Not) Match(commit *Commit) (bool, *CommitHighlights) {
	// Even if the child highlights, since we're negating, the match shouldn't be highlighted
	foundMatch, _ := n.CommitPredicate.Match(commit)
	return !foundMatch, nil
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
