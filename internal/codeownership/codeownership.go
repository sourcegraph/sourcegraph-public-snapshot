package codeownership

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
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
	ruleset := ResolveRuleset(db, r.Repo.Name, r.CommitID)
	return ForFilePath(ruleset, r.File.Path)
}

func ForFilePath(ruleset codeowners.Ruleset, path string) []string {
	if ruleset == nil {
		return []string{}
	}

	rule, err := ruleset.Match(path)
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

func ResolveRuleset(db database.DB, repoName api.RepoName, commitID api.CommitID) codeowners.Ruleset {
	var content []byte

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content = loadOwnershipFile(ctx, db, repoName, commitID)
	if content == nil {
		return nil
	}

	ruleSet, err := codeowners.ParseFile(strings.NewReader(string(content)))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return ruleSet
}

func loadOwnershipFile(ctx context.Context, db database.DB, repoName api.RepoName, commitID api.CommitID) []byte {
	for _, path := range []string{"CODEOWNERS", ".github/CODEOWNERS", ".gitlab/CODEOWNERS", "docs/CODEOWNERS"} {
		content, _ := gitserver.NewClient(db).ReadFile(
			ctx,
			repoName,
			commitID,
			path,
			authz.DefaultSubRepoPermsChecker,
		)

		if content != nil {
			return content
		}
	}

	return nil
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

	mux := &sync.Mutex{}
	rules := make(map[string]codeowners.Ruleset)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		// Filter matches in place
		filtered := event.Results[:0]

	OUTER:
		for _, m := range event.Results {
			switch mm := m.(type) {
			case *result.FileMatch:
				cachekey := fmt.Sprintf("%s@%s", mm.Repo.Name, mm.CommitID)

				mux.Lock()
				ruleset, ok := rules[cachekey]
				if !ok {
					rules[cachekey] = ResolveRuleset(s.db, mm.Repo.Name, mm.CommitID)
					ruleset = rules[cachekey]
				}
				if len(ruleset) == 0 {
					// If the repo has no ownership rules, it can never fullfil ownership query
					continue OUTER
				}
				owners := ForFilePath(rules[cachekey], mm.File.Path)
				mux.Unlock()

				for _, mustIncludeOwner := range s.fileOwnersMustInclude {
					if !contains(owners, mustIncludeOwner) {
						continue OUTER
					}
				}

				filtered = append(filtered, m)
			default:
				filtered = append(filtered, m)
			}
		}

		event.Results = filtered
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

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
