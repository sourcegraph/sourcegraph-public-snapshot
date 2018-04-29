package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	regexpsyntax "regexp/syntax"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	validRawLogDiffSearchFormatArgs = [][]string{
		{"--no-merges", "-z", "--decorate=full", "--patch", logFormatWithRefs},
		{"--no-merges", "-z", "--decorate=full", logFormatWithRefs},
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
	deadline, ok := ctx.Deadline()
	var timeoutLabel string
	if ok {
		timeoutLabel = time.Until(deadline).String()
	} else {
		timeoutLabel = "unlimited"
	}

	tr, ctx := trace.New(ctx, "Git: RawLogDiffSearch", fmt.Sprintf("%+v, timeout=%s", opt, timeoutLabel))
	defer func() {
		tr.LazyPrintf("%d results, complete=%v, err=%v", len(results), complete, err)
		tr.SetError(err)
		tr.Finish()
	}()

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
		// If we have exclude paths, we need to effectively unset the --max-count because we can't
		// filter out changes that match the exclude path (because there's no way to use full
		// regexps in git pathspecs).
		//
		// TODO(sqs): use git pathspec %(...) extensions to reduce the number of cases where this is
		// necessary; see https://git-scm.com/docs/gitglossary.html#def_pathspec.
		var addMaxCount500 bool
		if opt.Paths.ExcludePattern != "" {
			addMaxCount500 = true
		}

		// Args we append after this don't need to be checked for whitelisting because "--"
		// precedes them.
		var pathspecs []string
		for _, p := range opt.Paths.IncludePatterns {
			// Roughly try to convert IncludePatterns (regexps) to git pathspecs (globs).
			glob, equiv := regexpToGlobBestEffort(p)
			if !equiv {
				addMaxCount500 = true
			}
			if glob != "" {
				pathspecs = append(pathspecs, glob)
			}
		}

		if addMaxCount500 {
			*args = append(*args, "--max-count=500") // TODO(sqs): 500 is arbitrary high number
		}
		*args = append(*args, "--")
		*args = append(*args, pathspecs...)
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
		"--no-merges",
	)
	appendCommonQueryArgs(&onelineArgs)
	appendCommonDashDashArgs(&onelineArgs)

	// Time out the first `git log` operation prior to the parent context timeout, so we still have time to `git
	// show` the results it returns. These proportions are untuned guesses.
	//
	// TODO(sqs): this can be made much more efficient in many ways
	withTimeout := func(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
		if deadline.IsZero() {
			return ctx, func() {}
		}
		return context.WithTimeout(ctx, timeout)
	}
	// Run `git log` oneline command and read list of matching commits.
	onelineCmd := r.command("git", onelineArgs...)
	logTimeout := time.Until(deadline) / 2
	tr.LazyPrintf("git log %v with timeout %s", onelineCmd.Args, logTimeout)
	ctxLog, cancel := withTimeout(ctx, logTimeout)
	data, complete, err := readUntilTimeout(ctxLog, onelineCmd)
	tr.LazyPrintf("git log done: data %d bytes, complete=%v, err=%v", len(data), complete, err)
	cancel()
	if err != nil {
		// Don't fail if the repository is empty.
		if strings.Contains(err.Error(), "does not have any commits yet") {
			return nil, true, nil
		}

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
	// Need --patch (TODO(sqs): or just --raw, which is smaller) if we are filtering by file paths,
	// because we post-filter by path since we need to support regexps. Just the commit message
	// alone would be insufficient for our post-filtering.
	hasPathFilters := opt.Paths.ExcludePattern != "" || len(opt.Paths.IncludePatterns) > 0
	if hasPathFilters {
		showArgs = append(showArgs, "--patch")
	}
	appendCommonQueryArgs(&showArgs)
	appendCommonDashDashArgs(&showArgs)
	if !isWhitelistedGitCmd(showArgs) {
		return nil, false, fmt.Errorf("command failed: %q is not a whitelisted git command", showArgs)
	}
	showCmd := r.command("git", showArgs...)
	var complete2 bool
	showTimeout := time.Duration(float64(time.Until(deadline)) * 0.8) // leave time for the filterAndResolveRef calls (HACK(sqs): hacky heuristic!)
	tr.LazyPrintf("git show %v with timeout %s", showCmd.Args, showTimeout)
	ctxShow, cancel := withTimeout(ctx, showTimeout)
	data, complete2, err = readUntilTimeout(ctxShow, showCmd)
	tr.LazyPrintf("git show done: data %d bytes, complete=%v, err=%v", len(data), complete2, err)
	cancel()
	if err != nil {
		return nil, complete, err
	}
	complete = complete && complete2
	for len(data) > 0 {
		var commit *vcs.Commit
		var refs []string
		var err error
		commit, refs, data, err = parseCommitFromLog(data)
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
		if err == nil {
			result.SourceRefs, err = r.filterAndResolveRefs(ctx, result.SourceRefs)
		}
		sort.Strings(result.Refs)
		sort.Strings(result.SourceRefs)
		if err != nil {
			if ctx.Err() != nil {
				// Return partial data.
				complete = false
				return results, complete, err
			}
			return nil, false, err
		}

		if len(data) == 0 || (len(data) >= 1 && data[0] == '\x00') {
			// No diff patch.
			if hasPathFilters {
				continue // patch was empty for the filtered paths
			}
			if len(data) >= 1 {
				data = data[1:]
			}
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
			rawDiff, result.DiffHighlights, err = vcs.FilterAndHighlightDiff(rawDiff, query, opt.OnlyMatchingHunks, pathMatcher)
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
				cmd := r.command("git", "rev-parse", "--symbolic-full-name", "HEAD")
				stdout, err := cmd.Output(ctx)
				if err != nil {
					return nil, err
				}
				headRefTarget = string(bytes.TrimSpace(stdout))
			}
			ref = headRefTarget
		} else if strings.HasPrefix(ref, "tag: ") {
			ref = strings.TrimPrefix(ref, "tag: ")
		}
		filtered = append(filtered, ref)
	}
	return filtered, nil
}

