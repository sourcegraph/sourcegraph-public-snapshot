package workspace

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
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
) (err error) {
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

	proxyURL, cleanup, err := newGitProxyServer(options.EndpointURL, options.GitServicePath, job.RepositoryName, options.ExecutorToken)
	defer func() {
		err = errors.Append(err, cleanup())
	}()
	if err != nil {
		return errors.Wrap(err, "spawning git proxy server")
	}

	cloneURL, err := makeRelativeURL(proxyURL, job.RepositoryName)
	if err != nil {
		return err
	}

	fetchCommand := []string{
		"git",
		"-C", repoPath,
		"-c", "protocol.version=2",
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
			"-c", "protocol.version=2",
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

// newGitProxyServer creates a new HTTP proxy to the Sourcegraph instance on a random port.
// It handles authentication and additional headers required. The cleanup function
// should be called after the clone operations are done and _before_ the job is started.
// This is used so that we never have to tell git about the credentials used here.
//
// In the future, this will be used to provide different access tokens per job,
// so that we can tell _which_ job misused the token and also scope it's access
// to the particular repo in question.
func newGitProxyServer(endpointURL, gitServicePath, repositoryName, accessToken string) (string, func() error, error) {
	// Get new random free port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", func() error { return nil }, err
	}
	cleanupListener := func() error { return listener.Close() }

	upstream, err := makeRelativeURL(
		endpointURL,
		gitServicePath,
	)
	if err != nil {
		return "", cleanupListener, err
	}

	d := httputil.NewSingleHostReverseProxy(upstream).Director

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			d(req)

			req.Host = upstream.Host
			// Add authentication. We don't add this in the git clone URL directly
			// to never tell git about the clone secret.
			req.Header.Set("Authorization", fmt.Sprintf("%s %s", SchemeExecutorToken, accessToken))
			req.Header.Set("X-Sourcegraph-Actor-UID", "internal")
			req.URL.User = url.User("executor")
		},
	}

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent queries for repos other than this jobs repo.
		// This is _not_ a security measure, that should be handled by additional
		// clone tokens. This is mostly a gate to finding when we accidentally
		// would access another repo.
		if !strings.HasPrefix(r.URL.Path, "/"+repositoryName+"/") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// TODO: We might want to limit throughput here to the same level we limit
		// it _inside_ the firecracker VM.
		proxy.ServeHTTP(w, r)
	}))

	return fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port), cleanupListener, nil
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

	return urlx, nil
}
