package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/pkg/mv"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/pkg/vcs/util/tracer"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// RepoVCS is a local filesystem-backed implementation of the RepoVCS
// store interface.
type RepoVCS struct{}

var _ store.RepoVCS = (*RepoVCS)(nil)

func (s *RepoVCS) Open(ctx context.Context, repo string) (vcs.Repository, error) {
	dir := absolutePathForRepo(ctx, repo)
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return nil, err
	}

	r, err := vcs.Open("git", dir)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		r.Close()
	}()

	return tracer.Wrap(r, traceutil.Recorder(ctx)), nil
}

func (s *RepoVCS) Clone(ctx context.Context, repo string, bare, mirror bool, info *store.CloneInfo) error {
	name := filepath.Base(repo)
	dir := absolutePathForRepo(ctx, repo)
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return err
	}

	// Clone into a temporary dir. This allows us to rename into place +
	// in production our repo store performs better.
	cloneDir, err := ioutil.TempDir("", "sg-clone-"+name)
	if err == nil {
		log15.Debug("Cloning repo into temporary directory", "repo", repo, "tmp", cloneDir)
		defer os.RemoveAll(cloneDir)
	} else {
		cloneDir = dir
	}

	start := time.Now()
	r, err := vcs.Clone(info.VCS, info.CloneURL, cloneDir, vcs.CloneOpt{
		Bare:       bare,
		Mirror:     mirror,
		RemoteOpts: info.RemoteOpts,
	})
	if err != nil {
		return err
	}
	r.Close()

	// We cloned into a temporary directory, move into place
	if cloneDir != dir {
		log15.Debug("Moving cloned repo into repos dir", "repo", repo, "src", cloneDir, "dst", dir, "duration", time.Since(start))
		return mv.Atomic(cloneDir, dir)
	}
	return nil
}

func (s *RepoVCS) OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error) {
	dir := absolutePathForRepo(ctx, repo)
	return &localGitTransport{dir: dir}, nil
}
