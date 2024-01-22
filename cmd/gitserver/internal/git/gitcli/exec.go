package gitcli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// gogit implementation of list refs.
func listRefs(repoPath common.GitDir, limit, offset int) ([]gitdomain.Ref, error) {
	r, err := git.PlainOpen(repoPath.Path())
	if err != nil {
		// Failed to open the repo.
		return nil, err
	}
	it, err := r.References()
	if err != nil {
		return nil, err
	}
	defer it.Close()
	refs := make([]gitdomain.Ref, 0)
	skipped := 0
	for {
		ref, err := it.Next()
		if err != nil {
			if err == io.EOF {
				return refs, nil
			}
			return nil, err
		}
		if skipped < offset {
			skipped++
			continue
		}
		refs = append(refs, gitdomain.Ref{
			Name:     ref.Name().String(),
			CommitID: api.CommitID(ref.Name().String()),
		})
		if len(refs) == limit {
			return refs, nil
		}
	}
}

// gogit implementation of list tags.
func listTags(repoPath common.GitDir, limit, offset int) ([]*gitdomain.Tag, error) {
	r, err := git.PlainOpen(repoPath.Path())
	if err != nil {
		// Failed to open the repo.
		return nil, err
	}
	it, err := r.TagObjects()
	if err != nil {
		return nil, err
	}
	defer it.Close()
	tags := make([]*gitdomain.Tag, 0)
	skipped := 0
	for {
		tag, err := it.Next()
		if err != nil {
			if err == io.EOF {
				return tags, nil
			}
			return nil, err
		}
		if skipped < offset {
			skipped++
			continue
		}

		tags = append(tags, &gitdomain.Tag{
			Name:        tag.Name,
			CommitID:    api.CommitID(tag.Target.String()),
			CreatorDate: tag.Tagger.When,
		})
		if len(tags) == limit {
			return tags, nil
		}
	}
}

// gogit implementation of list branches.
func listBranches(repoPath common.GitDir, limit, offset int) ([]*gitdomain.Branch, error) {
	r, err := git.PlainOpen(repoPath.Path())
	if err != nil {
		// Failed to open the repo.
		return nil, err
	}
	it, err := r.Branches()
	if err != nil {
		return nil, err
	}
	defer it.Close()
	branches := make([]*gitdomain.Branch, 0)
	skipped := 0
	for {
		branch, err := it.Next()
		if err != nil {
			if err == io.EOF {
				return branches, nil
			}
			return nil, err
		}
		if skipped < offset {
			skipped++
			continue
		}

		branches = append(branches, &gitdomain.Branch{
			Name: branch.Name().Short(),
			Head: api.CommitID(branch.Target().Short()),
			// TODO: Optional fields here.
		})
		if len(branches) == limit {
			return branches, nil
		}
	}
}

// gogit implementation of ResolveRevision.
func resolveRev(repoPath common.GitDir, revspec string) (string, error) {
	r, err := git.PlainOpen(repoPath.Path())
	if err != nil {
		// Failed to open the repo.
		return "", err
	}
	rev, err := r.ResolveRevision(plumbing.Revision(revspec))
	if err != nil {
		return "", err
	}
	return rev.String(), nil
}

// gogit implementation of get file.
func readObj(repoPath common.GitDir, commitSHA string, path string) (io.ReadCloser, error) {
	r, err := git.PlainOpen(repoPath.Path())
	if err != nil {
		// Failed to open the repo.
		return nil, err
	}

	c, err := r.CommitObject(plumbing.NewHash(commitSHA))
	if err != nil {
		// If plumbing.ErrObjectNotFound, commit wasn't found.
		return nil, err
	}
	t, err := r.TreeObject(c.TreeHash)
	if err != nil {
		// If plumbing.ErrObjectNotFound, commit tree wasn't found.
		return nil, err
	}
	te, err := t.FindEntry(path)
	if err != nil {
		return nil, err
	}
	if te.Mode == filemode.Dir {
		return nil, &os.PathError{Op: "git ls-tree", Path: path, Err: os.ErrNotExist}
	}

	f, err := t.File(path)
	if err != nil {
		// TODO: Verify this fails for a submodule.
		// If so, check the TreeObject.FindEntry result for mode, if it's a submodule,
		// parse the submodule config and return the same error that lsTreeUncached returns.
		w, wErr := r.Worktree()
		if wErr != nil {
			return nil, wErr
		}
		_, sErr := w.Submodule(path)
		if sErr != nil {
			if sErr == git.ErrSubmoduleNotFound {
				// If plumbing.ErrObjectNotFound, file wasn't found in tree.
				return nil, err
			}
			return nil, sErr
		}
		// For submodules, we return io.EOF for now.
		return nil, io.EOF

	}
	return f.Reader()
}

