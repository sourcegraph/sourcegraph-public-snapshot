package codeownership

import (
	"context"
	"fmt"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ForResult(r result.FileMatch) []string {
	if r.Repo.Name != "github.com/philipp-spiess/codeowners-test" {
		return []string{}
	}

	switch r.File.Path {
	case "backend/backend-code.go":
		return []string{"@nicolasdular"}
	case "frontend/frontend-code.ts":
		return []string{"@philipp-spiess"}
	default:
		return []string{"@philipp-spiess", "@nicolasdular"}
	}
}

func ForOwner(owner string) []string {
	switch owner {
	case "@philipp-spiess":
		return []string{"frontend/frontend-code.ts", "unowned/legacy-code.php", "CODEOWNERS", "README.md"}
	case "@nicolasdular":
		return []string{"backend/backend-code.go", "unowned/legacy-code.php", "CODEOWNERS", "README.md"}
	default:
		panic("unexpected owner")
	}
}

// This returns a regex for potential filenames that are owned by the owner. The reason why this is
// not exact this is that CODEOWNERS also defines a precedence order.
//
// E.g. If a CODEOWNERS file contains a wildcard rule in the first line followed by a more specific
// rule in the next line, a file matching the more specific rule will not be matched by the
// wildcard.
//
// In the worst case, the owner we search for is part of the wildcard rule. In this case we do a
// regular search and filter the ownership information based on the other rules in post.
func PotentialFileReForOwner(_repo string, owner string) string {
	// Unfortunately in our example we have a wildcard rule with both possible owners.
	return ".*"
}

func NewFilterJob(child job.Job, fileOwnersMustInclude []string, fileOwnersMustExclude []string) job.Job {
	return &codeownershipFilterJob{
		child:                 child,
		fileOwnersMustInclude: fileOwnersMustInclude,
		fileOwnersMustExclude: fileOwnersMustExclude}
}

type codeownershipFilterJob struct {
	child job.Job

	fileOwnersMustInclude []string
	fileOwnersMustExclude []string
}

func (s *codeownershipFilterJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	var (
		_    sync.Mutex
		errs error
	)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {

		event.Results = applyCodeOwnershipFiltering(ctx, event.Results, s.fileOwnersMustInclude, s.fileOwnersMustExclude)
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (s *codeownershipFilterJob) Name() string {
	return "codeownershipFilterJob"
}

func (s *codeownershipFilterJob) Fields(job.Verbosity) []otlog.Field { return nil }

func (s *codeownershipFilterJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *codeownershipFilterJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}

func applyCodeOwnershipFiltering(ctx context.Context, matches []result.Match, fileOwnersMustInclude []string, fileOwnersMustExclude []string) []result.Match {
	// Filter matches in place
	filtered := matches[:0]

OUTER:
	for _, m := range matches {
		switch mm := m.(type) {
		case *result.FileMatch:
			fmt.Printf("event: %+v\n", mm.File)

			for _, mustIncludeOwner := range fileOwnersMustInclude {
				if !contains(ForResult(*mm), mustIncludeOwner) {
					continue OUTER
				}
			}

			filtered = append(filtered, m)
		default:
			filtered = append(filtered, m)
		}
	}

	return filtered
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
