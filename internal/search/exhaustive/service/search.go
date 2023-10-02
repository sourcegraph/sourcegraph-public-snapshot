package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

type NewSearcher interface {
	// NewSearch parses and minimally resolves the search query q. The
	// expectation is that this method is always fast and is deterministic, such
	// that calling this again in the future should return the same Searcher. IE
	// it can speak to the DB, but maybe not gitserver.
	//
	// userID is explicitly passed in and must match the actor for ctx. This
	// is done to prevent accidental bugs where we do a search on behalf of a
	// user as an internal user/etc.
	//
	// I expect this to be roughly equivalent to creation of a search plan in
	// our search codes job creator.
	//
	// Note: I expect things like feature flags for the user behind ctx could
	// affect what is returned. Alternatively as we release new versions of
	// Sourcegraph what is returned could change. This means we are not exactly
	// safe across repeated calls.
	NewSearch(ctx context.Context, userID int32, q string) (SearchQuery, error)
}

// SearchQuery represents a search in a way we can break up the work. The flow is
// something like:
//
//  1. RepositoryRevSpecs -> just speak to the DB to find the list of repos we need to search.
//  2. ResolveRepositoryRevSpec -> speak to gitserver to find out which commits to search.
//  3. Search -> actually do a search.
//
// This does mean that things like searching a commit in a monorepo are
// expected to run over a reasonable time frame (eg within a minute?).
//
// For example doing a diff search in an old repo may not be fast enough, but
// I'm not sure if we should design that in?
//
// We expect each step can be retried, but with the expectation it isn't
// idempotent due to backend state changing. The main purpose of breaking it
// out like this is so we can report progress, do retries, and spread out the
// work over time.
//
// Commentary on exhaustive worker jobs added in
// https://github.com/sourcegraph/sourcegraph/pull/55587:
//
//   - ExhaustiveSearchJob uses RepositoryRevSpecs to create ExhaustiveSearchRepoJob
//   - ExhaustiveSearchRepoJob uses ResolveRepositoryRevSpec to create ExhaustiveSearchRepoRevisionJob
//   - ExhaustiveSearchRepoRevisionJob uses Search
//
// In each case I imagine NewSearcher.NewSearch(query) to get hold of the
// SearchQuery. NewSearch is envisioned as being cheap to do. The only IO it
// does is maybe reading featureflags/site configuration/etc. This does mean
// it is possible for things to change over time, but this should be rare and
// will result in a well defined error. The alternative is a way to serialize
// a SearchQuery, but this makes it harder to make changes to search going
// forward for what should be rare errors.
type SearchQuery interface {
	RepositoryRevSpecs(context.Context) *iterator.Iterator[types.RepositoryRevSpecs]

	ResolveRepositoryRevSpec(context.Context, types.RepositoryRevSpecs) ([]types.RepositoryRevision, error)

	Search(context.Context, types.RepositoryRevision, CSVWriter) error
}

// CSVWriter makes it so we can avoid caring about search types and leave it
// up to the search job to decide the shape of data.
//
// Note: I expect the implementation of this to handle things like chunking up
// the CSV/etc. EG once we hit 100MB of data it can write the data out then
// start a new file. It takes care of remembering the header for the new file.
type CSVWriter interface {
	// WriteHeader should be called first and only once.
	WriteHeader(...string) error

	// WriteRow should have the same number of values as WriteHeader and can be
	// called zero or more times.
	WriteRow(...string) error
}

// NewBlobstoreCSVWriter creates a new BlobstoreCSVWriter which writes a CSV to
// the store. BlobstoreCSVWriter takes care of chunking the CSV into blobs of
// 100MiB, each with the same header row. Blobs are named {prefix}-{shard}
// except for the first blob, which is named {prefix}.
//
// Data is buffered in memory until the blob reaches the maximum allowed size,
// at which point the blob is uploaded to the store.
//
// The caller is expected to call Close() once and only once after the last call
// to WriteRow.
func NewBlobstoreCSVWriter(ctx context.Context, store uploadstore.Store, prefix string) *BlobstoreCSVWriter {

	c := &BlobstoreCSVWriter{
		maxBlobSizeBytes: 100 * 1024 * 1024,
		ctx:              ctx,
		prefix:           prefix,
		store:            store,
		// Start with "1" because we increment it before creating a new file. The second
		// shard will be called {prefix}-2.
		shard: 1,
	}

	c.startNewFile(ctx, prefix)

	return c
}

type BlobstoreCSVWriter struct {
	// ctx is the context we use for uploading blobs.
	ctx context.Context

	maxBlobSizeBytes int64

	prefix string

	w *csv.Writer

	// local buffer for the current blob.
	buf bytes.Buffer

	store uploadstore.Store

	// header keeps track of the header we write as the first row of a new file.
	header []string

	// close takes care of flushing w and closing the upload.
	close func() error

	// n is the total number of bytes we have buffered so far.
	n int64

	// shard is incremented before we create a new shard.
	shard int
}

func (c *BlobstoreCSVWriter) WriteHeader(s ...string) error {
	if c.header == nil {
		c.header = s
	}

	// Check that c.header matches s.
	if len(c.header) != len(s) {
		return errors.Errorf("header mismatch: %v != %v", c.header, s)
	}
	for i := range c.header {
		if c.header[i] != s[i] {
			return errors.Errorf("header mismatch: %v != %v", c.header, s)
		}
	}

	return c.write(s)
}