func (g *gitCLIBackend) Exec(ctx context.Context, args ...string) (io.ReadCloser, error) {
	cmd, cancel, err := g.gitCommand(ctx, args...)
	defer cancel()
	if err != nil {
		return nil, err
	}

	return g.runGitCommand(ctx, cmd)
}

var (
	// gitCmdAllowlist are commands and arguments that are allowed to execute and are
	// checked by IsAllowedGitCmd
	gitCmdAllowlist = map[string][]string{
		"log":    append([]string{}, gitCommonAllowlist...),
		"show":   append([]string{}, gitCommonAllowlist...),
		"remote": {"-v"},
		"diff":   append([]string{}, gitCommonAllowlist...),
		"blame":  {"--root", "--incremental", "-w", "-p", "--porcelain", "--"},
		"branch": {"-r", "-a", "--contains", "--merged", "--format"},

		"rev-parse":    {"--abbrev-ref", "--symbolic-full-name", "--glob", "--exclude"},
		"rev-list":     {"--first-parent", "--max-parents", "--reverse", "--max-count", "--count", "--after", "--before", "--", "-n", "--date-order", "--skip", "--left-right"},
		"ls-remote":    {"--get-url"},
		"symbolic-ref": {"--short"},
		"archive":      {"--worktree-attributes", "--format", "-0", "HEAD", "--"},
		"ls-tree":      {"--name-only", "HEAD", "--long", "--full-name", "--object-only", "--", "-z", "-r", "-t"},
		"ls-files":     {"--with-tree", "-z"},
		"for-each-ref": {"--format", "--points-at"},
		"tag":          {"--list", "--sort", "-creatordate", "--format", "--points-at"},
		"merge-base":   {"--"},
		"show-ref":     {"--heads"},
		"shortlog":     {"-s", "-n", "-e", "--no-merges", "--after", "--before"},
		"cat-file":     {"-p"},
		"lfs":          {},

		// Commands used by GitConfigStore:
		"config": {"--get", "--unset-all"},

		// Commands used by Batch Changes when publishing changesets.
		"init":       {},
		"reset":      {"-q"},
		"commit":     {"-m"},
		"push":       {"--force"},
		"update-ref": {},
		"apply":      {"--cached", "-p0"},

		// Used in tests to simulate errors with runCommand in handleExec of gitserver.
		"testcommand": {},
		"testerror":   {},
		"testecho":    {},
		"testcat":     {},
	}

	// `git log`, `git show`, `git diff`, etc., share a large common set of allowed args.
	gitCommonAllowlist = []string{
		"--name-only", "--name-status", "--full-history", "-M", "--date", "--format", "-i", "-n", "-n1", "-m", "--", "-n200", "-n2", "--follow", "--author", "--grep", "--date-order", "--decorate", "--skip", "--max-count", "--numstat", "--pretty", "--parents", "--topo-order", "--raw", "--follow", "--all", "--before", "--no-merges", "--fixed-strings",
		"--patch", "--unified", "-S", "-G", "--pickaxe-all", "--pickaxe-regex", "--function-context", "--branches", "--source", "--src-prefix", "--dst-prefix", "--no-prefix",
		"--regexp-ignore-case", "--glob", "--cherry", "-z", "--reverse", "--ignore-submodules",
		"--until", "--since", "--author", "--committer",
		"--all-match", "--invert-grep", "--extended-regexp",
		"--no-color", "--decorate", "--no-patch", "--exclude",
		"--no-merges",
		"--no-renames",
		"--full-index",
		"--find-copies",
		"--find-renames",
		"--first-parent",
		"--no-abbrev",
		"--inter-hunk-context",
		"--after",
		"--date.order",
		"-s",
		"-100",
	}
)

var gitObjectHashRegex = regexp.MustCompile(`^[a-fA-F\d]*$`)

// common revs used with diff
var knownRevs = map[string]struct{}{
	"master":     {},
	"main":       {},
	"head":       {},
	"fetch_head": {},
	"orig_head":  {},
	"@":          {},
}

// isAllowedDiffArg checks if diff arg exists as a file. We do some preliminary checks
// as well as OS calls are more expensive. The function checks for object hashes and
// common revision names.
func isAllowedDiffArg(arg string) bool {
	// a hash is probably not a local file
	if gitObjectHashRegex.MatchString(arg) {
		return true
	}

	// check for parent and copy branch notations
	for _, c := range []string{" ", "^", "~"} {
		if _, ok := knownRevs[strings.ToLower(strings.Split(arg, c)[0])]; ok {
			return true
		}
	}
	// make sure that arg is not a local file
	_, err := os.Stat(arg)

	return os.IsNotExist(err)
}

