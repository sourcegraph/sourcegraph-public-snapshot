package vcssyncer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/perforce"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/urlredactor"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PerforceDepotSyncer is a syncer for Perforce depots.
type PerforceDepotSyncer struct {
	// MaxChanges indicates to only import at most n changes when possible.
	MaxChanges int

	// Client configures the client to use with p4 and enables use of a client spec
	// to find the list of interesting files in p4.
	Client string

	// FusionConfig contains information about the experimental p4-fusion client.
	FusionConfig FusionConfig

	// P4Home is a directory we will pass to `git p4` commands as the
	// $HOME directory as it requires this to write cache data.
	P4Home string
}

func NewPerforceDepotSyncer(connection *schema.PerforceConnection, p4Home string) VCSSyncer {
	return &PerforceDepotSyncer{
		MaxChanges:   int(connection.MaxChanges),
		Client:       connection.P4Client,
		FusionConfig: configureFusionClient(connection),
		P4Home:       p4Home,
	}
}

func (s *PerforceDepotSyncer) Type() string {
	return "perforce"
}

// IsCloneable checks to see if the Perforce remote URL is cloneable.
func (s *PerforceDepotSyncer) IsCloneable(ctx context.Context, _ api.RepoName, remoteURL *vcs.URL) error {
	username, password, host, path, err := perforce.DecomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return errors.Wrap(err, "decompose")
	}

	return perforce.IsDepotPathCloneable(ctx, s.P4Home, host, username, password, path)
}

// CloneCommand returns the command to be executed for cloning a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (*exec.Cmd, error) {
	username, password, p4port, depot, err := perforce.DecomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return nil, errors.Wrap(err, "decompose")
	}

	err = perforce.P4TestWithTrust(ctx, s.P4Home, p4port, username, password)
	if err != nil {
		return nil, errors.Wrap(err, "test with trust")
	}

	var cmd *exec.Cmd
	if s.FusionConfig.Enabled {
		cmd = s.buildP4FusionCmd(ctx, depot, username, tmpPath, p4port)
	} else {
		// Example: git p4 clone --bare --max-changes 1000 //Sourcegraph/@all /tmp/clone-584194180/.git
		args := append([]string{"p4", "clone", "--bare"}, s.p4CommandOptions()...)
		args = append(args, depot+"@all", tmpPath)
		cmd = exec.CommandContext(ctx, "git", args...)
	}
	cmd.Env = s.p4CommandEnv(p4port, username, password)

	return cmd, nil
}

func (s *PerforceDepotSyncer) buildP4FusionCmd(ctx context.Context, depot, username, src, port string) *exec.Cmd {
	// Example: p4-fusion --path //depot/... --user $P4USER --src clones/ --networkThreads 64 --printBatch 10 --port $P4PORT --lookAhead 2000 --retries 10 --refresh 100 --noColor true
	return exec.CommandContext(ctx, "p4-fusion",
		"--path", depot+"...",
		"--client", s.FusionConfig.Client,
		"--user", username,
		"--src", src,
		"--networkThreads", strconv.Itoa(s.FusionConfig.NetworkThreads),
		"--printBatch", strconv.Itoa(s.FusionConfig.PrintBatch),
		"--port", port,
		"--lookAhead", strconv.Itoa(s.FusionConfig.LookAhead),
		"--retries", strconv.Itoa(s.FusionConfig.Retries),
		"--refresh", strconv.Itoa(s.FusionConfig.Refresh),
		"--maxChanges", strconv.Itoa(s.FusionConfig.MaxChanges),
		"--includeBinaries", strconv.FormatBool(s.FusionConfig.IncludeBinaries),
		"--fsyncEnable", strconv.FormatBool(s.FusionConfig.FsyncEnable),
		"--noColor", "true",
		// We don't want an empty commit for a sane merge base across branches,
		// since we don't use them and the empty commit breaks changelist parsing.
		"--noBaseCommit", "true",
	)
}

