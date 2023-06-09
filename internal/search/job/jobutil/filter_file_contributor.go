package jobutil

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewFileHasContributorsJob(child job.Job, ignoreCase bool, includeContributors, excludeContributors []string) job.Job {
	return &fileHasContributorsJob{
		child:               child,
		ignoreCase:          ignoreCase,
		includeContributors: includeContributors,
		excludeContributors: excludeContributors,
	}
}

type fileHasContributorsJob struct {
	child job.Job

	ignoreCase          bool
	includeContributors []string
	excludeContributors []string
}

func (j *fileHasContributorsJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer finish(alert, err)

	includeContributors, excludeContributors, err := j.compileRegexps()
	if err != nil {
		return nil, err
	}

	var (
		mu   sync.Mutex
		errs error
	)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		filtered := event.Results[:0]
		for _, res := range event.Results {
			// Filter out any result that is not a file
			if fm, ok := res.(*result.FileMatch); ok {
				buf := make([]byte, 1024)
				fileMatchContributors, err := getFileContributors(ctx, clients, fm, buf)
				if err != nil {
					mu.Lock()
					errs = errors.Append(errs, err)
					mu.Unlock()
					continue
				}

				if fileMatchContributors.Filtered(excludeContributors, true) ||
					fileMatchContributors.Filtered(includeContributors, false) {
					continue
				}

				filtered = append(filtered, fm)
			}
		}

		event.Results = filtered

		stream.Send(event)
	})

	return j.child.Run(ctx, clients, filteredStream)
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
			attribute.StringSlice("includeContributors", j.includeContributors),
			attribute.StringSlice("excludeContributors", j.excludeContributors),
		)
	}
	return res
}

func (j *fileHasContributorsJob) compileRegexps() (include, exclude []*casetransform.Regexp, err error) {
	include, err = j.regexps(j.includeContributors)
	if err != nil {
		return nil, nil, err
	}

	exclude, err = j.regexps(j.excludeContributors)
	if err != nil {
		return nil, nil, err
	}

	return include, exclude, nil
}

func (j *fileHasContributorsJob) regexps(filters []string) ([]*casetransform.Regexp, error) {
	var compiledFilters []*casetransform.Regexp
	for _, contributorExpression := range filters {
		re, err := casetransform.CompileRegexp(contributorExpression, j.ignoreCase)
		if err != nil {
			return nil, err
		}
		compiledFilters = append(compiledFilters, re)
	}
	return compiledFilters, nil
}

func getFileContributors(ctx context.Context, clients job.RuntimeClients, fm *result.FileMatch, buf []byte) (*FilterableContributors, error) {
	opts := gitserver.ContributorOptions{
		Range: string(fm.CommitID),
		Path:  fm.Path,
	}
	contributors, err := clients.Gitserver.ContributorCount(ctx, fm.Repo.Name, opts)

	if err != nil {
		return nil, err
	}

	return &FilterableContributors{
		Contributors: contributors,
		LowerBuf:     buf,
	}, nil
}

type FilterableContributors struct {
	Contributors []*gitdomain.ContributorCount
	LowerBuf     []byte
}

// Filtered returns true if the related match should be removed from results due to the set of provided filters.
// Filters are AND'ed together. Filters are negation filters if excludeContributors is true.
func (f *FilterableContributors) Filtered(filters []*casetransform.Regexp, excludeContributors bool) bool {
	for _, filter := range filters {
		if f.Match(filter) == excludeContributors {
			return false
		}
	}

	return true
}

func (f *FilterableContributors) Match(regexp *casetransform.Regexp) bool {
	for _, contributor := range f.Contributors {
		if regexp.Match([]byte(contributor.Name), &f.LowerBuf) || regexp.Match([]byte(contributor.Email), &f.LowerBuf) {
			return true
		}
	}

	return false
}
