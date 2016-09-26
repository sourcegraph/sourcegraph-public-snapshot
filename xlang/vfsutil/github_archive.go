package vfsutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

// NewGitHubRepoVFS creates a new VFS backed by a GitHub downloadable
// repository archive.
//
// If saveOnFS is set, then the zip files will be cached on the local
// file system. You need to take care that there is sufficient disk
// space.
func NewGitHubRepoVFS(repo, rev, subtree string, saveOnFS bool) (vfs.FileSystem, error) {
	if !githubRepoRx.MatchString(repo) {
		return nil, fmt.Errorf(`invalid GitHub repo %q: must be "github.com/user/repo"`, repo)
	}
	return &GitHubRepoVFS{repo: repo, rev: rev, subtree: subtree, save: saveOnFS}, nil
}

var githubRepoRx = regexp.MustCompile(`^github\.com/[\w.-]{1,100}/[\w.-]{1,100}$`)

// GitHubRepoVFS is a VFS backed by GitHub's downloadable .zip
// archives of any repository at any commit. It fetches the .zip file
// live.
//
// This is the preferred method for fetching a repo's files because
// downloading the .zip archive file from GitHub is generally much
// faster than a Git clone, even a shallow clone.
//
// See
// https://developer.github.com/v3/repos/contents/#get-archive-link
// for more information.
//
// TODO(sqs): add configurable timeout
//
// TODO(sqs): if needed, we can make this faster by having it start to
// read the zip header before the full file is downloaded, and it can
// start seeking to available files immediately.
type GitHubRepoVFS struct {
	repo    string // format is "github.com/user/repo"
	rev     string // Git revision (should be absolute, 40-char for consistency, unless nondeterminism is OK)
	subtree string // path prefix inside of repo root

	mu  sync.Mutex
	err error // the error encountered during the fetch
	fs  vfs.FileSystem

	save bool // save to the local file system for reuse
}

// fetchOrWait initiates the HTTP fetch if it has not yet
// started. Otherwise it waits for it to finish.
func (fs *GitHubRepoVFS) fetchOrWait() (err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.err != nil {
		return fs.err
	}
	if fs.fs != nil {
		return nil
	}

	defer func() {
		fs.err = err
	}()

	url := fmt.Sprintf("https://codeload.%s/zip/%s", fs.repo, fs.rev)

	if fs.save {
		urlMu := urlMu(url)
		urlMu.Lock()
		defer urlMu.Unlock()
	}

	var body []byte
	fsPath := filepath.Join("/tmp/xlang-github-cache", fs.repo, fs.rev+".zip")
	if fs.save {
		body, err = ioutil.ReadFile(fsPath)
	}
	if !fs.save || os.IsNotExist(err) {
		// https://github.com/a/b/archive/master.zip redirects to
		// codeload.github.com, so let's just use the latter directly and
		// save a roundtrip.
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("github repo archive: URL %s returned HTTP %d", url, resp.StatusCode)
		}
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Cache on the file system.
		if fs.save {
			if err := os.MkdirAll(filepath.Dir(fsPath), 0700); err != nil {
				return err
			}
			if err := ioutil.WriteFile(fsPath, body, 0600); err != nil {
				return err
			}
		}
	} else if err != nil {
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	// GitHub zip files have a top-level dir "{repobasename}-{sha}/",
	// so we need to remove that.
	//
	// TODO(sqs): if fs.repo is not in the canonical case (upper/lowercase), will this fail?
	ns := vfs.NameSpace{}
	prefix := path.Join(path.Base(fs.repo)+"-"+fs.rev, fs.subtree)
	ns.Bind("/", zipfs.New(&zip.ReadCloser{Reader: *zr}, fs.repo), prefix, vfs.BindReplace)
	fs.fs = ns
	return nil
}

func (fs *GitHubRepoVFS) Open(path string) (vfs.ReadSeekCloser, error) {
	if err := fs.fetchOrWait(); err != nil {
		return nil, err
	}
	return fs.fs.Open(path)
}

func (fs *GitHubRepoVFS) Lstat(path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(); err != nil {
		return nil, err
	}
	return fs.fs.Lstat(path)
}

func (fs *GitHubRepoVFS) Stat(path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(); err != nil {
		return nil, err
	}
	return fs.fs.Stat(path)
}

func (fs *GitHubRepoVFS) ReadDir(path string) ([]os.FileInfo, error) {
	if err := fs.fetchOrWait(); err != nil {
		return nil, err
	}
	return fs.fs.ReadDir(path)
}

func (fs *GitHubRepoVFS) String() string {
	return fmt.Sprintf("GitHubRepoVFS{repo: %q, rev: %q}", fs.repo, fs.rev)
}
