package search

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// Git formatting directives as described in man git-log / PRETTY FORMATS
const (
	hash           = "%H"
	refNames       = "%D"
	sourceRefs     = "%S"
	authorName     = "%aN"
	authorEmail    = "%aE"
	authorDate     = "%at"
	committerName  = "%cN"
	committerEmail = "%cE"
	committerDate  = "%ct"
	rawBody        = "%B"
	parentHashes   = "%P"
)

var (
	partsPerCommit = len(formatWithRefs)

	formatWithRefs = []string{
		hash,
		refNames,
		sourceRefs,
		authorName,
		authorEmail,
		authorDate,
		committerName,
		committerEmail,
		committerDate,
		rawBody,
		parentHashes,
	}

	formatWithoutRefs = []string{
		hash,
		"",
		sourceRefs,
		authorName,
		authorEmail,
		authorDate,
		committerName,
		committerEmail,
		committerDate,
		rawBody,
		parentHashes,
	}

	logArgsWithRefs = []string{
		"log",
		"--decorate=full",
		"-z",
		"--no-merges",
		"--format=format:" + strings.Join(formatWithRefs, "%x00") + "%x00",
	}

	logArgsWithoutRefs = []string{
		"log",
		"--decorate=full",
		"-z",
		"--no-merges",
		"--format=format:" + strings.Join(formatWithoutRefs, "%x00") + "%x00",
	}

	sep = []byte{0x0}
)

// RawCommit is a shallow parse of the output of git log
type RawCommit struct {
	Hash           []byte
	RefNames       []byte
	SourceRefs     []byte
	AuthorName     []byte
	AuthorEmail    []byte
	AuthorDate     []byte
	CommitterName  []byte
	CommitterEmail []byte
	CommitterDate  []byte
	Message        []byte
	ParentHashes   []byte
}

func NewCommitScanner(r io.Reader) *CommitScanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1<<22)

	// Split by commit
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) == 0 { // should only happen when atEOF
			return 0, nil, nil
		}

		// See if we have enough null bytes to constitute a full commit
		if bytes.Count(data, sep) < partsPerCommit+1 {
			if atEOF {
				return 0, nil, errors.Errorf("incomplete line")
			}
			return 0, nil, nil
		}

		// If we do, expand token to the end of that commit
		for i := 0; i < partsPerCommit+1; i++ {
			idx := bytes.IndexByte(data[len(token):], 0x0)
			if idx == -1 {
				panic("we already counted enough bytes in data")
			}
			token = data[:len(token)+idx+1]
		}
		return len(token), token, nil

	})

	return &CommitScanner{
		scanner: scanner,
	}
}

type CommitScanner struct {
	scanner *bufio.Scanner
	next    *RawCommit
	err     error
}

func (c *CommitScanner) Scan() bool {
	if !c.scanner.Scan() {
		return false
	}

	// Make a copy so the view can outlive the next scan
	buf := make([]byte, len(c.scanner.Bytes()))
	copy(buf, c.scanner.Bytes())

	parts := bytes.SplitN(buf, sep, partsPerCommit+1)
	if len(parts) < partsPerCommit+1 {
		c.err = errors.Errorf("invalid commit log entry: %q", parts)
		return false
	}

	c.next = &RawCommit{
		Hash:           parts[0],
		RefNames:       parts[1],
		SourceRefs:     parts[2],
		AuthorName:     parts[3],
		AuthorEmail:    parts[4],
		AuthorDate:     parts[5],
		CommitterName:  parts[6],
		CommitterEmail: parts[7],
		CommitterDate:  parts[8],
		Message:        parts[9],
		ParentHashes:   parts[10],
	}

	return true
}

func (c *CommitScanner) NextCommitView() *RawCommit {
	return c.next
}

func (c *CommitScanner) Err() error {
	return c.err
}

const (
	batchSize  = 512
	numWorkers = 4
)