// isAllowedGitArg checks if the arg is allowed.
func isAllowedGitArg(allowedArgs []string, arg string) bool {
	// Split the arg at the first equal sign and check the LHS against the allowlist args.
	splitArg := strings.Split(arg, "=")[0]

	// We use -- to specify the end of command options.
	// See: https://unix.stackexchange.com/a/11382/214756.
	if splitArg == "--" {
		return true
	}

	for _, allowedArg := range allowedArgs {
		if splitArg == allowedArg {
			return true
		}
	}
	return false
}

// isAllowedDiffPathArg checks if the diff path arg is allowed.
func isAllowedDiffPathArg(arg string, repoDir common.GitDir) bool {
	// allows diff command path that requires (dot) as path
	// example: diff --find-renames ... --no-prefix commit -- .
	if arg == "." {
		return true
	}

	arg = filepath.Clean(arg)
	if !filepath.IsAbs(arg) {
		arg = repoDir.Path(arg)
	}

	filePath, err := filepath.Abs(arg)
	if err != nil {
		return false
	}

	// Check if absolute path is a sub path of the repo dir
	repoRoot, err := filepath.Abs(repoDir.Path())
	if err != nil {
		return false
	}

	return strings.HasPrefix(filePath, repoRoot)
}

// IsAllowedGitCmd checks if the cmd and arguments are allowed.
//
// TODO: This should be unexported and solely be a concern of the CLI package,
// as other backends should do their own validation passes.
func IsAllowedGitCmd(logger log.Logger, args []string, dir common.GitDir) bool {
	if len(args) == 0 || len(gitCmdAllowlist) == 0 {
		return false
	}

	cmd := args[0]
	allowedArgs, ok := gitCmdAllowlist[cmd]
	if !ok {
		// Command not allowed
		logger.Warn("command not allowed", log.String("cmd", cmd))
		return false
	}

	// I hate state machines, but I hate them less than complicated multi-argument checking
	checkFileInput := false
	for i, arg := range args[1:] {
		if checkFileInput {
			if arg == "-" {
				checkFileInput = false
				continue
			}
			logger.Warn("isAllowedGitCmd: unallowed file input for `git commit`", log.String("cmd", cmd), log.String("arg", arg))
			return false
		}
		if strings.HasPrefix(arg, "-") {
			// Special-case `git log -S` and `git log -G`, which interpret any characters
			// after their 'S' or 'G' as part of the query. There is no long form of this
			// flags (such as --something=query), so if we did not special-case these, there
			// would be no way to safely express a query that began with a '-' character.
			// (Same for `git show`, where the flag has the same meaning.)
			if (cmd == "log" || cmd == "show") && (strings.HasPrefix(arg, "-S") || strings.HasPrefix(arg, "-G")) {
				continue // this arg is OK
			}

			// Special case handling of commands like `git blame -L15,60`.
			if cmd == "blame" && strings.HasPrefix(arg, "-L") {
				continue // this arg is OK
			}

			// Special case numeric arguments like `git log -20`.
			if _, err := strconv.Atoi(arg[1:]); err == nil {
				continue // this arg is OK
			}

			// For `git commit`, allow reading the commit message from stdin
			// but don't just blindly accept the `--file` or `-F` args
			// because they could be used to read arbitrary files.
			// Instead, accept only the forms that read from stdin.
			if cmd == "commit" {
				if arg == "--file=-" {
					continue
				}
				// checking `-F` requires a second check for `-` in the next argument
				// Instead of an obtuse check of next and previous arguments, set state and check it the next time around
				// Here's the alternative obtuse check of previous and next arguments:
				// (arg == "-F" && len(args) > i+2 && args[i+2] == "-") || (arg == "-" && args[i] == "-F")
				if arg == "-F" {
					checkFileInput = true
					continue
				}
			}

			if !isAllowedGitArg(allowedArgs, arg) {
				logger.Warn("IsAllowedGitCmd.isAllowedGitArgcmd", log.String("cmd", cmd), log.String("arg", arg))
				return false
			}
		}
		// diff argument may contains file path and isAllowedDiffArg and isAllowedDiffPathArg
		// helps verifying the file existence in disk
		if cmd == "diff" {
			dashIndex := slices.Index(args[1:], "--")
			if (dashIndex < 0 || i < dashIndex) && !isAllowedDiffArg(arg) {
				// verifies arguments before --
				logger.Warn("IsAllowedGitCmd.isAllowedDiffArg", log.String("cmd", cmd), log.String("arg", arg))
				return false
			} else if (i > dashIndex && dashIndex >= 0) && !isAllowedDiffPathArg(arg, dir) {
				// verifies arguments after --
				logger.Warn("IsAllowedGitCmd.isAllowedDiffPathArg", log.String("cmd", cmd), log.String("arg", arg))
				return false
			}
		}
	}
	return true
}
