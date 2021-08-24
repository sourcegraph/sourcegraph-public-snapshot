package search

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	git "github.com/libgit2/git2go/v31"
	"golang.org/x/sync/errgroup"
)

// Commit wraps a *git.Commit, lazily storing requested fields to avoid re-computing
// expensive fields like diff. Additionally, it holds a reference to the repo so the
// commit can compute its own diff with its parent.
type Commit struct {
	*git.Commit
	repo *git.Repository

	// These fields may not be initialized, so use their accessor methods instead
	diff      []DiffDelta    // lazily populated by Diff()
	author    *git.Signature // lazily populated by Author()
	committer *git.Signature // lazily populated by Committer()
}

// Author returns the commit's author signature, using a cached value if it's already been called
func (c *Commit) Author() *git.Signature {
	if c.author != nil {
		return c.author
	}
	return c.Commit.Author()
}

// Committer returns the commit's committer signature, using a cached value if it's already been called
func (c *Commit) Committer() *git.Signature {
	if c.committer != nil {
		return c.committer
	}
	return c.Commit.Committer()
}

func (c *Commit) Diff() ([]DiffDelta, error) {
	// If the diff has already been computed, use that
	if c.diff != nil {
		return c.diff, nil
	}

	var (
		// It's okay for the parentTree to be nil if there
		// are no parent commits. This will happen with the
		// first commit in the repository, and DiffTreeToTree
		// treats that as diffing against an empty index.
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

	diffOptions := git.DiffOptions{
		ContextLines:   1,
		InterhunkLines: 0,
	}
	diff, err := c.repo.DiffTreeToTree(parentTree, tree, &diffOptions)
	if err != nil {
		return nil, err
	}

	var (
		res          []DiffDelta
		currentDelta *DiffDelta
		currentHunk  *DiffHunk
	)

	// Build a full in-memory representation of the diff because the callbacks here make it _very_
	// difficult to do any sort of processing on the diff without convoluted detection of when
	// we finish iterating over a file or a chunk.
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

	// Clean up any in-progress hunks and deltas that weren't closed because iteration doesn't
	// provide any callbacks for "finished hunk/delta".
	if currentHunk != nil {
		currentDelta.Hunks = append(currentDelta.Hunks, *currentHunk)
	}
	if currentDelta != nil {
		res = append(res, *currentDelta)
	}

	return res, nil
}

// DiffDelta wraps *git.DiffDelta with the pre-computed hunks associated with that delta
type DiffDelta struct {
	*git.DiffDelta
	Hunks []DiffHunk
}

// DiffHunk wraps *git.DiffHunk with the pre-computed lines associated with that hunk
type DiffHunk struct {
	*git.DiffHunk
	Lines []DiffLine
}

// DiffHunk wraps *git.DiffLine
type DiffLine struct {
	*git.DiffLine
}

// CommitMatch represents a matched commit and highlights of the matched fields
type CommitMatch struct {
	Commit     *Commit
	Highlights *CommitHighlights
}

func IterCommitMatches(ctx context.Context, repoPath string, revs []RevisionSpecifier, pred CommitPredicate, fn func(CommitMatch) bool) error {
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)

	var (
		batchSize   = 128
		jobs        = make(chan searchJob, 256)
		resultChans = make(chan chan CommitMatch, 64)
	)

	// Start job feeder. It will iterate over all commits that match the
	// given revisions, batch them up, and send those batches of commits
	// to the workers to be searched over.
	g.Go(func() error {
		defer close(jobs)
		defer close(resultChans)

		repo, err := git.OpenRepository(repoPath)
		if err != nil {
			return err
		}

		walker, err := repo.Walk()
		if err != nil {
			return err
		}

		for _, rev := range revs {
			if rev.RevSpec != "" {
				if err := walker.PushRange(rev.RevSpec); err != nil {
					return err
				}
			} else if rev.RefGlob != "" {
				if err := walker.PushGlob(rev.RevSpec); err != nil {
					return err
				}
			} else if rev.ExcludeRefGlob != "" {
				if err := walker.HideGlob(rev.RevSpec); err != nil {
					return err
				}
			} else {
				if err := walker.PushHead(); err != nil {
					return err
				}
			}
		}

		send := func(batch []git.Oid) {
			resultChan := make(chan CommitMatch, 32)
			j := searchJob{
				CommitBatch: batch,
				ResultChan:  resultChan,
				Predicate:   pred,
			}
			jobs <- j
			resultChans <- resultChan
		}

		batch := make([]git.Oid, batchSize)
		for {
			// Fill the batch, keeping track of how far we filled it
			for i := 0; i < len(batch); i++ {
				err := walker.Next(&batch[i])
				if git.IsErrorCode(err, git.ErrorCodeIterOver) {
					// We've hit the end of the commits, so send
					// a partial batch and return
					send(batch[:i])
					return nil
				}
				if err != nil {
					return err
				}
			}

			// Send the batch and create a new one
			send(batch)
			batch = make([]git.Oid, batchSize)
		}
	})

	// Start workers, which read off our jobs channel
	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		g.Go(func() error {
			repo, err := git.OpenRepository(repoPath)
			if err != nil {
				return err
			}

			errs := new(multierror.Error)
			for job := range jobs {
				// Every job that's created must be run in order to
				// close channels correctly. The job will exit early
				// if the context is cancelled, so only minimal extra
				// work will happen after cancellation. This is also why
				// we collect errors here rather than returning immediately
				// on error.
				if err := job.Run(ctx, repo); err != nil {
					errs = multierror.Append(errs, err)
				}
			}
			return errs.ErrorOrNil()
		})
	}

	// Start consumer, which will read off our results channels (in order),
	// calling the callback for each result it receives.
	g.Go(func() error {
		defer cancel()
		for resultChan := range resultChans {
			for result := range resultChan {
				if !fn(result) {
					return nil
				}
			}
		}
		return nil
	})

	return g.Wait()
}

