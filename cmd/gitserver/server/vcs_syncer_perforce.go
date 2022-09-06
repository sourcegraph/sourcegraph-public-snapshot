package server

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
			"--refresh", strconv.Itoa(s.FusionConfig.Refresh),
			"--maxChanges", strconv.Itoa(s.FusionConfig.MaxChanges),
			"--includeBinaries", strconv.FormatBool(s.FusionConfig.IncludeBinaries),
			"--fsyncEnable", strconv.FormatBool(s.FusionConfig.FsyncEnable),
			"--noColor", "true",
		)
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
func (s *PerforceDepotSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir, _ string) error {
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
		cmd = exec.CommandContext(ctx, "p4-fusion",
			"--path", depot+"...",
			"--client", s.FusionConfig.Client,
			"--user", username,
			"--src", root+".git",
			"--networkThreads", strconv.Itoa(s.FusionConfig.NetworkThreadsFetch),
			"--printBatch", strconv.Itoa(s.FusionConfig.PrintBatch),
			"--port", host,
			"--lookAhead", strconv.Itoa(s.FusionConfig.LookAhead),
			"--retries", strconv.Itoa(s.FusionConfig.Retries),
			"--refresh", strconv.Itoa(s.FusionConfig.Refresh),
			"--maxChanges", strconv.Itoa(s.FusionConfig.MaxChanges),
			"--includeBinaries", strconv.FormatBool(s.FusionConfig.IncludeBinaries),
			"--fsyncEnable", strconv.FormatBool(s.FusionConfig.FsyncEnable),
			"--noColor", "true",
		)
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

// decomposePerforceRemoteURL decomposes information back from a clone URL for a
// Perforce depot.
func decomposePerforceRemoteURL(remoteURL *vcs.URL) (username, password, host, depot string, err error) {
	if remoteURL.Scheme != "perforce" {
		return "", "", "", "", errors.New(`scheme is not "perforce"`)
	}
	password, _ = remoteURL.User.Password()
	return remoteURL.User.Username(), password, remoteURL.Host, remoteURL.Path, nil
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
