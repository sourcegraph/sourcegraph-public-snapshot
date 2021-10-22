package server

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
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
	CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error)
	// Fetch tries to fetch updates from the remote to given directory.
	Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error
	// RemoteShowCommand returns the command to be executed for showing remote.
	RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error)
}

// GitRepoSyncer is a syncer for Git repositories.
type GitRepoSyncer struct{}

func (s *GitRepoSyncer) Type() string {
	return "git"
}

// IsCloneable checks to see if the Git remote URL is cloneable.
func (s *GitRepoSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	if isAlwaysCloningTest(api.RepoName(remoteURL.String())) {
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
			err = errors.Errorf("%s (output follows)\n\n%s", err, out)
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
		return errors.Wrapf(err, "failed to update with output %q", newURLRedactor(remoteURL).redact(string(output)))
	}
	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote of a Git repository.
func (s *GitRepoSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", remoteURL.String()), nil
}

// FusionConfig allows configuration of the p4-fusion client
type FusionConfig struct {
	// Enabled: Enable the p4-fusion client for cloning and fetching repos
	Enabled bool
	// Client: The client spec tht should be used
	Client string
	// LookAhead: How many CLs in the future, at most, shall we keep downloaded by the time it is to commit them
	LookAhead int
	// NetworkThreads: The number of threads in the threadpool for running network calls. Defaults to the number of logical CPUs.
	NetworkThreads int
	// PrintBatch:  The p4 print batch size
	PrintBatch int
	// Refresh: How many times a connection should be reused before it is refreshed
	Refresh int
	// Retries: How many times a command should be retried before the process exits in a failure
	Retries int
}

// PerforceDepotSyncer is a syncer for Perforce depots.
type PerforceDepotSyncer struct {
	// MaxChanges indicates to only import at most n changes when possible.
	MaxChanges int

	// Client configures the client to use with p4 and enables use of a client spec to
	// find the list of interesting files in p4.
	Client string

	// FusionConfig contains information about the experimental p4-fusion client
	FusionConfig FusionConfig
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
			err = errors.Errorf("%s (output follows)\n\n%s", err, out)
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
			err = errors.Errorf("%s (output follows)\n\n%s", err, out)
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

func (s *PerforceDepotSyncer) p4CommandOptions() []string {
	flags := []string{}
	if s.MaxChanges > 0 {
		flags = append(flags, "--max-changes", strconv.Itoa(s.MaxChanges))
	}
	if s.Client != "" {
		flags = append(flags, "--use-client-spec")
	}
	return flags
}

func (s *PerforceDepotSyncer) p4CommandEnv(host, username, password string) []string {
	env := append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
		"P4PASSWD="+password,
	)
	if s.Client != "" {
		env = append(env, "P4CLIENT="+s.Client)
	}
	return env
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

	var cmd *exec.Cmd
	if s.FusionConfig.Enabled {
		// Example: p4-fusion --path //depot/... --user $P4USER --src clones/ --networkThreads 64 --printBatch 10 --port $P4PORT --lookAhead 2000 --retries 10 --refresh 100
		cmd = exec.CommandContext(ctx, "p4-fusion",
			"--path", depot+"...",
			"--client", s.FusionConfig.Client,
			"--user", username,
			"--src", tmpPath,
			"--networkThreads", strconv.Itoa(s.FusionConfig.NetworkThreads),
			"--printBatch", strconv.Itoa(s.FusionConfig.PrintBatch),
			"--port", host,
			"--lookAhead", strconv.Itoa(s.FusionConfig.LookAhead),
			"--retries", strconv.Itoa(s.FusionConfig.Retries),
			"--refresh", strconv.Itoa(s.FusionConfig.Refresh))
	} else {
		// Example: git p4 clone --bare --max-changes 1000 //Sourcegraph/@all /tmp/clone-584194180/.git
		args := append([]string{"p4", "clone", "--bare"}, s.p4CommandOptions()...)
		args = append(args, depot+"@all", tmpPath)
		cmd = exec.CommandContext(ctx, "git", args...)
	}
	cmd.Env = s.p4CommandEnv(host, username, password)

	return cmd, nil
}

// Fetch tries to fetch updates of a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	username, password, host, depot, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return errors.Wrap(err, "decompose")
	}

	err = p4pingWithTrust(ctx, host, username, password)
	if err != nil {
		return errors.Wrap(err, "ping with trust")
	}

	// Example: git p4 sync --max-changes 1000
	args := append([]string{"p4", "sync"}, s.p4CommandOptions()...)

	var cmd *exec.Cmd
	if s.FusionConfig.Enabled {
		// Example: p4-fusion --path //depot/... --user $P4USER --src clones/ --networkThreads 64 --printBatch 10 --port $P4PORT --lookAhead 2000 --retries 10 --refresh 100
		root, _ := filepath.Split(string(dir))
		log15.Info("Fetching", "root", root)
		cmd = exec.CommandContext(ctx, "p4-fusion",
			"--path", depot+"...",
			"--client", s.FusionConfig.Client,
			"--user", username,
			"--src", root+".git",
			"--networkThreads", strconv.Itoa(s.FusionConfig.NetworkThreads),
			"--printBatch", strconv.Itoa(s.FusionConfig.PrintBatch),
			"--port", host,
			"--lookAhead", strconv.Itoa(s.FusionConfig.LookAhead),
			"--retries", strconv.Itoa(s.FusionConfig.Retries),
			"--refresh", strconv.Itoa(s.FusionConfig.Refresh))
	} else {
		cmd = exec.CommandContext(ctx, "git", args...)
	}
	cmd.Env = s.p4CommandEnv(host, username, password)
	dir.Set(cmd)

	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return errors.Wrapf(err, "failed to update with output %q", newURLRedactor(remoteURL).redact(string(output)))
	}

	if !s.FusionConfig.Enabled {
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
	}

	return nil
}

// RemoteShowCommand returns the command to be executed for showing Git remote of a Perforce depot.
func (s *PerforceDepotSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	// Remote info is encoded as in the current repository
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
}
