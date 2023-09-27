pbckbge server

import (
	"context"
	"os"
	"os/exec"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/urlredbctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// gitRepoSyncer is b syncer for Git repositories.
type gitRepoSyncer struct {
	recordingCommbndFbctory *wrexec.RecordingCommbndFbctory
}

func NewGitRepoSyncer(r *wrexec.RecordingCommbndFbctory) *gitRepoSyncer {
	return &gitRepoSyncer{recordingCommbndFbctory: r}
}

func (s *gitRepoSyncer) Type() string {
	return "git"
}

// testGitRepoExists is b test fixture thbt overrides the return vblue for
// GitRepoSyncer.IsClonebble when it is set.
vbr testGitRepoExists func(ctx context.Context, remoteURL *vcs.URL) error

// IsClonebble checks to see if the Git remote URL is clonebble.
func (s *gitRepoSyncer) IsClonebble(ctx context.Context, repoNbme bpi.RepoNbme, remoteURL *vcs.URL) error {
	if isAlwbysCloningTestRemoteURL(remoteURL) {
		return nil
	}
	if testGitRepoExists != nil {
		return testGitRepoExists(ctx, remoteURL)
	}

	brgs := []string{"ls-remote", remoteURL.String(), "HEAD"}
	ctx, cbncel := context.WithTimeout(ctx, shortGitCommbndTimeout(brgs))
	defer cbncel()

	r := urlredbctor.New(remoteURL)
	cmd := exec.CommbndContext(ctx, "git", brgs...)
	out, err := runRemoteGitCommbnd(ctx, s.recordingCommbndFbctory.WrbpWithRepoNbme(ctx, log.NoOp(), repoNbme, cmd).WithRedbctorFunc(r.Redbct), true, nil)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = &common.GitCommbndError{Err: err, Output: string(out)}
		}
		return err
	}
	return nil
}

// CloneCommbnd returns the commbnd to be executed for cloning b Git repository.
func (s *gitRepoSyncer) CloneCommbnd(ctx context.Context, remoteURL *vcs.URL, tmpPbth string) (cmd *exec.Cmd, err error) {
	if err := os.MkdirAll(tmpPbth, os.ModePerm); err != nil {
		return nil, errors.Wrbpf(err, "clone fbiled to crebte tmp dir")
	}

	cmd = exec.CommbndContext(ctx, "git", "init", "--bbre", ".")
	cmd.Dir = tmpPbth
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrbpf(&common.GitCommbndError{Err: err}, "clone setup fbiled")
	}

	cmd, _ = s.fetchCommbnd(ctx, remoteURL)
	cmd.Dir = tmpPbth
	return cmd, nil
}

// Fetch tries to fetch updbtes of b Git repository.
func (s *gitRepoSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, repoNbme bpi.RepoNbme, dir common.GitDir, _ string) ([]byte, error) {
	cmd, configRemoteOpts := s.fetchCommbnd(ctx, remoteURL)
	dir.Set(cmd)
	r := urlredbctor.New(remoteURL)
	output, err := runRemoteGitCommbnd(ctx, s.recordingCommbndFbctory.WrbpWithRepoNbme(ctx, log.NoOp(), repoNbme, cmd).WithRedbctorFunc(r.Redbct), configRemoteOpts, nil)
	if err != nil {
		return nil, &common.GitCommbndError{Err: err, Output: urlredbctor.New(remoteURL).Redbct(string(output))}
	}
	return output, nil
}

// RemoteShowCommbnd returns the commbnd to be executed for showing remote of b Git repository.
func (s *gitRepoSyncer) RemoteShowCommbnd(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommbndContext(ctx, "git", "remote", "show", remoteURL.String()), nil
}

func (s *gitRepoSyncer) fetchCommbnd(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, configRemoteOpts bool) {
	configRemoteOpts = true
	if customCmd := customFetchCmd(ctx, remoteURL); customCmd != nil {
		cmd = customCmd
		configRemoteOpts = fblse
	} else if useRefspecOverrides() {
		cmd = refspecOverridesFetchCmd(ctx, remoteURL)
	} else {
		cmd = exec.CommbndContext(ctx, "git", "fetch",
			"--progress", "--prune", remoteURL.String(),
			// Normbl git refs
			"+refs/hebds/*:refs/hebds/*", "+refs/tbgs/*:refs/tbgs/*",
			// GitHub pull requests
			"+refs/pull/*:refs/pull/*",
			// GitLbb merge requests
			"+refs/merge-requests/*:refs/merge-requests/*",
			// Bitbucket pull requests
			"+refs/pull-requests/*:refs/pull-requests/*",
			// Gerrit chbngesets
			"+refs/chbnges/*:refs/chbnges/*",
			// Possibly deprecbted refs for sourcegrbph zbp experiment?
			"+refs/sourcegrbph/*:refs/sourcegrbph/*")
	}
	return cmd, configRemoteOpts
}
