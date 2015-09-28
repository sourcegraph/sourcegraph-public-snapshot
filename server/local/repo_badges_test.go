package local

import (
	"net/url"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
)

func TestReposService_ListBadges(t *testing.T) {
	var s repoBadges
	ctx, mock := testContext()
	ctx = conf.WithExternalEndpoints(ctx, conf.ExternalEndpointsOpts{})
	ctx = conf.WithAppURL(ctx, &url.URL{})

	calledReposGet := mock.stores.Repos.MockGet(t, "r/r")

	badges, err := s.ListBadges(ctx, &sourcegraph.RepoSpec{URI: "r/r"})
	if err != nil {
		t.Fatal(err)
	}
	if len(badges.Badges) != len(allRepositoryBadges) {
		t.Errorf("got len(badges) == %d, want %d", len(badges.Badges), len(allRepositoryBadges))
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
}

func TestReposService_ListCounters(t *testing.T) {
	var s repoBadges
	ctx, mock := testContext()
	ctx = conf.WithExternalEndpoints(ctx, conf.ExternalEndpointsOpts{})
	ctx = conf.WithAppURL(ctx, &url.URL{})

	calledReposGet := mock.stores.Repos.MockGet(t, "r/r")

	counters, err := s.ListCounters(ctx, &sourcegraph.RepoSpec{URI: "r/r"})
	if err != nil {
		t.Fatal(err)
	}
	if len(counters.Counters) != len(allRepositoryCounters) {
		t.Errorf("got len(counters) == %d, want %d", len(counters.Counters), len(allRepositoryCounters))
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
}
