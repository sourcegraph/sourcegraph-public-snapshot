package jobutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFileContainsFilterJob creates a filter job to post-filter results for the
// file:contains.content() predicate.
//
// This filter job expects some setup in advance. File results streamed by the
// child should contain matched ranges both for the original pattern and for
// the file:contains.content() patterns. This job will filter out any ranges
// that are matches for the file:contains.content() patterns.
//
// This filter job will also handle filtering diff results so that they only
// include files that contain the pattern specified by file:contains.content().
// Note that this implementation is pretty inefficient, and relies on running
// an unindexed search for each streamed diff match. However, we cannot pre-filter
// because then are not checking whether the file contains the requested content
// at the commit of the diff match.
func NewFileContainsFilterJob(includePatterns []string, originalPattern query.Node, caseSensitive bool, child job.Job) (job.Job, error) {
	includeMatchers := make([]*regexp.Regexp, 0, len(includePatterns))
	for _, pattern := range includePatterns {
		if !caseSensitive {
			pattern = "(?i:" + pattern + ")"
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to regexp.Compile(%q) for file:contains.content() include patterns", pattern)
		}
		includeMatchers = append(includeMatchers, re)
	}

	originalPatternStrings := patternsInTree(originalPattern)
	originalPatternMatchers := make([]*regexp.Regexp, 0, len(originalPatternStrings))
	for _, originalPatternString := range originalPatternStrings {
		if !caseSensitive {
			originalPatternString = "(?i:" + originalPatternString + ")"
		}
		re, err := regexp.Compile(originalPatternString)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to regexp.Compile(%q) for file:contains.content() original patterns", originalPatternString)
		}
		originalPatternMatchers = append(originalPatternMatchers, re)
	}

	return &fileContainsFilterJob{
		caseSensitive:           caseSensitive,
		includePatterns:         includePatterns,
		includeMatchers:         includeMatchers,
		originalPatternMatchers: originalPatternMatchers,
		child:                   child,
	}, nil
}

type fileContainsFilterJob struct {
	// We maintain the original input patterns and case-sensitivity because
	// searcher does not correctly handle case-insensitive `(?i:)` regex
	// patterns. The logic for longest substring is incorrect for
	// case-insensitive patterns (it returns the all-upper-case version of the
	// longest substring) and will fail to find any matches.
	caseSensitive   bool
	includePatterns []string

	// Regex patterns specified by file:contains.content()
	includeMatchers []*regexp.Regexp

	// Regex patterns specified as part of the original pattern
	originalPatternMatchers []*regexp.Regexp

	child job.Job
}

func (j *fileContainsFilterJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		event = j.filterEvent(ctx, clients.SearcherURLs, clients.SearcherGRPCConnectionCache, event)
		stream.Send(event)
	})

	return j.child.Run(ctx, clients, filteredStream)
}

func (j *fileContainsFilterJob) filterEvent(ctx context.Context, searcherURLs *endpoint.Map, searcherGRPCConnectionCache *defaults.ConnectionCache, event streaming.SearchEvent) streaming.SearchEvent {
	// Don't filter out files with zero chunks because if the file contained
	// a result, we still want to return a match for the file even if it
	// has no matched ranges left.
	filtered := event.Results[:0]
	for _, res := range event.Results {
		switch v := res.(type) {
		case *result.FileMatch:
			filtered = append(filtered, j.filterFileMatch(v))
		case *result.CommitMatch:
			cm := j.filterCommitMatch(ctx, searcherURLs, searcherGRPCConnectionCache, v)
			if cm != nil {
				filtered = append(filtered, cm)
			}
		default:
			// Filter out any results that are not FileMatch or CommitMatch
		}
	}
	event.Results = filtered
	return event
}

func (j *fileContainsFilterJob) filterFileMatch(fm *result.FileMatch) result.Match {
	filteredChunks := fm.ChunkMatches[:0]
	for _, chunk := range fm.ChunkMatches {
		chunk = j.filterChunk(chunk)
		if len(chunk.Ranges) == 0 {
			// Skip any chunks where we filtered out all the matched ranges
			continue
		}
		filteredChunks = append(filteredChunks, chunk)
	}
	// A file match with zero chunks after filtering is still valid, and just
	// becomes a path match
	fm.ChunkMatches = filteredChunks
	return fm
}

func (j *fileContainsFilterJob) filterChunk(chunk result.ChunkMatch) result.ChunkMatch {
	filteredRanges := chunk.Ranges[:0]
	for i, val := range chunk.MatchedContent() {
		if matchesAny(val, j.includeMatchers) && !matchesAny(val, j.originalPatternMatchers) {
			continue
		}
		filteredRanges = append(filteredRanges, chunk.Ranges[i])
	}
	chunk.Ranges = filteredRanges
	return chunk
}

func matchesAny(val string, matchers []*regexp.Regexp) bool {
	for _, re := range matchers {
		if re.MatchString(val) {
			return true
		}
	}
	return false
}

