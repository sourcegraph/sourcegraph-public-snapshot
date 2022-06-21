package worker

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SchemeExecutorToken = "token-executor"

// These env vars should be set for git commands. We want to make sure it never hangs on interactive input.
var gitStdEnv = []string{"GIT_TERMINAL_PROMPT=0"}

// prepareWorkspace creates and returns a temporary director in which acts the workspace
// while processing a single job. It is up to the caller to ensure that this directory is
// removed after the job has finished processing. If a repository name is supplied, then
// that repository will be cloned (through the frontend API) into the workspace.
func (h *handler) prepareWorkspace(ctx context.Context, commandRunner command.Runner, repositoryName, repositoryDirectory, commit string, fetchTags bool, shallowClone bool, sparseCheckout []string) (_ string, err error) {
	tempDir, err := makeTempDir()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	if repositoryName != "" {
		repoPath := tempDir
		if repositoryDirectory != "" {
			repoPath = filepath.Join(tempDir, repositoryDirectory)

			if !strings.HasPrefix(repoPath, tempDir) {
				return "", errors.Newf("invalid repo path %q not a subdirectory of %q", repoPath, tempDir)
			}

			if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
				return "", errors.Wrap(err, "creating repo directory")
			}
		}

		cloneURL, err := makeRelativeURL(
			h.options.ClientOptions.EndpointOptions.URL,
			h.options.GitServicePath,
			repositoryName,
		)
		if err != nil {
			return "", err
		}

		authorizationOption := fmt.Sprintf(
			"http.extraHeader=Authorization: %s %s",
			SchemeExecutorToken,
			h.options.ClientOptions.EndpointOptions.Token,
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
			commit,
		}

		appendFetchArg := func(arg string) {
			l := len(fetchCommand)
			insertPos := l - 2
			fetchCommand = append(fetchCommand[:insertPos+1], fetchCommand[insertPos:]...)
			fetchCommand[insertPos] = arg
		}

		if fetchTags {
			appendFetchArg("--tags")
		}

		if shallowClone {
			if !fetchTags {
				appendFetchArg("--no-tags")
			}
			appendFetchArg("--depth=1")
		}

		// For a sparse checkout, we want to add a blob filter so we only fetch the minimum set of files initially.
		if len(sparseCheckout) > 0 {
			appendFetchArg("--filter=blob:none")
		}

		gitCommands := []command.CommandSpec{
			{Key: "setup.git.init", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "init"}, Operation: h.operations.SetupGitInit},
			{Key: "setup.git.add-remote", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "remote", "add", "origin", cloneURL.String()}, Operation: h.operations.SetupAddRemote},
			// Disable gc, this can improve performance and should never run for executor clones.
			{Key: "setup.git.disable-gc", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "config", "--local", "gc.auto", "0"}, Operation: h.operations.SetupGitDisableGC},
			{Key: "setup.git.fetch", Env: gitStdEnv, Command: fetchCommand, Operation: h.operations.SetupGitFetch},
		}

		if len(sparseCheckout) > 0 {
			gitCommands = append(gitCommands, command.CommandSpec{
				Key:       "setup.git.sparse-checkout-config",
				Env:       gitStdEnv,
				Command:   []string{"git", "-C", repoPath, "config", "--local", "core.sparseCheckout", "1"},
				Operation: h.operations.SetupGitSparseCheckoutConfig,
			})
			gitCommands = append(gitCommands, command.CommandSpec{
				Key:       "setup.git.sparse-checkout-set",
				Env:       gitStdEnv,
				Command:   append([]string{"git", "-C", repoPath, "sparse-checkout", "set", "--no-cone", "--"}, sparseCheckout...),
				Operation: h.operations.SetupGitSparseCheckoutSet,
			})
		}

		checkoutCommand := []string{
			"git",
			"-C", repoPath,
			"checkout",
			"--progress",
			"--force",
			commit,
		}

		// Sparse checkouts need to fetch additional blobs, so we need to add
		// auth config here.
		if len(sparseCheckout) > 0 {
			checkoutCommand = []string{
				"git",
				"-C", repoPath,
				"-c", "protocol.version=2", "-c", authorizationOption, "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal",
				"checkout",
				"--progress",
				"--force",
				commit,
			}
		}

		gitCommands = append(gitCommands, command.CommandSpec{
			Key:       "setup.git.checkout",
			Env:       gitStdEnv,
			Command:   checkoutCommand,
			Operation: h.operations.SetupGitCheckout,
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
				repositoryName,
			},
			Operation: h.operations.SetupGitSetRemoteUrl,
		})

		for _, spec := range gitCommands {
			if err := commandRunner.Run(ctx, spec); err != nil {
				return "", errors.Wrap(err, fmt.Sprintf("failed %s", spec.Key))
			}
		}
	}

	// Create the scripts path.
	if err := os.MkdirAll(filepath.Join(tempDir, command.ScriptsPath), os.ModePerm); err != nil {
		return "", errors.Wrap(err, "creating script path")
	}

	return tempDir, nil
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

// makeTempDir defaults to makeTemporaryDirectory and can be replaced for testing
// with determinstic workspace/scripts directories.
var makeTempDir = makeTemporaryDirectory

func makeTemporaryDirectory() (string, error) {
	// TMPDIR is set in the dev Procfile to avoid requiring developers to explicitly
	// allow bind mounts of the host's /tmp. If this directory doesn't exist,
	// os.MkdirTemp below will fail.
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, "")
	}

	return os.MkdirTemp("", "")
}
