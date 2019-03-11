package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.New("invalid git revision spec (begins with '-')")
	}
	return nil
}

// ExecSafe executes a Git subcommand iff it is allowed according to a whitelist.
//
// An error is only returned when there is a failure unrelated to the actual command being
// executed. If the executed command exits with a nonzero exit code, err == nil. This is similar to
// how http.Get returns a nil error for HTTP non-2xx responses.
func ExecSafe(ctx context.Context, repo gitserver.Repo, params []string) (stdout, stderr []byte, exitCode int, err error) {
	if Mocks.ExecSafe != nil {
		return Mocks.ExecSafe(params)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ExecSafe")
	defer span.Finish()

	if len(params) == 0 {
		return nil, nil, 0, errors.New("at least one argument required")
	}

	if !isWhitelistedGitCmd(params) {
		return nil, nil, 0, fmt.Errorf("command failed: %q is not a whitelisted git command", params)
	}

	cmd := gitserver.DefaultClient.Command("git", params...)
	cmd.Repo = repo
	stdout, stderr, err = cmd.DividedOutput(ctx)
	exitCode = cmd.ExitStatus
	if exitCode != 0 && err != nil {
		err = nil // the error must just indicate that the exit code was nonzero
	}
	return stdout, stderr, exitCode, err
}

// ExecReader executes an arbitrary `git` command (`git [args...]`) and returns a reader connected
// to its stdout.
func ExecReader(ctx context.Context, repo gitserver.Repo, args []string) (io.ReadCloser, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ExecReader")
	span.SetTag("args", args)
	defer span.Finish()

	if !isWhitelistedGitCmd(args) {
		return nil, fmt.Errorf("command failed: %v is not a whitelisted git command", args)
	}
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	return gitserver.StdoutReader(ctx, cmd)
}

func readUntilTimeout(ctx context.Context, cmd *gitserver.Cmd) (data []byte, complete bool, err error) {
	sr, err := gitserver.StdoutReader(ctx, cmd)
	if urlErr, ok := err.(*url.Error); ok && urlErr.Err == context.DeadlineExceeded {
		// Continue; the gitserver call exceeded our deadline before the command
		// produced any output.
	} else if err != nil {
		return nil, false, err
	}

	if sr != nil {
		defer sr.Close()
		var err error
		data, err = ioutil.ReadAll(sr)
		if err == nil {
			complete = true
		} else if err != nil && err != context.DeadlineExceeded {
			data = bytes.TrimSpace(data)
			if isBadObjectErr(string(data), "") || isInvalidRevisionRangeError(string(data), "") {
				return nil, true, &RevisionNotFoundError{Repo: cmd.Repo.Name, Spec: "UNKNOWN"}
			}
			if len(data) > 100 {
				data = append(data[:100], []byte("... (truncated)")...)
			}
			return nil, true, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, data))
		}
	}

	return data, complete, nil
}

var (
	// gitCmdWhitelist are commands and arguments that are allowed to execute when calling ExecSafe.
	gitCmdWhitelist = map[string][]string{
		"log":    append([]string{}, gitCommonWhitelist...),
		"show":   append([]string{}, gitCommonWhitelist...),
		"remote": {"-v"},
		"diff":   append([]string{}, gitCommonWhitelist...),
		"blame":  {"--root", "--incremental", "-w", "-p", "--porcelain", "--"},
		"branch": {"-r", "-a", "--contains"},

		"rev-parse":    {"--abbrev-ref", "--symbolic-full-name"},
		"rev-list":     {"--max-parents", "--reverse", "--max-count"},
		"ls-remote":    {"--get-url"},
		"symbolic-ref": {"--short"},
	}

	// `git log`, `git show`, `git diff`, etc., share a large common set of whitelisted args.
	gitCommonWhitelist = []string{
		"--name-status", "--full-history", "-M", "--date", "--format", "-i", "-n1", "-m", "--", "-n200", "-n2", "--follow", "--author", "--grep", "--date-order", "--decorate", "--skip", "--max-count", "--numstat", "--pretty", "--parents", "--topo-order", "--raw", "--follow", "--all", "--before", "--no-merges",
		"--patch", "--unified", "-S", "-G", "--pickaxe-all", "--pickaxe-regex", "--function-context", "--branches", "--source", "--src-prefix", "--dst-prefix", "--no-prefix",
		"--regexp-ignore-case", "--glob", "--cherry", "-z",
		"--until", "--since", "--author", "--committer",
		"--all-match", "--invert-grep", "--extended-regexp",
		"--no-color", "--decorate", "--no-patch", "--exclude",
		"--no-merges",
		"--full-index",
		"--find-copies",
		"--find-renames",
		"--inter-hunk-context",
	}
)

// isWhitelistedGitArg checks if the arg is whitelisted.
func isWhitelistedGitArg(whitelistedArgs []string, arg string) bool {
	// Split the arg at the first equal sign and check the LHS against the whitelist args.
	splitArg := strings.Split(arg, "=")[0]
	for _, whiteListedArg := range whitelistedArgs {
		if splitArg == whiteListedArg {
			return true
		}
	}
	return false
}