func (c *BlobstoreCSVWriter) WriteRow(s ...string) error {
	// Create new file if we've exceeded the max blob size.
	if c.n >= c.maxBlobSizeBytes {
		// Close the current upload.
		err := c.Close()
		if err != nil {
			return errors.Wrapf(err, "error closing upload")
		}

		c.shard++
		c.startNewFile(c.ctx, fmt.Sprintf("%s-%d", c.prefix, c.shard))
		err = c.WriteHeader(c.header...)
		if err != nil {
			return errors.Wrapf(err, "error writing header for new file")
		}
	}

	return c.write(s)
}

// startNewFile creates a new blob and sets up the CSV writer to write to it.
//
// The caller is expected to call c.Close() before calling startNewFile if a
// previous file was open.
func (c *BlobstoreCSVWriter) startNewFile(ctx context.Context, key string) {
	c.buf = bytes.Buffer{}
	csvWriter := csv.NewWriter(&c.buf)

	closeFn := func() error {
		csvWriter.Flush()
		// Don't upload empty files.
		if c.buf.Len() == 0 {
			return nil
		}
		_, err := c.store.Upload(ctx, key, &c.buf)
		return err
	}

	c.w = csvWriter
	c.close = closeFn
	c.n = 0
}

// write wraps Write to keep track of the number of bytes written. This is
// mainly for test purposes: The CSV writer is buffered (default 4096 bytes),
// and we don't have access to the number of bytes in the buffer. In production,
// we could just wrap the io.Pipe writer with a counter, ignore the buffer, and
// accept that size of the blobs is off by a few kilobytes.
func (c *BlobstoreCSVWriter) write(s []string) error {
	err := c.w.Write(s)
	if err != nil {
		return err
	}

	for _, field := range s {
		c.n += int64(len(field))
	}
	c.n += int64(len(s)) // len(s)-1 for the commas, +1 for the newline

	return nil
}

func (c *BlobstoreCSVWriter) Close() error {
	return c.close()
}

// NewSearcherFake is a convenient working implementation of SearchQuery which
// always will write results generated from the repoRevs. It expects a query
// string which looks like
//
//	 1@rev1 1@rev2 2@rev3
//
//	This is a space separated list of {repoid}@{revision}.
//
//	- RepositoryRevSpecs will return one RepositoryRevSpec per unique repository.
//	- ResolveRepositoryRevSpec returns the repoRevs for that repository.
//	- Search will write one result which is just the repo and revision.
func NewSearcherFake() NewSearcher {
	return newSearcherFunc(fakeNewSearch)
}

type newSearcherFunc func(context.Context, int32, string) (SearchQuery, error)

func (f newSearcherFunc) NewSearch(ctx context.Context, userID int32, q string) (SearchQuery, error) {
	return f(ctx, userID, q)
}

func fakeNewSearch(ctx context.Context, userID int32, q string) (SearchQuery, error) {
	if err := isSameUser(ctx, userID); err != nil {
		return nil, err
	}

	var repoRevs []types.RepositoryRevision
	for _, part := range strings.Fields(q) {
		var r types.RepositoryRevision
		if n, err := fmt.Sscanf(part, "%d@%s", &r.Repository, &r.Revision); n != 2 || err != nil {
			continue
		}
		r.RepositoryRevSpecs.Repository = r.Repository
		r.RepositoryRevSpecs.RevisionSpecifiers = types.RevisionSpecifiers("spec")
		repoRevs = append(repoRevs, r)
	}
	if len(repoRevs) == 0 {
		return nil, errors.Errorf("no repository revisions found in %q", q)
	}
	return searcherFake{
		userID:   userID,
		repoRevs: repoRevs,
	}, nil
}

type searcherFake struct {
	userID   int32
	repoRevs []types.RepositoryRevision
}

func (s searcherFake) RepositoryRevSpecs(ctx context.Context) *iterator.Iterator[types.RepositoryRevSpecs] {
	if err := isSameUser(ctx, s.userID); err != nil {
		iterator.New(func() ([]types.RepositoryRevSpecs, error) {
			return nil, err
		})
	}

	seen := map[types.RepositoryRevSpecs]bool{}
	var repoRevSpecs []types.RepositoryRevSpecs
	for _, r := range s.repoRevs {
		if seen[r.RepositoryRevSpecs] {
			continue
		}
		seen[r.RepositoryRevSpecs] = true
		repoRevSpecs = append(repoRevSpecs, r.RepositoryRevSpecs)
	}
	return iterator.From(repoRevSpecs)
}

func (s searcherFake) ResolveRepositoryRevSpec(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) ([]types.RepositoryRevision, error) {
	if err := isSameUser(ctx, s.userID); err != nil {
		return nil, err
	}

	var repoRevs []types.RepositoryRevision
	for _, r := range s.repoRevs {
		if r.RepositoryRevSpecs == repoRevSpec {
			repoRevs = append(repoRevs, r)
		}
	}
	return repoRevs, nil
}

func (s searcherFake) Search(ctx context.Context, r types.RepositoryRevision, w CSVWriter) error {
	if err := isSameUser(ctx, s.userID); err != nil {
		return err
	}

	if err := w.WriteHeader("repo", "revspec", "revision"); err != nil {
		return err
	}
	return w.WriteRow(strconv.Itoa(int(r.Repository)), string(r.RevisionSpecifiers), string(r.Revision))
}

func isSameUser(ctx context.Context, userID int32) error {
	if userID == 0 {
		return errors.New("exhaustive search must be done on behalf of an authenticated user")
	}
	a := actor.FromContext(ctx)
	if a == nil || a.UID != userID {
		return errors.Errorf("exhaustive search must be run as user %d", userID)
	}
	return nil
}
