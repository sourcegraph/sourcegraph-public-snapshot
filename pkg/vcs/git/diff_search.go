package git

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/pathmatch"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

// TextSearchOptions contains common options for text search commands.
type TextSearchOptions struct {
	Pattern         string // the pattern to look for
	IsRegExp        bool   // whether the pattern is a regexp (if false, treated as exact string)
	IsCaseSensitive bool   // whether the pattern should be matched case-sensitively
}

// PathOptions contains common options for commands that can be limited
// to only certain paths.
type PathOptions struct {
	IncludePatterns []string // include paths matching all of these patterns
	ExcludePattern  string   // exclude paths matching any of these patterns
	IsRegExp        bool     // whether the pattern is a regexp (if false, treated as exact string)
	IsCaseSensitive bool     // whether the pattern should be matched case-sensitively
}

// CompilePathMatcher compiles the path options into a PathMatcher.
func CompilePathMatcher(options PathOptions) (pathmatch.PathMatcher, error) {
	return pathmatch.CompilePathPatterns(
		options.IncludePatterns, options.ExcludePattern,
		pathmatch.CompileOptions{CaseSensitive: options.IsCaseSensitive, RegExp: options.IsRegExp},
	)
}

// RawLogDiffSearchOptions specifies options to (Repository).RawLogDiffSearch.
type RawLogDiffSearchOptions struct {
	// Query specifies the search query to find.
	Query TextSearchOptions

	// MatchChangedOccurrenceCount makes the operation run `git log -S` not `git log -G`.
	// See `git log --help` for more information.
	MatchChangedOccurrenceCount bool

	// Diff is whether the diff should be computed and returned.
	Diff bool

	// OnlyMatchingHunks makes the diff only include hunks that match the query. If false,
	// all hunks from files that match the query are included.
	OnlyMatchingHunks bool

	// Paths specifies the paths to include/exclude.
	Paths PathOptions

	// FormatArgs is a list of format args that are passed to the `git log` command.
	// Because the output is parsed, it is expected to be in a known format. If the
	// FormatArgs does not match one of the server's expected values, the operation
	// will fail.
	//
	// If nil, the default format args are used.
	FormatArgs []string

	// RawArgs is a list of non-format args that are passed to the `git log` command.
	// It should not contain any "--" elements; those should be passed using the Paths
	// field.
	//
	// No arguments that affect the format of the output should be present in this
	// slice.
	Args []string
}

// LogCommitSearchResult describes a matching diff from (Repository).RawLogDiffSearch.
type LogCommitSearchResult struct {
	Commit         Commit      // the commit whose diff was matched
	Diff           *Diff       // the diff, with non-matching/irrelevant portions deleted (respecting diff syntax)
	DiffHighlights []Highlight // highlighted query matches in the diff

	// Refs is the list of ref names of this commit (from `git log --decorate`).
	Refs []string

	// SourceRefs is the list of ref names by which this commit was reached. (See
	// `git log --help` documentation on the `--source` flag.)
	SourceRefs []string

	// Incomplete indicates that this result may represent a subset of the actual data.
	// This can occur when the underlying command returns early due to an impending
	// timeout.
	Incomplete bool
}

// A Diff represents changes between two commits.
type Diff struct {
	Raw string // the raw diff output
}

// Highlight represents a highlighted region in a string.
type Highlight struct {
	Line      int // the 1-indexed line number
	Character int // the 1-indexed character on the line
	Length    int // the length of the highlight, in characters (on the same line)
}

var validRawLogDiffSearchFormatArgs = [][]string{
	{"--no-merges", "-z", "--decorate=full", "--patch", logFormatWithRefs},
	{"--no-merges", "-z", "--decorate=full", logFormatWithRefs},
}

func isValidRawLogDiffSearchFormatArgs(formatArgs []string) bool {
	for _, validArgs := range validRawLogDiffSearchFormatArgs {
		if reflect.DeepEqual(formatArgs, validArgs) {
			return true
		}
	}
	return false
}

