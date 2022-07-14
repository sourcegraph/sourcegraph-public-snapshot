package codeownership

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/hmarr/codeowners"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ForResult(r result.FileMatch, db database.DB) []string {
	codeownersFile := ResolveCodeownerFile(db, r.Repo.Name, r.CommitID)

	if codeownersFile == "" {
		return []string{}
	}

	return ForFilePath(codeownersFile, r.File.Path)
}

func ForFilePath(ownersFile string, path string) []string {
	var err error

	if ownersFile == "" {
		return []string{}
	}

	ruleSet, err := codeowners.ParseFile(strings.NewReader(ownersFile))
	if err != nil {
		log.Fatal(err)
		return []string{}
	}

	rule, err := ruleSet.Match(path)
	if err != nil {
		log.Fatal(err)
		return []string{}
	}

	if rule == nil {
		return []string{}
	}

	owners := make([]string, len(rule.Owners))
	for i, owner := range rule.Owners {
		owners[i] = owner.String()
	}

	return owners
}

func ResolveCodeownerFile(db database.DB, repoName api.RepoName, commitID api.CommitID) string {
	var content []byte

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, _ = gitserver.NewClient(db).ReadFile(
		ctx,
		repoName,
		commitID,
		"CODEOWNERS",
		authz.DefaultSubRepoPermsChecker,
	)

	if content != nil {
		return string(content)
	}

	content, _ = gitserver.NewClient(db).ReadFile(
		ctx,
		repoName,
		commitID,
		".github/CODEOWNERS",
		authz.DefaultSubRepoPermsChecker,
	)

	if content != nil {
		return string(content)
	}

	return ""
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

func NewFilterJob(child job.Job, db database.DB, fileOwnersMustInclude []string, fileOwnersMustExclude []string) job.Job {
	return &codeownershipFilterJob{
		child:                 child,
		db:                    db,
		fileOwnersMustInclude: fileOwnersMustInclude,
		fileOwnersMustExclude: fileOwnersMustExclude}
}

type codeownershipFilterJob struct {
	child job.Job

	db database.DB

	fileOwnersMustInclude []string
	fileOwnersMustExclude []string
}

func (s *codeownershipFilterJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	var errs error

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		event.Results = applyCodeOwnershipFiltering(ctx, event.Results, s.db, s.fileOwnersMustInclude, s.fileOwnersMustExclude)
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

func applyCodeOwnershipFiltering(ctx context.Context, matches []result.Match, db database.DB, fileOwnersMustInclude []string, fileOwnersMustExclude []string) []result.Match {
	// Filter matches in place
	filtered := matches[:0]

OUTER:
	for _, m := range matches {
		switch mm := m.(type) {
		case *result.FileMatch:
			fmt.Printf("event: %+v\n", mm.File)

			for _, mustIncludeOwner := range fileOwnersMustInclude {
				if !contains(ForResult(*mm, db), mustIncludeOwner) {
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
