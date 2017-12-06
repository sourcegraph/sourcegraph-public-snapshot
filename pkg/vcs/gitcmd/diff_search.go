package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	validRawLogDiffSearchFormatArgs = [][]string{
		{"--no-merges", "-z", "--decorate=full", "--patch", logFormatFlag},
		{"--no-merges", "-z", "--decorate=full", logFormatFlag},
	}
)

func isValidRawLogDiffSearchFormatArgs(formatArgs []string) bool {
	for _, validArgs := range validRawLogDiffSearchFormatArgs {
		if reflect.DeepEqual(formatArgs, validArgs) {
			return true
		}
	}
	return false
}

// RawLogDiffSearch implements vcs.Repository.
func (r *Repository) RawLogDiffSearch(ctx context.Context, opt vcs.RawLogDiffSearchOptions) (results []*vcs.LogCommitSearchResult, complete bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: RawLogDiffSearch")
	span.SetTag("Opt", opt)
	defer span.Finish()

	if opt.FormatArgs == nil {
		if opt.Diff {
			opt.FormatArgs = validRawLogDiffSearchFormatArgs[0] // with --patch
		} else {
			opt.FormatArgs = validRawLogDiffSearchFormatArgs[1] // without --patch
		}
	}
	if opt.FormatArgs != nil && !isValidRawLogDiffSearchFormatArgs(opt.FormatArgs) {
		return nil, false, fmt.Errorf("invalid FormatArgs: %q", opt.FormatArgs)
	}
	for _, arg := range opt.Args {
		if arg == "--" {
			return nil, false, fmt.Errorf("invalid Args (must not contain \"--\" element): %q", opt.Args)
		}
	}

	if opt.Query.IsCaseSensitive != opt.Paths.IsCaseSensitive {
		// These options can't be set separately in `git log`, so fail.
		return nil, false, fmt.Errorf("invalid options: Query.IsCaseSensitive != Paths.IsCaseSensitive")
	}

	appendCommonQueryArgs := func(args *[]string) {
		if opt.Query.Pattern != "" {
			var queryArg string
			if opt.MatchChangedOccurrenceCount {
				queryArg = "-S"
			} else {
				queryArg = "-G"
			}
			*args = append(*args, queryArg+opt.Query.Pattern)
			if !opt.Query.IsCaseSensitive {
				*args = append(*args, "--regexp-ignore-case")
			}
			if opt.Query.IsRegExp {
				*args = append(*args, "--pickaxe-regex")
			}
		}
		if opt.Paths.IsRegExp {
			*args = append(*args, "--extended-regexp")
		}
	}

	args := []string{"log"}
	args = append(args, opt.Args...)
	if !isWhitelistedGitCmd(args) {
		return nil, false, fmt.Errorf("command failed: %q is not a whitelisted git command", args)
	}

	appendCommonDashDashArgs := func(args *[]string) {
		*args = append(*args, "--")
		// Args we append after this don't need to be checked for whitelisting because "--"
		// precedes them.
		*args = append(*args, opt.Paths.ArgsHint...)
	}

	// We need to get `git log --source` (the ref by which we reached each commit), but
	// there is no `git log --format=format:...` string that emits the source info; see
	// https://stackoverflow.com/questions/12712775/git-get-source-information-in-format.
	// So we first must run `git log --oneline --source ...` (which does have that info),
	// and then later we will go look up each commit's patch and other info.
	onelineArgs := append([]string{}, args...)
	onelineArgs = append(onelineArgs,
		"-z",
		"--no-abbrev-commit",
		"--format=oneline",
		"--no-color",
		"--source",
		"--no-patch",
	)
	appendCommonQueryArgs(&onelineArgs)
	appendCommonDashDashArgs(&onelineArgs)

	// Run `git log` oneline command and read list of matching commits.
	onelineCmd := gitserver.DefaultClient.Command("git", onelineArgs...)
	onelineCmd.Repo = r.Repo
	data, complete, err := readUntilTimeout(ctx, onelineCmd)
	if err != nil {
		return nil, complete, err
	}
	onelineCommits, err := parseCommitsFromOnelineLog(data)
	if err != nil {
		if !complete {
			// Tolerate parse errors when we received incomplete data.
		} else {
			return nil, complete, err
		}
	}
	// Build a map of commit -> source ref.
	commitSourceRefs := make(map[string]string, len(onelineCommits))
	for _, c := range onelineCommits {
		commitSourceRefs[c.sha1] = c.sourceRef
	}

	// Even though we've already searched using the query, we need to
	// search the returned diff again to filter to only matching hunks
	// and to highlight matches.
	var query *regexp.Regexp
	if pattern := opt.Query.Pattern; pattern != "" {
		if !opt.Query.IsRegExp {
			pattern = regexp.QuoteMeta(pattern)
		}
		if !opt.Query.IsCaseSensitive {
			pattern = "(?i:" + pattern + ")"
		}
		var err error
		query, err = regexp.Compile(pattern)
		if err != nil {
			return nil, false, err
		}
	}

	pathMatcher, err := vcs.CompilePathMatcher(opt.Paths)
	if err != nil {
		return nil, false, err
	}

	// Now fetch the full commit data for all of the commits.
	commitOIDs := make([]string, len(onelineCommits))
	for i, c := range onelineCommits {
		commitOIDs[i] = c.sha1
	}
	showArgs := append([]string{}, "show")
	showArgs = append(showArgs, "--no-patch") // will be overridden if opt.FormatArgs has --patch
	showArgs = append(showArgs, opt.FormatArgs...)
	showArgs = append(showArgs, opt.Args...)
	showArgs = append(showArgs, commitOIDs...)
	appendCommonQueryArgs(&showArgs)
	appendCommonDashDashArgs(&showArgs)
	if !isWhitelistedGitCmd(showArgs) {
		return nil, false, fmt.Errorf("command failed: %q is not a whitelisted git command", showArgs)
	}
	showCmd := gitserver.DefaultClient.Command("git", showArgs...)
	showCmd.Repo = r.Repo
	var complete2 bool
	data, complete2, err = readUntilTimeout(ctx, showCmd)
	if err != nil {
		return nil, complete, err
	}
	complete = complete && complete2
	for len(data) > 0 {
		var commit *vcs.Commit
		var refs []string
		var err error
		commit, refs, data, err = parseCommitFromLog(logFormatFlag, data)
		if err != nil {
			if !complete {
				// Partial data can yield parse errors, but we still want to return what we have.
				// We know all of the results we already found are complete, so we can return
				// immediately instead of marking the last one as incomplete.
				return results, false, nil
			}
			return nil, complete, err
		}

		result := &vcs.LogCommitSearchResult{
			Commit:     *commit,
			Refs:       refs,
			SourceRefs: []string{commitSourceRefs[string(commit.ID)]},
		}
		result.Refs, err = r.filterAndResolveRefs(ctx, result.Refs)
		if err != nil {
			return nil, false, err
		}
		result.SourceRefs, err = r.filterAndResolveRefs(ctx, result.SourceRefs)
		if err != nil {
			return nil, false, err
		}

		if len(data) >= 1 && data[0] == '\x00' {
			// No diff patch (probably no --patch arg).
			data = data[1:]
		} else if len(data) >= 1 && data[0] == '\n' {
			data = data[1:]

			// Next is the diff patch.
			patchEnd := bytes.Index(data, []byte("\n\x00"))
			var rawDiff []byte
			if patchEnd != -1 {
				rawDiff = data[:patchEnd+1]
				data = data[patchEnd+2:]
			} else {
				// Rest of data is the diff patch.
				rawDiff = data
				data = nil
			}

			var err error
			rawDiff, result.DiffHighlights, err = vcs.FilterAndHighlightDiff(rawDiff, query, opt.OnlyMatchingHunks && query != nil, pathMatcher)
			if err != nil {
				return nil, false, err
			}
			if rawDiff == nil {
				continue // it did not match
			}
			result.Diff = &vcs.Diff{Raw: string(rawDiff)}
		}

		results = append(results, result)
	}

	if !complete && len(results) > 0 {
		// The last result may have been parsed from an incomplete output stream (e.g., stdout
		// cut off halfway through), so mark it as such.
		results[len(results)-1].Incomplete = true
	}

	return results, complete, nil
}

// filterAndResolveRefs replaces "HEAD" entries with the names of the ref they refer to,
// and it omits "HEAD -> ..." entries.
func (r *Repository) filterAndResolveRefs(ctx context.Context, refs []string) ([]string, error) {
	var headRefTarget string

	filtered := refs[:0]
	for _, ref := range refs {
		if strings.HasPrefix(ref, "HEAD -> ") {
			continue
		}
		if ref == "HEAD" {
			if headRefTarget == "" {
				cmd := gitserver.DefaultClient.Command("git", "rev-parse", "--symbolic-full-name", "HEAD")
				cmd.Repo = r.Repo
				stdout, err := cmd.Output(ctx)
				if err != nil {
					return nil, err
				}
				headRefTarget = string(bytes.TrimSpace(stdout))
			}
			ref = headRefTarget
		}
		filtered = append(filtered, ref)
	}
	return filtered, nil
}
