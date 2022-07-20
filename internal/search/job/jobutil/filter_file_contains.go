package jobutil

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

	return &fileContainsFilterJob{
		includeMatchers:        includeMatchers,
		originalPatternMatcher: newPatternTreeMatcher(originalPattern, caseSensitive),
		child:                  child,
	}
}

type fileContainsFilterJob struct {
	includeMatchers        []*regexp.Regexp
	originalPatternMatcher patternTreeMatcher
	child                  job.Job
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
		if matchesAny(val, j.includeMatchers) && !j.originalPatternMatcher.Match(val) {
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
		res = append(res, trace.Stringer("originalPattern", j.originalPatternMatcher))
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

func newPatternTreeMatcher(originalPattern query.Node, caseSensitive bool) patternTreeMatcher {
	if originalPattern == nil {
		return &constMatcher{false}
	}
	switch v := originalPattern.(type) {
	case query.Operator:
		children := make([]patternTreeMatcher, 0, len(v.Operands))
		for _, operand := range v.Operands {
			children = append(children, newPatternTreeMatcher(operand, caseSensitive))
		}
		switch v.Kind {
		case query.And:
			return &andMatcher{children: children}
		case query.Or:
			return &orMatcher{children: children}
		default:
			panic(fmt.Sprintf("cannot handle operator kind %d", v.Kind))
		}
	case query.Pattern:
		value := v.Value
		if caseSensitive {
			value = "(?i:" + value + ")"

		}
		return &leafMatcher{
			re: regexp.MustCompile(value), // already validated
		}
	default:
		panic(fmt.Sprintf("unknown pattern node type %T", originalPattern))
	}
}

type patternTreeMatcher interface {
	Match(string) bool
	String() string
}

type orMatcher struct {
	children []patternTreeMatcher
}

func (m *orMatcher) Match(val string) bool {
	for _, child := range m.children {
		if child.Match(val) {
			return true
		}
	}
	return false
}

func (m *orMatcher) String() string {
	childStrings := make([]string, 0, len(m.children))
	for _, child := range m.children {
		childStrings = append(childStrings, child.String())
	}
	return "(" + strings.Join(childStrings, " OR ") + ")"
}

type andMatcher struct {
	children []patternTreeMatcher
}

func (m *andMatcher) Match(val string) bool {
	for _, child := range m.children {
		if !child.Match(val) {
			return false
		}
	}
	return true
}

func (m *andMatcher) String() string {
	childStrings := make([]string, 0, len(m.children))
	for _, child := range m.children {
		childStrings = append(childStrings, child.String())
	}
	return "(" + strings.Join(childStrings, " AND ") + ")"
}

type leafMatcher struct {
	re *regexp.Regexp
}

func (m *leafMatcher) Match(val string) bool {
	return m.re.MatchString(val)
}

func (m *leafMatcher) String() string {
	return fmt.Sprintf("%q", m.re.String())
}

type constMatcher struct {
	value bool
}

func (m *constMatcher) Match(val string) bool {
	return m.value
}

func (m *constMatcher) String() string {
	return strconv.FormatBool(m.value)
}
