package local

import (
	"fmt"
	"time"

	"strings"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	approuter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/store"
)

var RepoBadges sourcegraph.RepoBadgesServer = &repoBadges{}

type repoBadges struct{}

var _ sourcegraph.RepoBadgesServer = (*repoBadges)(nil)

var allRepositoryBadges = []sourcegraph.Badge{
	{Name: "docs-examples", Description: "Links to documentation and examples for your repository, and displays the sum of public/exported definitions in this repository and third-party usage examples of this repository"},
	{Name: "dependencies", Description: "Counts the number of repositories that this repository depends on"},
	{Name: "status", Description: "Indicates whether Sourcegraph successfully processed this repository"},
	{Name: "funcs", Description: "Counts the number of public/exported functions defined in this repository"},
	{Name: "top-func", Description: "Displays the function defined in this repository that is most frequently called by code in other repositories"},
	{Name: "library-users", Description: "Counts the number of people who refer to this repository's definitions from code in other repositories"},
	{Name: "dependents", Description: "Displays the number of repositories that depend on this repository"},
	{Name: "authors", Description: "Counts the number of people who have contributed to this repository"},
	{Name: "xrefs", Description: "Counts the number of references to this repository's definitions (functions, classes, etc.) from code in other repositories"},
}

func (s *repoBadges) ListBadges(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.BadgeList, error) {
	if _, err := store.ReposFromContext(ctx).Get(ctx, repo.URI); err != nil {
		return nil, err
	}

	var badges []*sourcegraph.Badge
	for _, b := range allRepositoryBadges {
		imageURL, err := approuter.Rel.URLToOrError(
			approuter.RepoBadge,
			"Repo", repo.URI, "Badge", b.Name, "Format", "svg",
		)
		if err != nil {
			return nil, err
		}
		imageURL = conf.AppURL(ctx).ResolveReference(imageURL)

		repoURL := conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(repo.URI))
		b.ImageURL = imageURL.String()
		b.UncountedImageURL = b.ImageURL + "?no-record=1"
		b.Markdown = fmt.Sprintf(`[![%s](%s)](%s)`, strings.Replace(b.Name, "-", " ", -1), b.ImageURL, repoURL)
		badges = append(badges, &b)
	}
	return &sourcegraph.BadgeList{Badges: badges}, nil
}

var allRepositoryCounters = []sourcegraph.Counter{
	{Name: "views", Description: "Total views"},
	{Name: "views-24h", Description: "Views in the last 24 hours"},
}

func (s *repoBadges) ListCounters(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.CounterList, error) {
	if _, err := store.ReposFromContext(ctx).Get(ctx, repo.URI); err != nil {
		return nil, err
	}

	var counters []*sourcegraph.Counter
	for _, c := range allRepositoryCounters {
		imageURL, err := approuter.Rel.URLToOrError(
			approuter.RepoCounter,
			"Repo", repo.URI, "Counter", c.Name, "Format", "svg",
		)
		if err != nil {
			return nil, err
		}
		imageURL = conf.AppURL(ctx).ResolveReference(imageURL)

		repoURL := conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(repo.URI))
		c.ImageURL = imageURL.String()
		c.UncountedImageURL = c.ImageURL + "?no-record=1"
		c.Markdown = fmt.Sprintf(`[![%s](%s)](%s)`, strings.Replace(c.Name, "-", " ", -1), c.ImageURL, repoURL)
		counters = append(counters, &c)
	}
	return &sourcegraph.CounterList{Counters: counters}, nil
}

func (s *repoBadges) RecordHit(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	return &pbtypes.Void{}, store.RepoCountersFromContext(ctx).RecordHit(ctx, repo.URI)
}

func (s *repoBadges) CountHits(ctx context.Context, op *sourcegraph.RepoBadgesCountHitsOp) (*sourcegraph.RepoBadgesCountHitsResult, error) {
	var since time.Time
	if op.Since != nil {
		since = op.Since.Time()
	}
	hits, err := store.RepoCountersFromContext(ctx).CountHits(ctx, op.Repo.URI, since)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.RepoBadgesCountHitsResult{Hits: int32(hits)}, nil
}