// regexpToGlobBestEffort performs a best-effort conversion of the regexp p to an equivalent glob
// pattern. The glob matches a superset of what the regexp matches. If equiv is true, then the glob
// is exactly equivalent to the pattern; otherwise it is a strict superset and post-filtering is
// necessary. The glob never matches a strict subset of p (that would make it possible to correctly
// post-filter).
//
// https://git-scm.com/docs/gitglossary#gitglossary-aiddefpathspecapathspec
func regexpToGlobBestEffort(p string) (glob string, equiv bool) {
	if p == "" {
		return "*", true
	}

	re, err := regexpsyntax.Parse(p, regexpsyntax.OneLine)
	if err != nil {
		return "", false
	}
	switch re.Op {
	case regexpsyntax.OpLiteral:
		return "*" + globQuoteMeta(re.Rune) + "*", true
	case regexpsyntax.OpConcat:
		if len(re.Sub) == 3 && re.Sub[0].Op == regexpsyntax.OpBeginText && re.Sub[1].Op == regexpsyntax.OpLiteral && re.Sub[2].Op == regexpsyntax.OpEndText {
			if len(re.Sub[1].Rune) > 0 && re.Sub[1].Rune[0] == ':' { // leading : has special meaning
				return "", false
			}
			return globQuoteMeta(re.Sub[1].Rune), true
		}
		if len(re.Sub) == 2 && re.Sub[0].Op == regexpsyntax.OpBeginText && re.Sub[1].Op == regexpsyntax.OpLiteral {
			if len(re.Sub[1].Rune) > 0 && re.Sub[1].Rune[0] == ':' { // leading : has special meaning
				return "", false
			}
			return globQuoteMeta(re.Sub[1].Rune) + "*", true
		}
		if len(re.Sub) == 2 && re.Sub[1].Op == regexpsyntax.OpEndText && re.Sub[0].Op == regexpsyntax.OpLiteral {
			return "*" + globQuoteMeta(re.Sub[0].Rune), true
		}
	}
	return "", false
}

func globQuoteMeta(s []rune) string {
	isSpecial := func(c rune) bool {
		switch c {
		case '*':
			return true
		case '?':
			return true
		case '[':
			return true
		case ']':
			return true
		case '\\':
			return true
		default:
			return false
		}
	}

	// Avoid extra work by counting additions. regexp.QuoteMeta does the same,
	// but is more efficient since it does it via bytes.
	count := 0
	for _, c := range s {
		if isSpecial(c) {
			count++
		}
	}
	if count == 0 {
		return string(s)
	}

	escaped := make([]rune, 0, len(s)+count)
	for _, c := range s {
		if isSpecial(c) {
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, c)
	}
	return string(escaped)
}