// RawLogDiffSearch runs a raw `git log` command that is expected to return logs with patches. It
// returns a subset of the output, including only hunks that actually match the given pattern.
//
// If complete is false, then the results may have been parsed from only partial output from the
// underlying git command (because, e.g., it timed out during execution and only returned partial
// output).
func RawLogDiffSearch(ctx context.Context, repo gitserver.Repo, opt RawLogDiffSearchOptions) (results []*LogCommitSearchResult, complete bool, err error) {
	if Mocks.RawLogDiffSearch != nil {
		return Mocks.RawLogDiffSearch(opt)
	}

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
			if !opt.Paths.IsCaseSensitive && glob != "" {
				// This relies on regexpToGlobBestEffort not returning `:`-prefixed globs.
				glob = ":(icase)" + glob
			}
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
	onelineCmd := gitserver.DefaultClient.Command("git", onelineArgs...)
	onelineCmd.Repo = repo
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

	pathMatcher, err := compilePathMatcher(opt.Paths)
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
	showCmd := gitserver.DefaultClient.Command("git", showArgs...)
	showCmd.Repo = repo
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
	var cache refResolveCache
	for len(data) > 0 {
		var commit *Commit
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

		result := &LogCommitSearchResult{
			Commit:     *commit,
			Refs:       refs,
			SourceRefs: []string{commitSourceRefs[string(commit.ID)]},
		}
		result.Refs, err = filterAndResolveRefs(ctx, repo, result.Refs, &cache)
		if err == nil {
			result.SourceRefs, err = filterAndResolveRefs(ctx, repo, result.SourceRefs, &cache)
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

		hasMatch := true
		if len(data) == 0 || (len(data) >= 1 && data[0] == '\x00') {
			// No diff patch.
			if len(data) >= 1 {
				data = data[1:]
			}
			if hasPathFilters {
				hasMatch = false // patch was empty for the filtered paths, don't add to results
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
			rawDiff, result.DiffHighlights, err = FilterAndHighlightDiff(rawDiff, query, opt.OnlyMatchingHunks, pathMatcher)
			if err != nil {
				return nil, false, err
			}
			if rawDiff == nil {
				hasMatch = false // patch was empty (after applying filters), don't add to results
			} else {
				result.Diff = &Diff{Raw: string(rawDiff)}
			}
		}

		if hasMatch {
			results = append(results, result)
		}
	}

	if !complete && len(results) > 0 {
		// The last result may have been parsed from an incomplete output stream (e.g., stdout
		// cut off halfway through), so mark it as such.
		results[len(results)-1].Incomplete = true
	}

	return results, complete, nil
}

// cachedRefResolver is a short-lived cache for ref resolutions. Only use it for the lifetime of a
// single request and for a single repo.
type refResolveCache struct {
	mu      sync.Mutex
	results map[string]struct {
		target string
		err    error
	}
}

func (r *refResolveCache) resolveHEADSymbolicRef(ctx context.Context, repo gitserver.Repo) (target string, err error) {
	resolve := func() (string, error) {
		cmd := gitserver.DefaultClient.Command("git", "rev-parse", "--symbolic-full-name", "HEAD")
		cmd.Repo = repo
		stdout, err := cmd.Output(ctx)
		return string(bytes.TrimSpace(stdout)), err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.results == nil {
		r.results = map[string]struct {
			target string
			err    error
		}{}
	}
	const name = "HEAD" // only needed for HEAD right now
	e, ok := r.results[name]
	if !ok {
		e.target, e.err = resolve()
		r.results[name] = e
	}
	return e.target, e.err
}

// filterAndResolveRefs replaces "HEAD" entries with the names of the ref they refer to,
// and it omits "HEAD -> ..." entries.
func filterAndResolveRefs(ctx context.Context, repo gitserver.Repo, refs []string, cache *refResolveCache) ([]string, error) {
	filtered := refs[:0]
	for _, ref := range refs {
		if strings.HasPrefix(ref, "HEAD -> ") {
			continue
		}
		if ref == "HEAD" {
			var err error
			ref, err = cache.resolveHEADSymbolicRef(ctx, repo)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(ref, "tag: ") {
			ref = strings.TrimPrefix(ref, "tag: ")
		}
		filtered = append(filtered, ref)
	}
	return filtered, nil
}
