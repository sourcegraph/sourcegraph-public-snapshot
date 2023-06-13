package jobutil

import (
	"context"
	"regexp"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewFileHasContributorsJob(child job.Job, caseSensitive bool, include, exclude []string) (job.Job, error) {
	includeContributors, excludeContributors, err := compileRegexps(include, exclude, caseSensitive)
	if err != nil {
		return nil, err
	}

	return &fileHasContributorsJob{
		child:               child,
		includeContributors: includeContributors,
		excludeContributors: excludeContributors,
		includeTerms:        include,
		excludeTerms:        exclude,
	}, nil
}

type fileHasContributorsJob struct {
	child job.Job

	includeContributors []*regexp.Regexp
	excludeContributors []*regexp.Regexp

	includeTerms []string
	excludeTerms []string
}

func (j *fileHasContributorsJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer finish(alert, err)

	var (
		mu   sync.Mutex
		errs error
	)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		filtered := event.Results[:0]
		for _, res := range event.Results {
			// Filter out any result that is not a file
			if fm, ok := res.(*result.FileMatch); ok {
				// We send one fetch contributors request per file path.
				// We should quit early on context deadline exceeded.
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					errs = errors.Append(errs, ctx.Err())
					break
				}
				fileMatchContributors, err := getFileContributors(ctx, clients.Gitserver, fm)
				if err != nil {
					mu.Lock()
					errs = errors.Append(errs, err)
					mu.Unlock()
					continue
				}

				// ensure match passes all exclusion filters
				excludeFilters := j.Filtered(fileMatchContributors, true)

				// ensure match passes all inclusion filters
				includeFilters := j.Filtered(fileMatchContributors, false)

				if !excludeFilters || !includeFilters {
					continue
				}

				filtered = append(filtered, fm)
			}
		}

		event.Results = filtered

		stream.Send(event)
	})

	alert, err = j.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (j *fileHasContributorsJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, fn)
	return &cp
}

func (j *fileHasContributorsJob) Name() string {
	return "FileHasContributorsFilterJob"
}

func (j *fileHasContributorsJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *fileHasContributorsJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.StringSlice("includeContributors", j.includeTerms),
			attribute.StringSlice("excludeContributors", j.excludeTerms),
		)
	}
	return res
}

func compileRegexps(include, exclude []string, caseSensitive bool) (includeRegexp, excludeRegexp []*regexp.Regexp, err error) {
	includeRegexp, err = regexps(include, caseSensitive)
	if err != nil {
		return nil, nil, err
	}

	excludeRegexp, err = regexps(exclude, caseSensitive)
	if err != nil {
		return nil, nil, err
	}

	return includeRegexp, excludeRegexp, nil
}

func regexps(filters []string, caseSensitive bool) ([]*regexp.Regexp, error) {
	var compiledFilters []*regexp.Regexp
	for _, contributorExpression := range filters {
		if !caseSensitive {
			contributorExpression = "(?i)" + contributorExpression
		}
		re, err := regexp.Compile(contributorExpression)
		if err != nil {
			return nil, err
		}
		compiledFilters = append(compiledFilters, re)
	}
	return compiledFilters, nil
}

func getFileContributors(ctx context.Context, client gitserver.Client, fm *result.FileMatch) ([]*gitdomain.ContributorCount, error) {
	opts := gitserver.ContributorOptions{
		Range: string(fm.CommitID),
		Path:  fm.Path,
	}
	contributors, err := client.ContributorCount(ctx, fm.Repo.Name, opts)

	if err != nil {
		return nil, err
	}

	return contributors, nil
}

// Filtered returns true if the match passes filter validation and should be returned with results page.
// Filters are AND'ed together. Filters are negation filters if excludeContributors is true.
func (j *fileHasContributorsJob) Filtered(contributors []*gitdomain.ContributorCount, excludeContributors bool) bool {
	filters := j.includeContributors
	if excludeContributors {
		filters = j.excludeContributors
	}
	for _, filter := range filters {
		if match(contributors, filter) == excludeContributors {
			// Result needs to be filtered out
			return false
		}
	}

	// Result passed all filters
	return true
}

func match(contributors []*gitdomain.ContributorCount, regexp *regexp.Regexp) bool {
	for _, contributor := range contributors {
		if regexp.Match([]byte(contributor.Name)) || regexp.Match([]byte(contributor.Email)) {
			return true
		}
	}

	return false
}
