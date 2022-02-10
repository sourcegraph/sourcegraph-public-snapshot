package gitdomain

import (
	"strings"

	"github.com/inconshreveable/log15"
)

var (
	// gitCmdAllowlist are commands and arguments that are allowed to execute when calling execSafe.
	gitCmdAllowlist = map[string][]string{
		"log":    append([]string{}, gitCommonAllowlist...),
		"show":   append([]string{}, gitCommonAllowlist...),
		"remote": {"-v"},
		"diff":   append([]string{}, gitCommonAllowlist...),
		"blame":  {"--root", "--incremental", "-w", "-p", "--porcelain", "--"},
		"branch": {"-r", "-a", "--contains", "--merged"},

		"rev-parse":    {"--abbrev-ref", "--symbolic-full-name"},
		"rev-list":     {"--max-parents", "--reverse", "--max-count", "--count", "--after", "--before", "--", "-n", "--date-order", "--skip", "--left-right"},
		"ls-remote":    {"--get-url"},
		"symbolic-ref": {"--short"},
		"archive":      {"--worktree-attributes", "--format", "-0", "HEAD", "--"},
		"ls-tree":      {"--name-only", "HEAD", "--long", "--full-name", "--", "-z", "-r"},
		"ls-files":     {"--with-tree", "-z"},
		"for-each-ref": {"--format"},
		"tag":          {"--list", "--sort", "-creatordate", "--format"},
		"merge-base":   {"--"},
		"show-ref":     {"--heads"},

		// Used in tests to simulate errors with runCommand in handleExec of gitserver.
		"testcommand": {},
		"testerror":   {},
	}

	// `git log`, `git show`, `git diff`, etc., share a large common set of allowed args.
	gitCommonAllowlist = []string{
		"--name-only", "--name-status", "--full-history", "-M", "--date", "--format", "-i", "-n", "-n1", "-m", "--", "-n200", "-n2", "--follow", "--author", "--grep", "--date-order", "--decorate", "--skip", "--max-count", "--numstat", "--pretty", "--parents", "--topo-order", "--raw", "--follow", "--all", "--before", "--no-merges",
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
		"--after",
		"--date.order",
		"-s",
	}
)

// isAllowedGitArg checks if the arg is allowed.
func isAllowedGitArg(allowedArgs []string, arg string) bool {
	// Split the arg at the first equal sign and check the LHS against the allowlist args.
	splitArg := strings.Split(arg, "=")[0]
	for _, allowedArg := range allowedArgs {
		if splitArg == allowedArg {
			return true
		}
	}
	return false
}

// IsAllowedGitCmd checks if the cmd and arguments are allowed.
func IsAllowedGitCmd(args []string) bool {
	if len(args) == 0 || len(gitCmdAllowlist) == 0 {
		return false
	}

	cmd := args[0]
	allowedArgs, ok := gitCmdAllowlist[cmd]
	if !ok {
		// Command not allowed
		log15.Warn("IsAllowedGitCmd: command not allowed", "cmd", cmd)
		return false
	}
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			// Special-case `git log -S` and `git log -G`, which interpret any characters
			// after their 'S' or 'G' as part of the query. There is no long form of this
			// flags (such as --something=query), so if we did not special-case these, there
			// would be no way to safely express a query that began with a '-' character.
			// (Same for `git show`, where the flag has the same meaning.)
			if (cmd == "log" || cmd == "show") && (strings.HasPrefix(arg, "-S") || strings.HasPrefix(arg, "-G")) {
				continue // this arg is OK
			}

			if !isAllowedGitArg(allowedArgs, arg) {
				log15.Warn("IsAllowedGitCmd.isAllowedGitArgcmd", "cmd", cmd, "arg", arg)
				return false
			}
		}
	}
	return true
}
