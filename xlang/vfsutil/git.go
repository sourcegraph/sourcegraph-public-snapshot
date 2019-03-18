package vfsutil

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/ctxvfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

type GitRepoVFS struct {
	CloneURL string // Git clone URL (e.g., "https://github.com/foo/bar")
	Rev      string // Git revision (should be absolute, 40-char for consistency, unless nondeterminism is OK)
	Subtree  string // only include this subtree

	once sync.Once
	err  error // the error encountered during the fetch
	fs   vfs.FileSystem
}

var gitCommitSHARx = regexp.MustCompile(`^[0-9a-f]{40}$`)

// fetchOrWait initiates the fetch if it has not yet
// started. Otherwise it waits for it to finish.
func (fs *GitRepoVFS) fetchOrWait(ctx context.Context) error {
	fs.once.Do(func() {
		fs.err = fs.fetch(ctx)
	})
	return fs.err
}

var gitArchiveBasePath = "/tmp/xlang-git-clone-cache"

func (fs *GitRepoVFS) fetch(ctx context.Context) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GitRepoVFS fetch")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	urlMu := urlMu(fs.CloneURL)
	urlMu.Lock()
	defer urlMu.Unlock()
	span.LogFields(otlog.String("event", "urlMu acquired"))

	h := sha256.Sum256([]byte(fs.CloneURL))
	urlHash := hex.EncodeToString(h[:])
	repoDir := filepath.Join(gitArchiveBasePath, urlHash+".git")

	// Make sure the rev arg can't be misinterpreted as a command-line
	// flag.
	if err := gitCheckArgSafety(fs.Rev); err != nil {
		return err
	}

	// Try resolving the revision immediately. If we can resolve it, no need to clone or update.
	commitID, err := gitObjectNameSHA(repoDir+"^0", fs.Rev)
	if err != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		if _, err := os.Stat(repoDir); err == nil {
			// Update the repo and hope that we fetch the rev (and can
			// resolve it afterwards).
			cmd := exec.CommandContext(ctx, "git", "remote", "update")
			cmd.Dir = repoDir
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("command %v failed: %s (output follows)\n%s", cmd.Args, err, out)
			}
		} else if os.IsNotExist(err) {
			// Clone the repo with full history (we will reuse it).
			cmd := exec.CommandContext(ctx, "git", "clone", "--bare", "--mirror", "--", fs.CloneURL, repoDir)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("command %v failed: %s (output follows)\n%s", cmd.Args, err, out)
			}
		} else if err != nil {
			return err
		}
	}

	// Resolve revision (if we didn't already do so above successfully
	// and needed to clone/update).
	if commitID == "" {
		commitID, err = gitObjectNameSHA(repoDir, fs.Rev)
		if err != nil {
			return err
		}
	}

	if !gitCommitSHARx.MatchString(commitID) {
		return fmt.Errorf("git rev %q from %s resolved to suspicious commit ID %q", fs.Rev, fs.CloneURL, commitID)
	}

	zr, err := gitZipArchive(repoDir, commitID, fs.Subtree)
	if err != nil {
		return err
	}
	fs.fs = zipfs.New(zr, "")
	return nil
}

func (fs *GitRepoVFS) Open(ctx context.Context, path string) (ctxvfs.ReadSeekCloser, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Open(path)
}

func (fs *GitRepoVFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Lstat(path)
}

func (fs *GitRepoVFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Stat(path)
}

func (fs *GitRepoVFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.ReadDir(path)
}

func (fs *GitRepoVFS) String() string {
	return fmt.Sprintf("GitRepoVFS{CloneURL: %q, Rev: %q}", fs.CloneURL, fs.Rev)
}

// gitCheckArgSafety returns a non-nil error if rev is unsafe to
// use as a command-line argument (i.e., it begins with "-" and thus
// could be confused with a command-line flag).
func gitCheckArgSafety(rev string) error {
	if strings.HasPrefix(rev, "-") {
		return fmt.Errorf("invalid Git revision (can't start with '-'): %q", rev)
	}
	return nil
}

func gitObjectNameSHA(dir, arg string) (string, error) {
	if err := gitCheckArgSafety(arg); err != nil {
		return "", err
	}
	cmd := exec.Command("git", "rev-parse", "--verify", arg)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %v failed: %s (output follows)\n%s", cmd.Args, err, out)
	}
	return string(bytes.TrimSpace(out)), nil
}

func gitZipArchive(dir, rev, subtree string) (*zip.ReadCloser, error) {
	if err := gitCheckArgSafety(rev); err != nil {
		return nil, err
	}
	if err := gitCheckArgSafety(subtree); err != nil {
		return nil, err
	}
	if subtree != "" {
		// This is how you specify that you want all of a subtree, not
		// just a specific file.
		rev += ":" + subtree
	}
	// TODO(sqs): for efficiency, can specify a subtree here and then
	// the vfs is only of a certain subtree
	cmd := exec.Command("git", "archive", "--format=zip", rev)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command %v failed: %s (stderr follows)\n%s", cmd.Args, err, buf.Bytes())
	}
	zr, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		return nil, err
	}
	return &zip.ReadCloser{Reader: *zr}, nil
}
