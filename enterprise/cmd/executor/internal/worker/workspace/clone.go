package workspace

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SchemeExecutorToken = "token-executor"

// These env vars should be set for git commands. We want to make sure it never hangs on interactive input.
var gitStdEnv = []string{"GIT_TERMINAL_PROMPT=0"}

func cloneRepo(
	ctx context.Context,
	workspaceDir string,
	job executor.Job,
	commandRunner command.Runner,
	options CloneOptions,
	operations *command.Operations,
) error {
	repoPath := workspaceDir
	if job.RepositoryDirectory != "" {
		repoPath = filepath.Join(workspaceDir, job.RepositoryDirectory)

		if !strings.HasPrefix(repoPath, workspaceDir) {
			return errors.Newf("invalid repo path %q not a subdirectory of %q", repoPath, workspaceDir)
		}

		if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
			return errors.Wrap(err, "creating repo directory")
		}
	}

	cloneURL, err := makeRelativeURL(
		options.EndpointURL,
		options.GitServicePath,
		job.RepositoryName,
	)
	if err != nil {
		return err
	}

	authorizationOption := fmt.Sprintf(
		"http.extraHeader=Authorization: %s %s",
		SchemeExecutorToken,
		options.ExecutorToken,
	)

	fetchCommand := []string{
		"git",
		"-C", repoPath,
		"-c", "protocol.version=2",
		"-c", authorizationOption,
		"-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal",
		"fetch",
		"--progress",
		"--no-recurse-submodules",
		"origin",
		job.Commit,
	}

	appendFetchArg := func(arg string) {
		l := len(fetchCommand)
		insertPos := l - 2
		fetchCommand = append(fetchCommand[:insertPos+1], fetchCommand[insertPos:]...)
		fetchCommand[insertPos] = arg
	}

	if job.FetchTags {
		appendFetchArg("--tags")
	}

	if job.ShallowClone {
		if !job.FetchTags {
			appendFetchArg("--no-tags")
		}
		appendFetchArg("--depth=1")
	}

	// For a sparse checkout, we want to add a blob filter so we only fetch the minimum set of files initially.
	if len(job.SparseCheckout) > 0 {
		appendFetchArg("--filter=blob:none")
	}

	gitCommands := []command.CommandSpec{
		{Key: "setup.git.init", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "init"}, Operation: operations.SetupGitInit},
		{Key: "setup.git.add-remote", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "remote", "add", "origin", cloneURL.String()}, Operation: operations.SetupAddRemote},
		// Disable gc, this can improve performance and should never run for executor clones.
		{Key: "setup.git.disable-gc", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "config", "--local", "gc.auto", "0"}, Operation: operations.SetupGitDisableGC},
		{Key: "setup.git.fetch", Env: gitStdEnv, Command: fetchCommand, Operation: operations.SetupGitFetch},
	}

	if len(job.SparseCheckout) > 0 {
		gitCommands = append(gitCommands, command.CommandSpec{
			Key:       "setup.git.sparse-checkout-config",
			Env:       gitStdEnv,
			Command:   []string{"git", "-C", repoPath, "config", "--local", "core.sparseCheckout", "1"},
			Operation: operations.SetupGitSparseCheckoutConfig,
		})
		gitCommands = append(gitCommands, command.CommandSpec{
			Key:       "setup.git.sparse-checkout-set",
			Env:       gitStdEnv,
			Command:   append([]string{"git", "-C", repoPath, "sparse-checkout", "set", "--no-cone", "--"}, job.SparseCheckout...),
			Operation: operations.SetupGitSparseCheckoutSet,
		})
	}

	checkoutCommand := []string{
		"git",
		"-C", repoPath,
		"checkout",
		"--progress",
		"--force",
		job.Commit,
	}

	// Sparse checkouts need to fetch additional blobs, so we need to add
	// auth config here.
	if len(job.SparseCheckout) > 0 {
		checkoutCommand = []string{
			"git",
			"-C", repoPath,
			"-c", "protocol.version=2", "-c", authorizationOption, "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal",
			"checkout",
			"--progress",
			"--force",
			job.Commit,
		}
	}

	gitCommands = append(gitCommands, command.CommandSpec{
		Key:       "setup.git.checkout",
		Env:       gitStdEnv,
		Command:   checkoutCommand,
		Operation: operations.SetupGitCheckout,
	})

	// This is for LSIF, it relies on the origin being set to the upstream repo
	// for indexing.
	gitCommands = append(gitCommands, command.CommandSpec{
		Key: "setup.git.set-remote",
		Env: gitStdEnv,
		Command: []string{
			"git",
			"-C", repoPath,
			"remote",
			"set-url",
			"origin",
			job.RepositoryName,
		},
		Operation: operations.SetupGitSetRemoteUrl,
	})

	for _, spec := range gitCommands {
		if err := commandRunner.Run(ctx, spec); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed %s", spec.Key))
		}
	}

	return nil
}

func makeRelativeURL(base string, path ...string) (*url.URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	urlx, err := baseURL.ResolveReference(&url.URL{Path: filepath.Join(path...)}), nil
	if err != nil {
		return nil, err
	}

	urlx.User = url.User("executor")
	return urlx, nil
}
