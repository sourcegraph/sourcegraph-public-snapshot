package search

import (
	"strings"

	git "github.com/libgit2/git2go/v31"
)

type Commit struct {
	*git.Commit
	repo *git.Repository

	// These fields may not be initialized, so use their accessor methods instead
	diff      []DiffDelta    // lazily populated by Diff()
	author    *git.Signature // lazily populated by Author()
	committer *git.Signature // lazily populated by Committer()
}

func (c *Commit) Author() *git.Signature {
	if c.author != nil {
		return c.author
	}
	return c.Commit.Author()
}

func (c *Commit) Committer() *git.Signature {
	if c.committer != nil {
		return c.committer
	}
	return c.Commit.Committer()
}

func (c *Commit) Diff() ([]DiffDelta, error) {
	if c.diff != nil {
		return c.diff, nil
	}
	var (
		parentTree *git.Tree
		err        error
	)
	if c.ParentCount() > 0 {
		parentTree, err = c.Parent(0).Tree()
		if err != nil {
			return nil, err
		}
	}

	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}

	// If the first commit in the repo, parentTree is nil, which is fine.
	// diffOptions := git.DiffOptions{
	// 	ContextLines:   2,
	// 	InterhunkLines: 2,
	// 	IdAbbrev:       8,
	// 	// TODO max size?
	// }
	diff, err := c.repo.DiffTreeToTree(parentTree, tree, nil)
	if err != nil {
		return nil, err
	}

	var (
		res          []DiffDelta
		currentDelta *DiffDelta
		currentHunk  *DiffHunk
	)

	diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		if currentDelta != nil {
			res = append(res, *currentDelta)
		}
		currentDelta = &DiffDelta{
			DiffDelta: &delta,
		}
		return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
			if currentHunk != nil {
				currentDelta.Hunks = append(currentDelta.Hunks, *currentHunk)
			}
			currentHunk = &DiffHunk{
				DiffHunk: &hunk,
			}
			return func(line git.DiffLine) error {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{&line})
				return nil
			}, nil
		}, nil
	}, git.DiffDetailLines)

	if currentHunk != nil {
		currentDelta.Hunks = append(currentDelta.Hunks, *currentHunk)
	}

	if currentDelta != nil {
		res = append(res, *currentDelta)
	}

	return res, nil
}

func IterCommits(repo *git.Repository, rev *git.Oid, fn func(commit *Commit) bool) error {
	walker, err := repo.Walk()
	if err != nil {
		return err
	}

	if err = walker.Push(rev); err != nil {
		return err
	}

	return walker.Iterate(func(commit *git.Commit) bool {
		return fn(&Commit{
			Commit: commit,
			repo:   repo,
		})
	})
}

func offset(ranges Ranges, amount int) Ranges {
	res := make(Ranges, 0, len(ranges))
	for _, oldRange := range ranges {
		res = append(res, Range{
			Start: Location{Offset: oldRange.Start.Offset + amount},
			End:   Location{Offset: oldRange.End.Offset + amount},
		})
	}
	return res
}

type DiffDelta struct {
	*git.DiffDelta
	Hunks []DiffHunk
}

type DiffHunk struct {
	*git.DiffHunk
	Lines []DiffLine
}

type DiffLine struct {
	*git.DiffLine
}

func FormatDiffWithHighlights(deltas []DiffDelta, deltaHighlights DeltaHighlights) (string, Ranges) {
	var buf strings.Builder
	var ranges Ranges
	buf.Grow(1024) // conservative minimum to reduce many small allocations

	for _, dh := range deltaHighlights {
		delta := deltas[dh.Index]

		ranges = append(ranges, offset(dh.OldFileHighlights, buf.Len())...)
		buf.WriteString(delta.OldFile.Path)
		buf.WriteByte(' ')

		ranges = append(ranges, offset(dh.NewFileHighlights, buf.Len())...)
		buf.WriteString(delta.NewFile.Path)
		buf.WriteByte('\n')

		for _, hh := range dh.Hunks {
			hunk := delta.Hunks[hh.Index]

			buf.WriteString(hunk.Header)
			buf.WriteByte('\n')

			contextLinesToWrite := make(map[int]struct{}, len(hh.Lines))
			// Populate all the possible context line indices
			for _, lh := range hh.Lines {
				prev := lh.Index - 1
				if prev >= 0 {
					contextLinesToWrite[prev] = struct{}{}
				}
				next := lh.Index + 1
				if next < len(hunk.Lines) {
					contextLinesToWrite[next] = struct{}{}
				}
			}

			// Remove all the requested context lines that are highlighted lines
			for _, lh := range hh.Lines {
				delete(contextLinesToWrite, lh.Index)
			}

			// Write all the lines and the context lines surrounding them
			for _, lh := range hh.Lines {
				prev := lh.Index - 1
				if _, ok := contextLinesToWrite[prev]; ok {
					writeLine(&buf, hunk.Lines[prev])
					delete(contextLinesToWrite, prev)
				}

				ranges = append(ranges, offset(lh.Highlights, buf.Len()+1)...)
				writeLine(&buf, hunk.Lines[lh.Index])

				next := lh.Index + 1
				if _, ok := contextLinesToWrite[next]; ok {
					writeLine(&buf, hunk.Lines[next])
					delete(contextLinesToWrite, next)
				}
			}

			if len(contextLinesToWrite) > 0 {
				panic("all context lines should have been written")
			}
		}
	}
	return buf.String(), ranges
}

func writeLine(buf *strings.Builder, line DiffLine) {
	switch line.Origin {
	case git.DiffLineContext:
		buf.WriteByte(' ')
		buf.WriteString(line.Content)
	case git.DiffLineAddition:
		buf.WriteByte('+')
		buf.WriteString(line.Content)
	case git.DiffLineDeletion:
		buf.WriteByte('-')
		buf.WriteString(line.Content)
	}
}
