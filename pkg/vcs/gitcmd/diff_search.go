package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	validRawLogDiffSearchFormatArgs = [][]string{
		{"-z", "--patch", logFormatFlag},
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
func (r *Repository) RawLogDiffSearch(ctx context.Context, opt vcs.RawLogDiffSearchOptions) ([]*vcs.LogCommitSearchResult, error) {
	if opt.FormatArgs == nil {
		opt.FormatArgs = validRawLogDiffSearchFormatArgs[0]
	}
	if opt.FormatArgs != nil && !isValidRawLogDiffSearchFormatArgs(opt.FormatArgs) {
		return nil, fmt.Errorf("invalid FormatArgs: %q", opt.FormatArgs)
	}
	for _, arg := range opt.Args {
		if arg == "--" {
			return nil, fmt.Errorf("invalid Args (must not contain \"--\" element): %q", opt.Args)
		}
	}

	if opt.Query.IsCaseSensitive != opt.Paths.IsCaseSensitive {
		// These options can't be set separately in `git log`, so fail.
		return nil, fmt.Errorf("invalid options: Query.IsCaseSensitive != Paths.IsCaseSensitive")
	}

	args := []string{"log"}
	args = append(args, opt.FormatArgs...)
	args = append(args, opt.Args...)
	if opt.Query.Pattern != "" {
		var queryArg string
		if opt.MatchChangedOccurrenceCount {
			queryArg = "-S"
		} else {
			queryArg = "-G"
		}
		args = append(args, queryArg+opt.Query.Pattern)
		if !opt.Query.IsCaseSensitive {
			args = append(args, "--regexp-ignore-case")
		}
		if opt.Query.IsRegExp {
			args = append(args, "--pickaxe-regex")
		}
	}
	if !isWhitelistedGitCmd(args) {
		return nil, fmt.Errorf("command failed: %q is not a whitelisted git command", args)
	}
	args = append(args, "--")
	// Args we append after this don't need to be checked for whitelisting because "--"
	// precedes them.
	args = append(args, opt.Paths.ArgsHint...)

	// Run command and parse output.
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = r.Repo
	data, err := cmd.CombinedOutput(ctx)
	if err != nil {
		data = bytes.TrimSpace(data)
		if isBadObjectErr(string(data), "") || isInvalidRevisionRangeError(string(data), "") {
			return nil, vcs.ErrRevisionNotFound
		}
		return nil, fmt.Errorf("exec `git log` failed: %s. Output was:\n\n%s", err, data)
	}

	// Even though we've already searched using the query, we need to
	// search the returned diff again to filter to only matching hunks
	// and to highlight matches.
	pattern := opt.Query.Pattern
	if !opt.Query.IsRegExp {
		pattern = regexp.QuoteMeta(pattern)
	}
	if !opt.Query.IsCaseSensitive {
		pattern = "(?i:" + pattern + ")"
	}
	query, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	pathMatcher, err := vcs.CompilePathMatcher(opt.Paths)
	if err != nil {
		return nil, err
	}

	var results []*vcs.LogCommitSearchResult
	for len(data) > 0 {
		var commit *vcs.Commit
		var err error
		commit, data, err = parseCommitFromLog(logFormatFlag, data)
		if err != nil {
			return nil, err
		}

		result := &vcs.LogCommitSearchResult{Commit: *commit}

		if len(data) >= 1 && data[0] == '\n' {
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
			rawDiff, result.Highlights, err = vcs.FilterAndHighlightDiff(rawDiff, query, opt.OnlyMatchingHunks, pathMatcher)
			if err != nil {
				return nil, err
			}
			if rawDiff == nil {
				continue // it did not match
			}
			result.Diff = vcs.Diff{Raw: string(rawDiff)}
		}

		results = append(results, result)
	}

	return results, nil
}
