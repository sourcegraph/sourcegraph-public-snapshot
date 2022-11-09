package jobutil

import (
	"context"
	"github.com/grafana/regexp"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func NewSanitizeJob(sanitizePatterns []*regexp.Regexp, child job.Job) job.Job {
	if len(sanitizePatterns) == 0 {
		return &NoopJob{}
	}

	return &SanitizeJob{
		sanitizePatterns: sanitizePatterns,
		child:            child,
	}
}

type SanitizeJob struct {
	sanitizePatterns []*regexp.Regexp
	child            job.Job
}

func (j *SanitizeJob) Name() string {
	return "SanitizeJob"
}

func (j *SanitizeJob) Fields(job.Verbosity) []otlog.Field {
	return nil
}

func (j *SanitizeJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *SanitizeJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, fn)
	return &cp
}

func (j *SanitizeJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, s, j)
	defer func() { finish(alert, err) }()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		event = j.sanitizeEvent(event)
		stream.Send(event)
	})

	return j.child.Run(ctx, clients, filteredStream)
}

func (j *SanitizeJob) sanitizeEvent(event streaming.SearchEvent) streaming.SearchEvent {
	sanitized := event.Results[:0]

	for _, res := range event.Results {
		switch v := res.(type) {
		case *result.FileMatch:
			if sanitizedFileMatch := j.sanitizeFileMatch(v); sanitizedFileMatch != nil {
				sanitized = append(sanitized, sanitizedFileMatch)
			}
		case *result.CommitMatch:
			if sanitizedCommitMatch := j.sanitizeCommitMatch(v); sanitizedCommitMatch != nil {
				sanitized = append(sanitized, sanitizedCommitMatch)
			}
		default:
			// Don't worry about result types that don't expose file content
			sanitized = append(sanitized, v)
		}
	}

	event.Results = sanitized
	return event
}

func (j *SanitizeJob) sanitizeFileMatch(fm *result.FileMatch) result.Match {
	sanitizedChunks := fm.ChunkMatches[:0]

	for _, chunk := range fm.ChunkMatches {
		chunk = j.sanitizeChunk(chunk)
		if len(chunk.Ranges) == 0 {
			continue
		}
		sanitizedChunks = append(sanitizedChunks, chunk)
	}

	if len(sanitizedChunks) == 0 {
		return nil
	}
	fm.ChunkMatches = sanitizedChunks
	return fm
}

func (j *SanitizeJob) sanitizeChunk(chunk result.ChunkMatch) result.ChunkMatch {
	sanitizedRanges := chunk.Ranges[:0]

	for i, val := range chunk.MatchedContent() {
		if j.matchesAnySanitizePattern(val) {
			continue
		}
		sanitizedRanges = append(sanitizedRanges, chunk.Ranges[i])
	}

	chunk.Ranges = sanitizedRanges
	return chunk
}

func (j *SanitizeJob) sanitizeCommitMatch(cm *result.CommitMatch) result.Match {
	if cm.DiffPreview == nil {
		return cm
	}
	if j.matchesAnySanitizePattern(cm.DiffPreview.Content) {
		return nil
	}
	return cm
}

func (j *SanitizeJob) matchesAnySanitizePattern(val string) bool {
	for _, re := range j.sanitizePatterns {
		if re.MatchString(val) {
			return true
		}
	}
	return false
}
