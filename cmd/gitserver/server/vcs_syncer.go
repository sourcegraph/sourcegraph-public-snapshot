package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/maven/coursier"
	"github.com/sourcegraph/sourcegraph/internal/vcs"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// VCSSyncer describes whether and how to sync content from a VCS remote to
// local disk.
type VCSSyncer interface {
	// Type returns the type of the syncer.
	Type() string
	// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
	// error indicates there is a problem.
	IsCloneable(ctx context.Context, remoteURL *vcs.URL) error
	// CloneCommand returns the command to be executed for cloning from remote.
	// Returning nil will prevent progress on the command from being reported to the user.
	CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error)
	// Fetch tries to fetch updates from the remote to given directory.
	Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error
	// RemoteShowCommand returns the command to be executed for showing remote.
	// Returning nil will cause functionality which depends on this command
	// (right now just picking the default branch) to pick arbitrary values.
	RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error)
}

// GitRepoSyncer is a syncer for Git repositories.
type GitRepoSyncer struct{}

func (s *GitRepoSyncer) Type() string {
	return "git"
}

// IsCloneable checks to see if the Git remote URL is cloneable.
func (s *GitRepoSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	if strings.ToLower(string(protocol.NormalizeRepo(api.RepoName(remoteURL.String())))) == "github.com/sourcegraphtest/alwayscloningtest" {
		return nil
	}
	if testGitRepoExists != nil {
		return testGitRepoExists(ctx, remoteURL)
	}

	args := []string{"ls-remote", remoteURL.String(), "HEAD"}
	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout(args))
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := runWithRemoteOpts(ctx, cmd, nil)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// CloneCommand returns the command to be executed for cloning a Git repository.
func (s *GitRepoSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error) {
	if err := os.MkdirAll(tmpPath, os.ModePerm); err != nil {
		return nil, errors.Wrapf(err, "clone failed to create tmp dir")
	}

	cmd = exec.CommandContext(ctx, "git", "init", "--bare", ".")
	cmd.Dir = tmpPath
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "clone setup failed")
	}

	cmd, _ = s.fetchCommand(ctx, remoteURL)
	cmd.Dir = tmpPath
	return cmd, nil
}

func (s *GitRepoSyncer) fetchCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, configRemoteOpts bool) {
	configRemoteOpts = true
	if customCmd := customFetchCmd(ctx, remoteURL); customCmd != nil {
		cmd = customCmd
		configRemoteOpts = false
	} else if useRefspecOverrides() {
		cmd = refspecOverridesFetchCmd(ctx, remoteURL)
	} else {
		cmd = exec.CommandContext(ctx, "git", "fetch",
			"--progress", "--prune", remoteURL.String(),
			// Normal git refs
			"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*",
			// GitHub pull requests
			"+refs/pull/*:refs/pull/*",
			// GitLab merge requests
			"+refs/merge-requests/*:refs/merge-requests/*",
			// Bitbucket pull requests
			"+refs/pull-requests/*:refs/pull-requests/*",
			// Gerrit changesets
			"+refs/changes/*:refs/changes/*",
			// Possibly deprecated refs for sourcegraph zap experiment?
			"+refs/sourcegraph/*:refs/sourcegraph/*")
	}
	return cmd, configRemoteOpts
}

// Fetch tries to fetch updates of a Git repository.
func (s *GitRepoSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	cmd, configRemoteOpts := s.fetchCommand(ctx, remoteURL)
	dir.Set(cmd)
	if output, err := runWith(ctx, cmd, configRemoteOpts, nil); err != nil {
		return errors.Wrapf(err, "failed to update with output %q", string(output))
	}
	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote of a Git repository.
func (s *GitRepoSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", remoteURL.String()), nil
}

// PerforceDepotSyncer is a syncer for Perforce depots.
type PerforceDepotSyncer struct {
	// MaxChanges indicates to only import at most n changes when possible.
	MaxChanges int
}

func (s *PerforceDepotSyncer) Type() string {
	return "perforce"
}

// decomposePerforceRemoteURL decomposes information back from a clone URL for a
// Perforce depot.
func decomposePerforceRemoteURL(remoteURL *vcs.URL) (username, password, host, depot string, err error) {
	if remoteURL.Scheme != "perforce" {
		return "", "", "", "", errors.New(`scheme is not "perforce"`)
	}
	password, _ = remoteURL.User.Password()
	return remoteURL.User.Username(), password, remoteURL.Host, remoteURL.Path, nil
}

// p4ping sends one message to the Perforce server to check connectivity.
func p4ping(ctx context.Context, host, username, password string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "p4", "ping", "-c", "1")
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
		"P4PASSWD="+password,
	)

	out, err := runWith(ctx, cmd, false, nil)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// p4trust blindly accepts fingerprint of the Perforce server.
