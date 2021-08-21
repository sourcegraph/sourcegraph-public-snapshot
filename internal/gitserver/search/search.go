package search

import (
	"strings"

	"github.com/libgit2/git2go/v31"
)

type Commit struct {
	*git.Commit
	repo *git.Repository

	// These fields may not be initialized, so use their accessor methods instead
	diff      *git.Diff      // lazily populated by Diff()
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

func (c *Commit) Diff() (*git.Diff, error) {
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
	c.diff = diff
	return diff, nil
}

func (c *Commit) FormatPatchWithHighlights(hl *CommitHighlights) (string, Ranges, error) {
	diff, err := c.Diff()
	if err != nil {
		return "", nil, err
	}

	var (
		fileNum   = -1
		hunkNum   = -1
		lineNum   = -1
		skipHunks = false
		skipLines = false
		buf       strings.Builder
		ranges    Ranges
	)

	lineCallback := func(line git.DiffLine) error {
		lineNum++
		if skipLines {
			return nil
		}

		lh, lineMatched := hl.Diff[fileNum].Hunk[hunkNum][lineNum]

		switch line.Origin {
		case git.DiffLineAddition:
			buf.WriteByte('+')
		case git.DiffLineDeletion:
			buf.WriteByte('-')
		default:
			buf.WriteByte(' ')
		}

		if lineMatched {
			lineStart := buf.Len()
			ranges = append(ranges, offset(lh, lineStart)...)
		}

		buf.WriteString(line.Content)

		return nil
	}

	hunkCallback := func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
		hunkNum++
		lineNum = -1
		skipLines = false
		if skipHunks {
			skipLines = true
			return lineCallback, nil
		}

		_, hunkMatched := hl.Diff[fileNum].Hunk[hunkNum]
		if !hunkMatched {
			skipLines = true
			return lineCallback, nil
		}

		buf.WriteString(hunk.Header)
		buf.WriteByte('\n')
		return lineCallback, nil
	}

	fileCallback := func(delta git.DiffDelta, _ float64) (git.DiffForEachHunkCallback, error) {
		fileNum++
		hunkNum = -1
		skipHunks = false

		fh, fileMatched := hl.Diff[fileNum]
		if !fileMatched {
			skipHunks = true
			return hunkCallback, nil
		}

		ranges = append(ranges, offset(fh.OldFile, buf.Len())...)

		buf.WriteString(delta.OldFile.Path)
		buf.WriteByte(' ')

		ranges = append(ranges, offset(fh.NewFile, buf.Len())...)

		buf.WriteString(delta.NewFile.Path)
		buf.WriteByte('\n')

		return hunkCallback, nil
	}

	if err := diff.ForEach(fileCallback, git.DiffDetailLines); err != nil {
		return "", nil, err
	}
	return buf.String(), nil, nil
}

func offset(ranges Ranges, amount int) Ranges {
	res := make(Ranges, len(ranges))
	for _, oldRange := range ranges {
		res = append(res, Range{
			Start: Location{Offset: oldRange.Start.Offset + amount},
			End:   Location{Offset: oldRange.End.Offset + amount},
		})
	}
	return res
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