type searchJob struct {
	CommitBatch []git.Oid
	ResultChan  chan CommitMatch
	Predicate   CommitPredicate
}

func (j searchJob) Run(ctx context.Context, repo *git.Repository) error {
	defer close(j.ResultChan)

	for _, oid := range j.CommitBatch {
		// Early return if context has already been canceled
		if ctx.Err() != nil {
			return nil
		}

		commit, err := repo.LookupCommit(&oid)
		if err != nil {
			return err
		}

		wrappedCommit := &Commit{
			Commit: commit,
			repo:   repo,
		}
		matched, highlights := j.Predicate.Match(wrappedCommit)
		if matched {
			commitMatch := CommitMatch{
				Commit:     wrappedCommit,
				Highlights: highlights,
			}
			select {
			case <-ctx.Done():
				return nil
			case j.ResultChan <- commitMatch:
			}
		}
	}
	return nil
}

// FormatDiffWithHighlights formats a diff into a string, keeping only lines that were highlighted
// and one line of context around them. It returns the formatted diff and a set of new highlights that
// map to the formatted diff rather than the original []DiffDelta. If we match a delta that has no hunk
// matches, we include the full diff content for that file.
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

		// if len(dh.Hunks) == 0 {
		// 	// TODO format all hunks if none of the hunks matched
		// }

		for _, hh := range dh.Hunks {
			hunk := delta.Hunks[hh.Index]

			buf.WriteString(hunk.Header)

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

				// offset by buf.Len()+1 because each line will be prepended with '+', '-', or ' '
				ranges = append(ranges, offset(lh.Highlights, buf.Len()+1)...)
				writeLine(&buf, hunk.Lines[lh.Index])

				next := lh.Index + 1
				if _, ok := contextLinesToWrite[next]; ok {
					writeLine(&buf, hunk.Lines[next])
					delete(contextLinesToWrite, next)
				}
			}

			if len(contextLinesToWrite) > 0 {
				// TODO(camdencheek): This isn't actually necessary, but it's a nice safety
				// check to keep around for a while to check my assumptions.
				panic("all context lines should have been written")
			}
		}
	}
	return buf.String(), ranges
}

// writeLine writes a DiffLine to the builder, prepended with:
// '+' for an added line
// '-' for a removed line
// ' ' for a context line
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

// offset takes a set of ranges and creates a new set of ranges whose
// start and end locations are offset by the given amount. Note, we could
// mutate the ranges in place to avoid an allocation, but it's a relatively
// small cost, and formatting a diff shouldn't mutate it.
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

// TODO(camdencheek) this is copied straight from internal/search to avoid import cycles
type RevisionSpecifier struct {
	// RevSpec is a revision range specifier suitable for passing to git. See
	// the manpage gitrevisions(7).
	RevSpec string

	// RefGlob is a reference glob to pass to git. See the documentation for
	// "--glob" in git-log.
	RefGlob string

	// ExcludeRefGlob is a glob for references to exclude. See the
	// documentation for "--exclude" in git-log.
	ExcludeRefGlob string
}
