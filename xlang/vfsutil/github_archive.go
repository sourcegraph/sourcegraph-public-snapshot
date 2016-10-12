package vfsutil

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/sourcegraph/ctxvfs"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

// NewGitHubRepoVFS creates a new VFS backed by a GitHub downloadable
// repository archive.
//
// If saveOnFS is set, then the zip files will be cached on the local
// file system. You need to take care that there is sufficient disk
// space.
func NewGitHubRepoVFS(repo, rev, subtree string, saveOnFS bool) (ctxvfs.FileSystem, error) {
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

	once sync.Once
	err  error // the error encountered during the fetch
	fs   vfs.FileSystem
	c    io.Closer // closes the zip file if it was read from disk

	save bool // save to the local file system for reuse
}

// fetchOrWait initiates the HTTP fetch if it has not yet
// started. Otherwise it waits for it to finish.
func (fs *GitHubRepoVFS) fetchOrWait(ctx context.Context) error {
	fs.once.Do(func() {
		fs.err = fs.fetch(ctx)
	})
	return fs.err
}

func (fs *GitHubRepoVFS) fetch(ctx context.Context) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GitRepoVFS fetch")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	url := fmt.Sprintf("https://codeload.%s/zip/%s", fs.repo, fs.rev)
	span.SetTag("repo", fs.repo)
	span.SetTag("commit", fs.rev)
	span.SetTag("save", fs.save)

	if fs.save {
		urlMu := urlMu(url)
		urlMu.Lock()
		defer urlMu.Unlock()
		span.LogEvent("urlMu acquired")
	}

	var zr *zip.ReadCloser
	fsPath := filepath.Join("/tmp/xlang-github-cache", fs.repo, fs.rev+".zip")
	if fs.save {
		zr, err = zip.OpenReader(fsPath)
		if err == nil {
			span.LogEvent("read from " + fsPath)
			fs.c = zr
		}
	}
	if !fs.save || os.IsNotExist(err) {
		// https://github.com/a/b/archive/master.zip redirects to
		// codeload.github.com, so let's just use the latter directly and
		// save a roundtrip.
		span.LogEvent("fetching " + url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("github repo archive: URL %s returned HTTP %d", url, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		span.LogEvent("fetched " + url)

		// Cache on the file system.
		if fs.save {
			if err := os.MkdirAll(filepath.Dir(fsPath), 0700); err != nil {
				return err
			}
			if err := ioutil.WriteFile(fsPath, body, 0600); err != nil {
				return err
			}
			span.LogEvent("cached to " + fsPath)
		}

		zrNoCloser, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			return err
		}
		zr = &zip.ReadCloser{Reader: *zrNoCloser}
	} else if err != nil {
		return err
	}

	// GitHub zip files have a top-level dir "{repobasename}-{sha}/",
	// so we need to remove that. The repobasename is in the canonical
	// casing, which may be different from fs.repo.
	var prefix string
	if len(zr.File) > 0 {
		name := zr.File[0].Name
		if strings.Contains(name, "/") {
			prefix = path.Dir(name)
		} else {
			prefix = name
		}
	} else {
		return errors.New("zip archive has no files")
	}
	ns := vfs.NameSpace{}
	ns.Bind("/", zipfs.New(zr, fs.repo), path.Join("/"+prefix, fs.subtree), vfs.BindReplace)
	fs.fs = ns
	return nil
}

func (fs *GitHubRepoVFS) Open(ctx context.Context, path string) (ctxvfs.ReadSeekCloser, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Open(path)
}

func (fs *GitHubRepoVFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Lstat(path)
}

func (fs *GitHubRepoVFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Stat(path)
}

func (fs *GitHubRepoVFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.ReadDir(path)
}

// Close closes the .zip file on disk. If it is buffered in memory,
// Close is a no-op.
func (fs *GitHubRepoVFS) Close() error {
	fs.once.Do(func() {}) // this synchronizes our access to fs.fs
	if fs.c != nil {
		return fs.c.Close()
	}
	return nil
}

func (fs *GitHubRepoVFS) String() string {
	return fmt.Sprintf("GitHubRepoVFS{repo: %q, rev: %q}", fs.repo, fs.rev)
}
