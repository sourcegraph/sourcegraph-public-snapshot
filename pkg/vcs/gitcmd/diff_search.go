package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"reflect"
	"regexp"

	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	validRawLogDiffSearchFormatArgs = [][]string{
		{"-z", "--patch", logFormatFlag},
		{"-z", logFormatFlag},
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
		if opt.Query.Pattern == "" {
			opt.FormatArgs = validRawLogDiffSearchFormatArgs[1] // without --patch
		} else {
			opt.FormatArgs = validRawLogDiffSearchFormatArgs[0] // with --patch
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
	if opt.Paths.IsRegExp {
		args = append(args, "--extended-regexp")
	}
	if !isWhitelistedGitCmd(args) {
		return nil, false, fmt.Errorf("command failed: %q is not a whitelisted git command", args)
	}
	args = append(args, "--")
	// Args we append after this don't need to be checked for whitelisting because "--"
	// precedes them.
	args = append(args, opt.Paths.ArgsHint...)

	// Run command and read output.
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = r.Repo
	sr, err := gitserver.StdoutReader(ctx, cmd)
	if urlErr, ok := err.(*url.Error); ok && urlErr.Err == context.DeadlineExceeded {
		// Continue; the gitserver call exceeded our deadline before the command
		// produced any output.
	} else if err != nil {
		return nil, false, err
	}
	var data []byte
	if sr != nil {
		defer sr.Close()
		var err error
		data, err = ioutil.ReadAll(sr)
		if err == nil {
			complete = true
		} else if err != nil && err != context.DeadlineExceeded {
			data = bytes.TrimSpace(data)
			if isBadObjectErr(string(data), "") || isInvalidRevisionRangeError(string(data), "") {
				return nil, true, vcs.ErrRevisionNotFound
			}
			if len(data) > 100 {
				data = append(data[:100], []byte("... (truncated)")...)
			}
			return nil, true, fmt.Errorf("exec `git log` failed: %s. Output was:\n\n%s", err, data)
		}
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
		return nil, false, err
	}
	pathMatcher, err := vcs.CompilePathMatcher(opt.Paths)
	if err != nil {
		return nil, false, err
	}

	for len(data) > 0 {
		var commit *vcs.Commit
		var err error
		commit, data, err = parseCommitFromLog(logFormatFlag, data)
		if err != nil {
			if !complete {
				// Partial data can yield parse errors, but we still want to return what we have.
				// We know all of the results we already found are complete, so we can return
				// immediately instead of marking the last one as incomplete.
				return results, false, nil
			}
			return nil, complete, err
		}

		result := &vcs.LogCommitSearchResult{Commit: *commit}

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
