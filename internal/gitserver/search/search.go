package search

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	git "github.com/libgit2/git2go/v31"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// Commit wraps a *git.Commit, lazily storing requested fields to avoid re-computing
// expensive fields like diff. Additionally, it holds a reference to the repo so the
// commit can compute its own diff with its parent.
type Commit struct {
	*git.Commit
	repo *git.Repository

	// These fields may not be initialized, so use their accessor methods instead
	diff      Diff           // lazily populated by Diff()
	author    *git.Signature // lazily populated by Author()
	committer *git.Signature // lazily populated by Committer()
	message   *string        // lazily populated by Message()
}

func (c *Commit) Message() string {
	if c.message != nil {
		return *c.message
	}
	m := strings.TrimSpace(c.Commit.Message())
	c.message = &m
	return m
}

// Author returns the commit's author signature, using a cached value if it's already been called
func (c *Commit) Author() *git.Signature {
	if c.author != nil {
		return c.author
	}
	c.author = c.Commit.Author()
	return c.author
}

// Committer returns the commit's committer signature, using a cached value if it's already been called
func (c *Commit) Committer() *git.Signature {
	if c.committer != nil {
		return c.committer
	}
	c.committer = c.Commit.Committer()
	return c.committer
}

// Parents returns the IDs for this commits parents
func (c *Commit) Parents() []api.CommitID {
	parentCount := c.ParentCount()
	res := make([]api.CommitID, 0, parentCount)
	for i := uint(0); i < parentCount; i++ {
		res = append(res, api.CommitID(c.ParentId(i).String()))
	}
	return res
}

func (c *Commit) Diff() (Diff, error) {
	// If the diff has already been computed, use that
	if c.diff != "" {
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
			return "", err
		}
	}

	tree, err := c.Tree()
	if err != nil {
		return "", err
	}

	diffOptions := git.DiffOptions{
		ContextLines:   1,
		InterhunkLines: 0,
	}
	diff, err := c.repo.DiffTreeToTree(parentTree, tree, &diffOptions)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	buf.Grow(1024)

	err = diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		buf.WriteString(delta.OldFile.Path)
		buf.WriteByte('\t')
		buf.WriteString(delta.NewFile.Path)
		buf.WriteByte('\n')

		return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
			buf.WriteString(hunk.Header)

			return func(line git.DiffLine) error {
				switch line.Origin {
				case git.DiffLineContext:
					buf.WriteByte(' ')
				case git.DiffLineAddition:
					buf.WriteByte('+')
				case git.DiffLineDeletion:
					buf.WriteByte('-')
				default:
					return nil
				}

				buf.WriteString(line.Content)
				return nil
			}, nil
		}, nil
	}, git.DiffDetailLines)

	return Diff(buf.String()), err
}

type CommitMatch struct {
	*Commit
	*HighlightedCommit
}

func IterCommitMatches(ctx context.Context, repoPath string, revs []RevisionSpecifier, pred CommitPredicate, fn func(*Commit, *HighlightedCommit) bool) error {
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)

	var (
		batchSize   = 256
		jobs        = make(chan searchJob, 256)
		resultChans = make(chan chan CommitMatch, 256)
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

		// TODO(camdencheek): figure out a way to replicate git's --source
		// so we can populate SourceRefs
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

		walker.SimplifyFirstParent()

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
				if !fn(result.Commit, result.HighlightedCommit) {
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
			select {
			case <-ctx.Done():
				return nil
			case j.ResultChan <- CommitMatch{wrappedCommit, highlights}:
			}
		}
	}
	return nil
}

// TODO(camdencheek): this is copied straight from internal/search to avoid import cycles
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
