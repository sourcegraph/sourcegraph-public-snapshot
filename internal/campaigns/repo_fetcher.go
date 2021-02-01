package campaigns

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

// RepoFetcher abstracts the process of retrieving an archive for the given
// repository.
type RepoFetcher interface {
	// Checkout returns a RepoZip for the given repository that's possibly
	// unfetched. Users need to call `Fetch()` on the RepoZip before using it
	// and `Close()` once they're done using it.
	Checkout(*graphql.Repository) RepoZip
}

// repoFetcher is the concrete implementation of the RepoFetcher interface used
// outside of tests.
type repoFetcher struct {
	client     api.Client
	dir        string
	deleteZips bool

	zipsMu sync.Mutex
	zips   map[string]*repoZip
}

var _ RepoFetcher = &repoFetcher{}

func (rf *repoFetcher) zipFor(repo *graphql.Repository) *repoZip {
	rf.zipsMu.Lock()
	defer rf.zipsMu.Unlock()

	if rf.zips == nil {
		rf.zips = make(map[string]*repoZip)
	}

	path := filepath.Join(rf.dir, repo.Slug()+".zip")
	zip, ok := rf.zips[path]
	if !ok {
		zip = &repoZip{path: path, repo: repo, client: rf.client, deleteOnClose: rf.deleteZips}
		rf.zips[path] = zip
	}
	return zip
}

func (rf *repoFetcher) Checkout(repo *graphql.Repository) RepoZip {
	zip := rf.zipFor(repo)
	zip.mu.Lock()
	defer zip.mu.Unlock()

	zip.checkouts += 1
	return zip
}

// RepoZip implementations represent a downloaded repository archive.
type RepoZip interface {
	// Fetch downloads the archive if it's not on disk yet.
	Fetch(context.Context) error

	// Close must finalise the downloaded archive. If one or more temporary
	// files were created, they should be deleted here.
	Close() error

	// Path must return the path to the archive on the filesystem.
	Path() string
}

var _ RepoZip = &repoZip{}

// repoZip is the concrete implementation of the RepoZip interface used outside
// of tests.
type repoZip struct {
	mu sync.Mutex

	deleteOnClose bool
	path          string
	repo          *graphql.Repository

	client api.Client

	// uses is the number of *active* tasks that currently use the archive.
	uses int
	// checkouts is the number of tasks that *will* make use of the archive.
	checkouts int
}

func (rz *repoZip) Close() error {
	rz.mu.Lock()
	defer rz.mu.Unlock()

	rz.uses -= 1
	if rz.uses == 0 && rz.checkouts == 0 && rz.deleteOnClose {
		return os.Remove(rz.path)
	}

	return nil
}

func (rz *repoZip) Path() string {
	return rz.path
}

func (rz *repoZip) Fetch(ctx context.Context) error {
	rz.mu.Lock()
	defer rz.mu.Unlock()

	// Someone already fetched it
	if rz.uses > 0 {
		rz.uses += 1
		rz.checkouts -= 1
		return nil
	}

	exists, err := fileExists(rz.path)
	if err != nil {
		return err
	}

	if !exists {
		// Unlike the mkdirAll() calls elsewhere in this file, this is only
		// giving us a temporary place on the filesystem to keep the archive.
		// Since it's never mounted into the containers being run, we can keep
		// these directories 0700 without issue.
		if err := os.MkdirAll(filepath.Dir(rz.path), 0700); err != nil {
			return err
		}

		err = fetchRepositoryArchive(ctx, rz.client, rz.repo, rz.path)
		if err != nil {
			// If the context got cancelled, or we ran out of disk space, or ...
			// while we were downloading the file, we remove the partially
			// downloaded file.
			os.Remove(rz.path)

			return errors.Wrap(err, "fetching ZIP archive")
		}
	}

	rz.uses += 1
	rz.checkouts -= 1
	return nil
}

func fetchRepositoryArchive(ctx context.Context, client api.Client, repo *graphql.Repository, dest string) error {
	req, err := client.NewHTTPRequest(ctx, "GET", repositoryZipArchivePath(repo), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/zip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to fetch archive (HTTP %d from %s)", resp.StatusCode, req.URL.String())
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}

func repositoryZipArchivePath(repo *graphql.Repository) string {
	return path.Join("", repo.Name+"@"+repo.BaseRef(), "-", "raw")
}
