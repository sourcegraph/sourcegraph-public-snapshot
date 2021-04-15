package batches

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

// RepoFetcher abstracts the process of retrieving an archive for the given
// repository.
type RepoFetcher interface {
	// Checkout returns a RepoZip for the given repository and the given
	// relative path in the repository. The RepoZip is possibly unfetched.
	// Users need to call `Fetch()` on the RepoZip before using it and
	// `Close()` once // they're done using it.
	Checkout(repo *graphql.Repository, path string) RepoZip
}

func NewRepoFetcher(client api.Client, dir string, deleteZips bool) RepoFetcher {
	return &repoFetcher{client: client, dir: dir, deleteZips: deleteZips}
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

// AdditionalWorkspaceFiles is a list of files the RepoFetcher *tries* to fetch
// when the desired archive is subdirectory in the given repository. It makes
// sense to also fetch these files, even if the steps are executed in a
// subdirectory, since they might influence some global state, such as
// `.gitignore`.
//
// If the file is not found, that's not an error.
var AdditionalWorkspaceFiles = []string{
	".gitignore",
	".gitattributes",
}

func (rf *repoFetcher) zipFor(repo *graphql.Repository, workspacePath string) *repoZip {
	rf.zipsMu.Lock()
	defer rf.zipsMu.Unlock()

	if rf.zips == nil {
		rf.zips = make(map[string]*repoZip)
	}

	slug := repo.SlugForPath(workspacePath)

	zipPath := filepath.Join(rf.dir, slug+".zip")
	zip, ok := rf.zips[zipPath]
	if !ok {
		zip = &repoZip{
			zipPath:       zipPath,
			repo:          repo,
			client:        rf.client,
			deleteOnClose: rf.deleteZips,
			pathInRepo:    workspacePath,
		}

		if workspacePath != "" {
			// We're doing another loop here to catch all
			// AdditionalWorkspaceFiles on the way *up* from the workspace to the
			// root.
			//
			// Example: path = /examples/cool/project3
			//
			// Then we want to fetch the following files:
			//
			// /.gitignore
			// /.gitattributes
			// /examples/.gitignore
			// /examples/.gitattributes
			// /examples/cool/.gitignore
			// /examples/cool/.gitattributes

			// Split on '/' because the path comes from Sourcegraph and always
			// has a "/".
			pathComponents := strings.Split(workspacePath, "/")

			var currentPath string
			for _, component := range pathComponents {
				for _, name := range []string{".gitignore", ".gitattributes"} {
					filename := path.Join(currentPath, name)
					localPath := filepath.Join(rf.dir, repo.SlugForPath(workspacePath+filename))

					zip.additionalFiles = append(zip.additionalFiles, &additionalFile{
						filename:  filename,
						localPath: localPath,
						fetched:   false,
					})
				}

				currentPath = path.Join(currentPath, component)
			}
		}

		rf.zips[zipPath] = zip
	}
	return zip
}

func (rf *repoFetcher) Checkout(repo *graphql.Repository, path string) RepoZip {
	zip := rf.zipFor(repo, path)
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

	// AdditionalFilePaths returns a map of filenames that should be put into
	// the workspace's root. The value of each entry in the map is the location
	// on the local filesystem. WorkspaceCreators need to copy the files into
	// the workspaces.
	AdditionalFilePaths() map[string]string
}

var _ RepoZip = &repoZip{}

// repoZip is the concrete implementation of the RepoZip interface used outside
// of tests.
type repoZip struct {
	mu sync.Mutex

	deleteOnClose bool

	repo       *graphql.Repository
	pathInRepo string

	client api.Client

	// zipPath is the path of the downloaded ZIP archive on the local filesystem.
	zipPath string
	// additionalFiles can contain a list of additional files that need to be copied
	// into the unzipped archive before using it as a workspace.
	additionalFiles []*additionalFile

	// uses is the number of *active* tasks that currently use the archive.
	uses int
	// checkouts is the number of tasks that *will* make use of the archive.
	checkouts int
}

type additionalFile struct {
	filename  string
	localPath string

	fetched bool
}

func (rz *repoZip) Close() error {
	rz.mu.Lock()
	defer rz.mu.Unlock()

	rz.uses -= 1
	if rz.uses == 0 && rz.checkouts == 0 && rz.deleteOnClose {
		for _, addFile := range rz.additionalFiles {
			if addFile.fetched {
				if err := os.Remove(addFile.localPath); err != nil {
					return err
				}
			}
		}
		return os.Remove(rz.zipPath)
	}

	return nil
}

func (rz *repoZip) Path() string {
	return rz.zipPath
}

func (rz *repoZip) AdditionalFilePaths() map[string]string {
	paths := map[string]string{}
	for _, f := range rz.additionalFiles {
		if f.fetched {
			paths[f.filename] = f.localPath
		}
	}
	return paths
}

func (rz *repoZip) Fetch(ctx context.Context) (err error) {
	rz.mu.Lock()
	defer rz.mu.Unlock()

	// Someone already fetched it
	if rz.uses > 0 {
		rz.uses += 1
		rz.checkouts -= 1
		return nil
	}

	if err := rz.fetchArchiveAndFiles(ctx); err != nil {
		return err
	}

	rz.uses += 1
	rz.checkouts -= 1
	return nil
}

func (rz *repoZip) fetchArchiveAndFiles(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			// If the context got cancelled, or we ran out of disk space, or ...
			// while we were downloading the file, we remove the partially
			// downloaded file.
			os.Remove(rz.zipPath)

			for _, addFile := range rz.additionalFiles {
				os.Remove(addFile.localPath)
			}
		}
	}()

	exists, err := fileExists(rz.zipPath)
	if err != nil {
		return err
	}

	if !exists {
		// Unlike the mkdirAll() calls elsewhere in this file, this is only
		// giving us a temporary place on the filesystem to keep the archive.
		// Since it's never mounted into the containers being run, we can keep
		// these directories 0700 without issue.
		if err := os.MkdirAll(filepath.Dir(rz.zipPath), 0700); err != nil {
			return err
		}

		ok, err := fetchRepositoryFile(ctx, rz.client, rz.repo, rz.pathInRepo, rz.zipPath)
		if err != nil {
			return errors.Wrap(err, "fetching ZIP archive")
		}
		if !ok {
			return errors.New("failed to download repository archive: not found")
		}
	}

	for _, addFile := range rz.additionalFiles {
		exists, err := fileExists(addFile.localPath)
		if err != nil {
			return err
		}

		if exists {
			addFile.fetched = true
			continue
		}

		ok, err := fetchRepositoryFile(ctx, rz.client, rz.repo, addFile.filename, addFile.localPath)
		if err != nil {
			return errors.Wrapf(err, "fetching %s for repository archive", addFile.filename)
		}
		// We don't return an error here, because downloading the additional
		// files is best effort. If they don't exist we skip them.
		addFile.fetched = ok
	}

	return nil
}

// fetchRepositoryInFile fetches the given `pathInRepo` using the Sourcegraph's
// raw endpoint and writes it to `dest`.
// If `pathInRepo` is empty and `dest` ends in `.zip` a ZIP archive of the
// whole repository is downloaded.
func fetchRepositoryFile(ctx context.Context, client api.Client, repo *graphql.Repository, pathInRepo string, dest string) (bool, error) {
	endpoint := repositoryRawFileEndpoint(repo, pathInRepo)
	req, err := client.NewHTTPRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	if strings.HasSuffix(dest, ".zip") {
		req.Header.Set("Accept", "application/zip")
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("unable to fetch archive (HTTP %d from %s)", resp.StatusCode, req.URL.String())
	}

	f, err := os.Create(dest)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return false, err
	}

	return true, nil
}

func repositoryRawFileEndpoint(repo *graphql.Repository, pathInRepo string) string {
	p := path.Join(repo.Name+"@"+repo.BaseRef(), "-", "raw")
	if pathInRepo != "" {
		p = path.Join(p, pathInRepo)
	}
	return p
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