func (j *fileContainsFilterJob) filterCommitMatch(ctx context.Context, searcherURLs *endpoint.Map, searcherGRPCConnectionCache *defaults.ConnectionCache, cm *result.CommitMatch) result.Match {
	// Skip any commit matches -- we only handle diff matches
	if cm.DiffPreview == nil {
		return nil
	}

	fileNames := make([]string, 0, len(cm.Diff))
	for _, fileDiff := range cm.Diff {
		fileNames = append(fileNames, regexp.QuoteMeta(fileDiff.NewName))
	}

	// For each pattern specified by file:contains.content(), run a search at
	// the commit to ensure that the file does, in fact, contain that content.
	// We cannot do this all at once because searcher does not support complex patterns.
	// Additionally, we cannot do this in advance because we don't know which commit
	// we are searching at until we get a result.
	matchedFileCounts := make(map[string]int)
	for _, includePattern := range j.includePatterns {
		patternInfo := search.TextPatternInfo{
			Pattern:               includePattern,
			IsCaseSensitive:       j.caseSensitive,
			IsRegExp:              true,
			FileMatchLimit:        99999999,
			Index:                 query.No,
			IncludePatterns:       []string{query.UnionRegExps(fileNames)},
			PatternMatchesContent: true,
		}

		onMatch := func(fm *protocol.FileMatch) {
			matchedFileCounts[fm.Path] += 1
		}

		_, err := searcher.Search(
			ctx,
			searcherURLs,
			searcherGRPCConnectionCache,
			cm.Repo.Name,
			cm.Repo.ID,
			"",
			cm.Commit.ID,
			false,
			&patternInfo,
			time.Hour,
			search.Features{},
			0, // we don't care about the actual content, so don't fetch extra lines
			onMatch,
		)
		if err != nil {
			// Ignore any files where the search errors
			return nil
		}
	}

	return j.removeUnmatchedFileDiffs(cm, matchedFileCounts)
}

func (j *fileContainsFilterJob) removeUnmatchedFileDiffs(cm *result.CommitMatch, matchedFileCounts map[string]int) result.Match {
	// Ensure the matched ranges are sorted by start offset
	slices.SortFunc(cm.DiffPreview.MatchedRanges, func(a, b result.Range) bool {
		return a.Start.Offset < b.End.Offset
	})

	// Convert each file diff to a string so we know how much we are removing if we drop that file
	diffStrings := make([]string, 0, len(cm.Diff))
	for _, fileDiff := range cm.Diff {
		diffStrings = append(diffStrings, result.FormatDiffFiles([]result.DiffFile{fileDiff}))
	}

	// groupedRanges[i] will be the set of ranges that are contained by diffStrings[i]
	groupedRanges := make([]result.Ranges, len(cm.Diff))
	{
		rangeNumStart := 0
		currentDiffEnd := 0
	OUTER:
		for i, diffString := range diffStrings {
			currentDiffEnd += len(diffString)
			for rangeNum := rangeNumStart; rangeNum < len(cm.DiffPreview.MatchedRanges); rangeNum++ {
				currentRange := cm.DiffPreview.MatchedRanges[rangeNum]
				if currentRange.Start.Offset > currentDiffEnd {
					groupedRanges[i] = cm.DiffPreview.MatchedRanges[rangeNumStart:rangeNum]
					rangeNumStart = rangeNum
					continue OUTER
				}
			}
			groupedRanges[i] = cm.DiffPreview.MatchedRanges[rangeNumStart:]
		}

	}

	filteredRanges := groupedRanges[:0]
	filteredDiffs := cm.Diff[:0]
	filteredDiffStrings := diffStrings[:0]
	removedAmount := result.Location{}
	for i, fileDiff := range cm.Diff {
		if count := matchedFileCounts[fileDiff.NewName]; count == len(j.includeMatchers) {
			filteredDiffs = append(filteredDiffs, fileDiff)
			filteredDiffStrings = append(filteredDiffStrings, diffStrings[i])
			filteredRanges = append(filteredRanges, groupedRanges[i].Sub(removedAmount))
		} else {
			// If count != len(j.includeMatchers), that means that not all of our file:contains.content() patterns
			// matched and this fileDiff should be dropped. Skip appending it, and add its length to the removed amount
			// so we can adjust the matched ranges down.
			removedAmount = removedAmount.Add(result.Location{Offset: len(diffStrings[i]), Line: strings.Count(diffStrings[i], "\n")})
		}
	}

	// Re-merge groupedRanges
	ungroupedRanges := result.Ranges{}
	for _, grouped := range filteredRanges {
		ungroupedRanges = append(ungroupedRanges, grouped...)
	}

	// Update the commit match with the filtered slices
	cm.DiffPreview.MatchedRanges = ungroupedRanges
	cm.DiffPreview.Content = strings.Join(filteredDiffStrings, "")
	cm.Diff = filteredDiffs
	if len(cm.Diff) > 0 {
		return cm
	} else {
		// Return nil if this whole result should be filtered out
		return nil
	}
}

func (j *fileContainsFilterJob) MapChildren(f job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, f)
	return &cp
}

func (j *fileContainsFilterJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *fileContainsFilterJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		originalPatternStrings := make([]string, 0, len(j.originalPatternMatchers))
		for _, re := range j.originalPatternMatchers {
			originalPatternStrings = append(originalPatternStrings, re.String())
		}
		res = append(res, attribute.StringSlice("originalPatterns", originalPatternStrings))

		filterStrings := make([]string, 0, len(j.includeMatchers))
		for _, re := range j.includeMatchers {
			filterStrings = append(filterStrings, re.String())
		}
		res = append(res, attribute.StringSlice("filterPatterns", filterStrings))
	}
	return res
}

func (j *fileContainsFilterJob) Name() string {
	return "FileContainsFilterJob"
}

func patternsInTree(originalPattern query.Node) (res []string) {
	if originalPattern == nil {
		return nil
	}
	switch v := originalPattern.(type) {
	case query.Operator:
		for _, operand := range v.Operands {
			res = append(res, patternsInTree(operand)...)
		}
	case query.Pattern:
		res = append(res, v.Value)
	default:
		panic(fmt.Sprintf("unknown pattern node type %T", originalPattern))
	}
	return res
}
