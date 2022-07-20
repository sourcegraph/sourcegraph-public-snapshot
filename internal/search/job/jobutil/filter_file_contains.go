package jobutil

import (
	"context"
	"fmt"
	"regexp"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func NewFileContainsFilterJob(includePatterns []string, originalPattern query.Node, caseSensitive bool, child job.Job) job.Job {
	includeMatchers := make([]*regexp.Regexp, 0, len(includePatterns))
	for _, pattern := range includePatterns {
		if !caseSensitive {
			pattern = "(?i:" + pattern + ")"
		}
		includeMatchers = append(includeMatchers, regexp.MustCompile(pattern))
	}

	originalPatternStrings := patternsInTree(originalPattern)
	originalPatternMatchers := make([]*regexp.Regexp, 0, len(originalPatternStrings))
	for _, originalPatternString := range originalPatternStrings {
		if !caseSensitive {
			originalPatternString = "(?i:" + originalPatternString + ")"
		}
		originalPatternMatchers = append(originalPatternMatchers, regexp.MustCompile(originalPatternString))
	}

	return &fileContainsFilterJob{
		includeMatchers:         includeMatchers,
		originalPatternMatchers: originalPatternMatchers,
		child:                   child,
	}
}

type fileContainsFilterJob struct {
	includeMatchers         []*regexp.Regexp
	originalPatternMatchers []*regexp.Regexp
	child                   job.Job
}

func (j *fileContainsFilterJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		event = j.filterEvent(event)
		stream.Send(event)
	})

	return j.child.Run(ctx, clients, filteredStream)
}

func (j *fileContainsFilterJob) filterEvent(event streaming.SearchEvent) streaming.SearchEvent {
	// Don't filter out files with zero chunks because if the file contained
	// the a result, we still want to return a match for the file even if it
	// has no matched ranges left.
	for i := range event.Results {
		event.Results[i] = j.filterFileMatch(event.Results[i])
	}
	return event
}

func (j *fileContainsFilterJob) filterFileMatch(m result.Match) result.Match {
	fm, ok := m.(*result.FileMatch)
	if !ok {
		return m
	}

	filteredChunks := fm.ChunkMatches[:0]
	for _, chunk := range fm.ChunkMatches {
		chunk = j.filterChunk(chunk)
		if len(chunk.Ranges) == 0 {
			continue
		}
		filteredChunks = append(filteredChunks, chunk)
	}
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

func (j *fileContainsFilterJob) MapChildren(f job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, f)
	return &cp
}

func (j *fileContainsFilterJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *fileContainsFilterJob) Fields(v job.Verbosity) (res []otlog.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		originalPatternStrings := make([]string, 0, len(j.originalPatternMatchers))
		for _, re := range j.originalPatternMatchers {
			originalPatternStrings = append(originalPatternStrings, re.String())
		}
		res = append(res, trace.Strings("originalPatterns", originalPatternStrings))

		filterStrings := make([]string, 0, len(j.includeMatchers))
		for _, re := range j.includeMatchers {
			filterStrings = append(filterStrings, re.String())
		}
		res = append(res, trace.Strings("filterPatterns", filterStrings))
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