// isWhitelistedGitCmd checks if the cmd and arguments are whitelisted.
func isWhitelistedGitCmd(args []string) bool {
	// check if the supplied command is a whitelisted cmd
	if len(gitCmdWhitelist) == 0 {
		return false
	}
	cmd := args[0]
	whiteListedArgs, ok := gitCmdWhitelist[cmd]
	if !ok {
		// Command not whitelisted
		return false
	}
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			// Special-case `git log -S` and `git log -G`, which interpret any characters
			// after their 'S' or 'G' as part of the query. There is no long form of this
			// flags (such as --something=query), so if we did not special-case these,
			// there would be no way to safely express a query that began with a '-'
			// character. (Same for `git show`, where the flag has the same meaning.)
			if (cmd == "log" || cmd == "show") && (strings.HasPrefix(arg, "-S") || strings.HasPrefix(arg, "-G")) {
				continue // this arg is OK
			}

			if !isWhitelistedGitArg(whiteListedArgs, arg) {
				return false
			}
		}
	}
	return true
}

func gitserverCmdFunc(repo gitserver.Repo) cmdFunc {
	return func(args []string) cmd {
		cmd := gitserver.DefaultClient.Command("git", args...)
		cmd.Repo = gitserver.Repo(repo)
		return cmd
	}
}

// cmdFunc is a func that creates a new executable Git command.
type cmdFunc func(args []string) cmd

// cmd is an executable Git command.
type cmd interface {
	Output(context.Context) ([]byte, error)
	String() string
}

// commandRetryer executes a gitserver command first without a remote URL and
// ensured revision, then secondarily retries with a remote URL and ensured
// revision.
//
// This is such that gitserver commands invoked very often do not need to
// lookup the remote URL through repo-updater (an expensive process which
// consumes 2 code host API requests), unless the revision is actually missing
// and gitserver would want to try fetching it.
type commandRetryer struct {
	// cmd is the gitserver command to execute. It is never modified, except
	// when setting cmd.Repo.URL in the case that remoteURLFunc is called.
	cmd *gitserver.Cmd

	// remoteURLFunc is called to get the Git remote URL if it's not set in
	// repo and if it is needed. The Git remote URL is only required if the
	// gitserver doesn't already contain a clone of the repository or if the
	// commit must be fetched from the remote.
	//
	// If cmd.EnsureRevision == "", this field is ignored.
	remoteURLFunc func() (string, error)

	// exec is called when the cmd should be executed. It is expected to run
	// the gitserver command and return errors (e.g. RevisionNotFoundError),
	// which will be handled by the retryer by invoking exec again.
	//
	// For basic usage, see the implementation of DividedOutput.
	//
	// Any case involving the need to parse out missing revision errors from
	// the Git command output yourself will need to use this instead of the
	// DividedOutput helper.
	exec func() error
}

// DividedOutput is a helper which sets c.exec to a function which invokes
// c.cmd.DividedOutput and returns the result after calling c.run.
//
// It is the most basic usage of c.exec and c.run, and more complex usage
// patterns can be based on this implementation.
func (c *commandRetryer) DividedOutput(ctx context.Context) (data []byte, stderr []byte, err error) {
	c.exec = func() error {
		data, stderr, err = c.cmd.DividedOutput(ctx)
		return err
	}
	err = c.run()
	return
}

func (c *commandRetryer) run() error {
	// First, we try executing the command but without any EnsureRevision or
	// URL. The command most likely did not have either of these, but we zero
	// them just to make the code flow here more straightforward.
	cpy := *c.cmd
	cpy.EnsureRevision = ""
	cpy.Repo.URL = ""
	err := c.exec()
	if err == nil {
		// We didn't encounter any error, so gitserver did not need to fetch
		// the repository in order to fulfill the request.
		return nil
	}

	// Second, we retry the request if we can determine a URL and have a
	// revision we want to ensure exists, etc.
	tryAgain := func(err error) bool {
		haveURL := c.cmd.Repo.URL != "" || c.remoteURLFunc != nil
		if !haveURL || c.cmd.EnsureRevision == "" {
			// We don't have a repository URL or know the revision in question,
			// so we cannot retry the request.
			return false
		}
		if vcs.IsRepoNotExist(err) {
			return true // The repository doesn't exist yet, so retry after pulling.
		}
		if IsRevisionNotFound(err) {
			if c.cmd.EnsureRevision == "HEAD" {
				return false // We didn't find HEAD, so the repo is empty: nothing to do.
			}
			return true // The revision wasn't found, so we try again.
		}
		return false // All other error types (e.g. network failure).
	}
	if !tryAgain(err) {
		return err
	}

	// Determine the remote URL, if needed, then retry the command.
	if c.cmd.Repo.URL == "" {
		// We do modify c.cmd here because the caller may want to reuse this
		// information.
		c.cmd.Repo.URL, err = c.remoteURLFunc()
		if err != nil {
			return err
		}
	}
	cpy.EnsureRevision = c.cmd.EnsureRevision
	cpy.Repo.URL = c.cmd.Repo.URL
	return c.exec()
}