// Fetch tries to fetch updates of a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, _ api.RepoName, dir common.GitDir, _ string) ([]byte, error) {
	username, password, host, depot, err := perforce.DecomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return nil, errors.Wrap(err, "decompose")
	}

	err = perforce.P4TestWithTrust(ctx, s.P4Home, host, username, password)
	if err != nil {
		return nil, errors.Wrap(err, "test with trust")
	}

	var cmd *wrexec.Cmd
	if s.FusionConfig.Enabled {
		// Example: p4-fusion --path //depot/... --user $P4USER --src clones/ --networkThreads 64 --printBatch 10 --port $P4PORT --lookAhead 2000 --retries 10 --refresh 100
		root, _ := filepath.Split(string(dir))
		cmd = wrexec.Wrap(ctx, nil, s.buildP4FusionCmd(ctx, depot, username, root+".git", host))
	} else {
		// Example: git p4 sync --max-changes 1000
		args := append([]string{"p4", "sync"}, s.p4CommandOptions()...)
		cmd = wrexec.CommandContext(ctx, nil, "git", args...)
	}
	cmd.Env = s.p4CommandEnv(host, username, password)
	dir.Set(cmd.Cmd)

	// TODO(keegancsmith)(indradhanush) This is running a remote command and
	// we have runRemoteGitCommand which sets TLS settings/etc. Do we need
	// something for p4?
	output, err := executil.RunCommandCombinedOutput(ctx, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update with output %q", urlredactor.New(remoteURL).Redact(string(output)))
	}

	if !s.FusionConfig.Enabled {
		// Force update "master" to "refs/remotes/p4/master" where changes are synced into
		cmd = wrexec.CommandContext(ctx, nil, "git", "branch", "-f", "master", "refs/remotes/p4/master")
		cmd.Cmd.Env = append(os.Environ(),
			"P4PORT="+host,
			"P4USER="+username,
			"P4PASSWD="+password,
			"HOME="+s.P4Home,
		)
		dir.Set(cmd.Cmd)
		if output, err := executil.RunCommandCombinedOutput(ctx, cmd); err != nil {
			return nil, errors.Wrapf(err, "failed to force update branch with output %q", string(output))
		}
	}

	return output, nil
}

// RemoteShowCommand returns the command to be executed for showing Git remote of a Perforce depot.
func (s *PerforceDepotSyncer) RemoteShowCommand(ctx context.Context, _ *vcs.URL) (cmd *exec.Cmd, err error) {
	// Remote info is encoded as in the current repository
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
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

func (s *PerforceDepotSyncer) p4CommandEnv(port, username, password string) []string {
	env := append(os.Environ(),
		"P4PORT="+port,
		"P4USER="+username,
		"P4PASSWD="+password,
	)

	if s.Client != "" {
		env = append(env, "P4CLIENT="+s.Client)
	}

	if s.P4Home != "" {
		// git p4 commands write to $HOME/.gitp4-usercache.txt, we should pass in a
		// directory under our control and ensure that it is writeable.
		env = append(env, "HOME="+s.P4Home)
	}

	return env
}

// FusionConfig allows configuration of the p4-fusion client
type FusionConfig struct {
	// Enabled: Enable the p4-fusion client for cloning and fetching repos
	Enabled bool
	// Client: The client spec tht should be used
	Client string
	// LookAhead: How many CLs in the future, at most, shall we keep downloaded by
	// the time it is to commit them
	LookAhead int
	// NetworkThreads: The number of threads in the threadpool for running network
	// calls. Defaults to the number of logical CPUs.
	NetworkThreads int
	// NetworkThreadsFetch: The same as network threads but specifically used when
	// fetching rather than cloning.
	NetworkThreadsFetch int
	// PrintBatch:  The p4 print batch size
	PrintBatch int
	// Refresh: How many times a connection should be reused before it is refreshed
	Refresh int
	// Retries: How many times a command should be retried before the process exits
	// in a failure
	Retries int
	// MaxChanges limits how many changes to fetch during the initial clone. A
	// default of -1 means fetch all changes
	MaxChanges int
	// IncludeBinaries sets whether to include binary files
	IncludeBinaries bool
	// FsyncEnable enables fsync() while writing objects to disk to ensure they get
	// written to permanent storage immediately instead of being cached. This is to
	// mitigate data loss in events of hardware failure.
	FsyncEnable bool
}

func configureFusionClient(conn *schema.PerforceConnection) FusionConfig {
	// Set up default settings first
	fc := FusionConfig{
		Enabled:             false,
		Client:              conn.P4Client,
		LookAhead:           2000,
		NetworkThreads:      12,
		NetworkThreadsFetch: 12,
		PrintBatch:          100,
		Refresh:             1000,
		Retries:             10,
		MaxChanges:          -1,
		IncludeBinaries:     false,
		FsyncEnable:         false,
	}

	if conn.FusionClient == nil {
		return fc
	}

	// Required
	fc.Enabled = conn.FusionClient.Enabled

	// Optional
	if conn.FusionClient.LookAhead > 0 {
		fc.LookAhead = conn.FusionClient.LookAhead
	}
	if conn.FusionClient.NetworkThreads > 0 {
		fc.NetworkThreads = conn.FusionClient.NetworkThreads
	}
	if conn.FusionClient.NetworkThreadsFetch > 0 {
		fc.NetworkThreadsFetch = conn.FusionClient.NetworkThreadsFetch
	}
	if conn.FusionClient.PrintBatch > 0 {
		fc.PrintBatch = conn.FusionClient.PrintBatch
	}
	if conn.FusionClient.Refresh > 0 {
		fc.Refresh = conn.FusionClient.Refresh
	}
	if conn.FusionClient.Retries > 0 {
		fc.Retries = conn.FusionClient.Retries
	}
	if conn.FusionClient.MaxChanges > 0 {
		fc.MaxChanges = conn.FusionClient.MaxChanges
	}
	fc.IncludeBinaries = conn.FusionClient.IncludeBinaries
	fc.FsyncEnable = conn.FusionClient.FsyncEnable

	return fc
}