func p4trust(ctx context.Context, host string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "p4", "trust", "-y", "-f")
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
	)

	out, err := runWith(ctx, cmd, false, nil)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// p4pingWithTrust attempts to ping the Perforce server and performs a trust operation when needed.
func p4pingWithTrust(ctx context.Context, host, username, password string) error {
	// Attempt to check connectivity, may be prompted to trust.
	err := p4ping(ctx, host, username, password)
	if err == nil {
		return nil // The ping worked, session still validate for the user
	}

	if strings.Contains(err.Error(), "To allow connection use the 'p4 trust' command.") {
		err := p4trust(ctx, host)
		if err != nil {
			return errors.Wrap(err, "trust")
		}
		return nil
	}

	// Something unexpected happened, bubble up the error
	return err
}

// IsCloneable checks to see if the Perforce remote URL is cloneable.
func (s *PerforceDepotSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	username, password, host, _, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return errors.Wrap(err, "decompose")
	}

	// FIXME: Need to find a way to determine if depot exists instead of a general ping to the Perforce server.
	return p4pingWithTrust(ctx, host, username, password)
}

// CloneCommand returns the command to be executed for cloning a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (*exec.Cmd, error) {
	username, password, host, depot, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return nil, errors.Wrap(err, "decompose")
	}

	err = p4pingWithTrust(ctx, host, username, password)
	if err != nil {
		return nil, errors.Wrap(err, "ping with trust")
	}

	// Example: git p4 clone --bare --max-changes 1000 //Sourcegraph/@all /tmp/clone-584194180/.git
	args := []string{"p4", "clone", "--bare"}
	if s.MaxChanges > 0 {
		args = append(args, "--max-changes", strconv.Itoa(s.MaxChanges))
	}
	args = append(args, depot+"@all", tmpPath)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
		"P4PASSWD="+password,
	)

	return cmd, nil
}

// Fetch tries to fetch updates of a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	username, password, host, _, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return errors.Wrap(err, "decompose")
	}

	err = p4pingWithTrust(ctx, host, username, password)
	if err != nil {
		return errors.Wrap(err, "ping with trust")
	}

	// Example: git p4 sync --max-changes 1000
	args := []string{"p4", "sync"}
	if s.MaxChanges > 0 {
		args = append(args, "--max-changes", strconv.Itoa(s.MaxChanges))
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
		"P4PASSWD="+password,
	)
	dir.Set(cmd)
	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return errors.Wrapf(err, "failed to update with output %q", string(output))
	}

	// Force update "master" to "refs/remotes/p4/master" where changes are synced into
	cmd = exec.CommandContext(ctx, "git", "branch", "-f", "master", "refs/remotes/p4/master")
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
		"P4PASSWD="+password,
	)
	dir.Set(cmd)
	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return errors.Wrapf(err, "failed to force update branch with output %q", string(output))
	}

	return nil
}

// RemoteShowCommand returns the command to be executed for showing Git remote of a Perforce depot.
func (s *PerforceDepotSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	// Remote info is encoded as in the current repository
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
}

type MavenArtifactSyncer struct {
}

var _ VCSSyncer = &MavenArtifactSyncer{}

func (s MavenArtifactSyncer) Type() string {
	return "maven"
}

func decomposeMavenURL(url *vcs.URL) (repository, groupID, artifactID string) {
	pathIdx := strings.Index(url.String(), "maven2")
	repository = url.String()[:pathIdx+len("maven2")]
	path := url.String()[pathIdx+len("maven2")+1:]
	groupID, artifactID = reposource.DecomposeMavenPath(path)
	return repository, groupID, artifactID
}

// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
// error indicates there is a problem.
func (s MavenArtifactSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	repository, groupID, artifactID := decomposeMavenURL(remoteURL)
	exists, err := coursier.Exists(ctx, repository, groupID, artifactID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return errors.New(fmt.Sprintf("Maven repo %v not found", remoteURL))
}

// CloneCommand returns the command to be executed for cloning from remote.
func (s MavenArtifactSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error) {
	repository, groupID, artifactID := decomposeMavenURL(remoteURL)
	versions, err := coursier.ListVersions(ctx, repository, groupID, artifactID)
	if err != nil {
		return nil, err
	}

	paths, err := coursier.FetchVersions(ctx, repository, groupID, artifactID, versions)
	if err != nil {
		return nil, err
	}

	initCmd := exec.CommandContext(ctx, "git", "init")
	initCmd.Dir = tmpPath
	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return nil, errors.Wrapf(err, "failed to init git repository with output %q", string(output))
	}

	return nil, s.commitJars(ctx, GitDir(tmpPath), paths, versions)
}

var versionPattern = lazyregexp.New(`refs/heads/(.+)$`)

// Fetch tries to fetch updates from the remote to given directory.
func (s MavenArtifactSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	repository, groupID, artifactID := decomposeMavenURL(remoteURL)
	coursierVersions, err := coursier.ListVersions(ctx, repository, groupID, artifactID)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "show-branch")
	dir.Set(cmd)
	gitVersionsRaw, err := runWith(ctx, cmd, false, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to git show-branch with output %q", string(gitVersionsRaw))
	}
	gitVersionsLines := strings.Split(string(gitVersionsRaw), "\n")

	gitVersions := make(map[string]struct{})
	for _, rawVersion := range gitVersionsLines {
		gitVersions[versionPattern.FindStringSubmatch(rawVersion)[1]] = struct{}{}
	}
	var filteredCoursierVersions []string
	for _, version := range coursierVersions {
		if _, exists := gitVersions[version]; !exists {
			filteredCoursierVersions = append(filteredCoursierVersions, version)
		}
	}

	paths, err := coursier.FetchVersions(ctx, repository, groupID, artifactID, filteredCoursierVersions)

	return s.commitJars(ctx, dir, paths, filteredCoursierVersions)
}

func (s MavenArtifactSyncer) commitJars(ctx context.Context, dir GitDir, paths, versions []string) error {
	for idx, path := range paths {
		cmd := exec.CommandContext(ctx, "git", "checkout", "--orphan", versions[idx])
		dir.Set(cmd)
		if output, err := runWith(ctx, cmd, false, nil); err != nil {
			return errors.Wrapf(err, "failed to git checkout with output %q", string(output))
		}

		cmd = exec.CommandContext(ctx, "git", "reset")
		dir.Set(cmd)
		if output, err := runWith(ctx, cmd, false, nil); err != nil {
			return errors.Wrapf(err, "failed to git reset with output %q", string(output))
		}

		cmd = exec.CommandContext(ctx, "git", "clean", "-d", "-f")
		dir.Set(cmd)
		if output, err := runWith(ctx, cmd, false, nil); err != nil {
			return errors.Wrapf(err, "failed to git clean with output %q", string(output))
		}

		cmd = exec.CommandContext(ctx, "unzip", path, "-d", "./")
		dir.Set(cmd)
		if output, err := runWith(ctx, cmd, false, nil); err != nil {
			return errors.Wrapf(err, "failed to unzip with output %q", string(output))
		}

		// TODO: codeintel, need some extra logic here to set up maven repo properly for indexing

		cmd = exec.CommandContext(ctx, "git", "add", "*")
		dir.Set(cmd)
		if output, err := runWith(ctx, cmd, false, nil); err != nil {
			return errors.Wrapf(err, "failed to git add with output %q", string(output))
		}

		cmd = exec.CommandContext(ctx, "git", "commit", "-m", versions[idx])
		dir.Set(cmd)
		if output, err := runWith(ctx, cmd, false, nil); err != nil {
			return errors.Wrapf(err, "failed to git commit with output %q", string(output))
		}
	}

	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote.
func (s MavenArtifactSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	// TODO: we return nil
	return nil, nil
}