func Search(dir string, revisionArgs []string, p CommitPredicate, onMatch func(*LazyCommit, *HighlightedCommit) bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	jobs := make(chan job, 128)
	resultChans := make(chan chan searchResult, 128)

	// Start feeder
	g.Go(func() error {
		defer close(resultChans)
		defer close(jobs)

		cmd := exec.CommandContext(ctx, "git", logArgsWithoutRefs...)
		pr, pw := io.Pipe()
		cmd.Stdout = pw
		cmd.Dir = dir
		if err := cmd.Start(); err != nil {
			return err
		}

		var cmdErr error
		go func() {
			cmdErr = cmd.Wait()
			pw.Close()
		}()

		batch := make([]*RawCommit, 0, batchSize)
		sendBatch := func() {
			resultChan := make(chan searchResult, 128)
			resultChans <- resultChan
			jobs <- job{
				batch:      batch,
				resultChan: resultChan,
			}
			batch = make([]*RawCommit, 0, batchSize)
		}

		cs := NewCommitScanner(pr)
		for cs.Scan() {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			cv := cs.NextCommitView()
			batch = append(batch, cv)
			if len(batch) == batchSize {
				sendBatch()
			}
		}

		if len(batch) > 0 {
			sendBatch()
		}

		if cmdErr != nil {
			return cmdErr
		}
		return cs.Err()
	})

	// Start workers
	for i := 0; i < numWorkers; i++ {
		g.Go(func() error {
			// Create a new diff fetcher subprocess for each worker
			diffFetcher, err := StartDiffFetcher(ctx, dir)
			if err != nil {
				return err
			}

			runJob := func(j job) error {
				defer close(j.resultChan)

				if ctx.Err() != nil {
					// ignore context error, but don't spend time
					// running this job
					return nil
				}

				for _, cv := range j.batch {
					lc := &LazyCommit{
						RawCommit:   cv,
						diffFetcher: diffFetcher,
					}
					commitMatches, highlights, err := p.Match(lc)
					if err != nil {
						return err
					}
					if commitMatches {
						j.resultChan <- searchResult{
							lazyCommit:        lc,
							highlightedCommit: highlights,
						}
					}
				}
				return nil
			}

			var errors error
			for j := range jobs {
				if err := runJob(j); err != nil {
					multierror.Append(errors, err)
				}
			}
			return err
		})
	}

	// Consumer goroutine that consumes results in the order jobs were
	// submitted to the job queue
	g.Go(func() error {
	OUTER:
		for resultChan := range resultChans {
			for result := range resultChan {
				keepGoing := onMatch(result.lazyCommit, result.highlightedCommit)
				if !keepGoing {
					cancel()
					break OUTER
				}
			}
		}

		// Drain all the channels so writers never block
		for resultChan := range resultChans {
			for range resultChan {
			}
		}

		return nil
	})

	return g.Wait()
}

type job struct {
	batch      []*RawCommit
	resultChan chan searchResult
}

type searchResult struct {
	lazyCommit        *LazyCommit
	highlightedCommit *HighlightedCommit
}

type LazyCommit struct {
	*RawCommit
	diffFetcher *DiffFetcher
}

func (l *LazyCommit) AuthorDate() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.RawCommit.AuthorDate))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

func (l *LazyCommit) CommitterDate() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.RawCommit.CommitterDate))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

// RawDiff returns the diff exactly as returned by git diff-tree
func (l *LazyCommit) RawDiff() ([]byte, error) {
	return l.diffFetcher.FetchDiff(l.Hash)
}

// Diff fetches the diff, then formats it in the format used throughout our app
func (l *LazyCommit) Diff() (FormattedDiff, error) {
	// TODO lazy fetch
	rawDiff, err := l.RawDiff()
	if err != nil {
		return "", err
	}
	return FormatDiff(rawDiff), nil
}

func (l *LazyCommit) ParentIDs() []api.CommitID {
	strs := strings.Split(string(l.ParentHashes), " ")
	commitIDs := make([]api.CommitID, 0, len(strs))
	for _, str := range strs {
		commitIDs = append(commitIDs, api.CommitID(str))
	}
	return commitIDs
}

func (l *LazyCommit) RefNames() []string {
	return strings.Split(string(l.RawCommit.RefNames), ", ")
}

func (l *LazyCommit) SourceRefs() []string {
	return strings.Split(string(l.RawCommit.SourceRefs), ", ")
}
