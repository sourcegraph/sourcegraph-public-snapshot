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
)

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.New("invalid git revision spec (begins with '-')")
	}
	return nil
}

func GitCmdRaw(ctx context.Context, repo gitserver.Repo, params []string) (string, error) {
	if Mocks.GitCmdRaw != nil {
		return Mocks.GitCmdRaw(params)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ExtensionGitCmd")
	defer span.Finish()

	if len(params) == 0 {
		return "", errors.New("at least one argument required")
	}

	if !isWhitelistedGitCmd(params) {
		return "", fmt.Errorf("command failed: %s is not a whitelisted git command", strings.Join(params, ""))
	}

	cmd := gitserver.DefaultClient.Command("git", params...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git raw command %q failed (output: %q)", params, out))
	}

	return string(out), nil
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
	// gitCmdWhitelist are commands and arguments that are allowed to execute when calling GitCmdRaw.
	gitCmdWhitelist = map[string][]string{
		"log":    append([]string{}, gitCommonWhitelist...),
		"show":   append([]string{}, gitCommonWhitelist...),
		"remote": []string{"-v"},
		"diff":   append([]string{}, gitCommonWhitelist...),
		"blame":  []string{"--root", "--incremental", "--"},
		"branch": []string{"-r", "-a", "--contains"},

		"rev-parse":    []string{"--abbrev-ref", "--symbolic-full-name"},
		"rev-list":     []string{"--max-parents", "--reverse", "--max-count"},
		"ls-remote":    []string{"--get-url"},
		"symbolic-ref": []string{"--short"},
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
