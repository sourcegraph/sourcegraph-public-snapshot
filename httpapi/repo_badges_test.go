package httpapi

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	apirouter "sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sqs/pbtypes"
)

func TestRepoBadge(t *testing.T) {
	c, mock := newTest()

	calledGet := mock.Repos.MockGet(t, "my/repo")
	calledGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")

	want := "https://img.shields.io/badge/Sourcegraph-Status-blue.svg"

	u, err := apirouter.URL(apirouter.RepoBadge, map[string]string{"Repo": "my/repo", "Badge": "status", "Format": "svg"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GetNoFollowRedirects(u.String())
	if err != nil {
		t.Fatal(err)
	}

	if want := http.StatusSeeOther; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
	if got := resp.Header.Get("location"); got != want {
		t.Errorf("got redirected to %q, want %q", got, want)
	}

	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledGetCommit {
		t.Error("!calledGetCommit")
	}
}

func TestRepoCounter_record(t *testing.T) { testRepoCounter(t, true) }

func TestRepoCounter_noRecord(t *testing.T) { testRepoCounter(t, false) }

func testRepoCounter(t *testing.T, record bool) {
	c, mock := newTest()

	calledGet := mock.Repos.MockGet(t, "my/repo")
	var calledRecordHit, calledCountHits bool
	mock.RepoBadges.RecordHit_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
		calledRecordHit = true
		return &pbtypes.Void{}, nil
	}
	mock.RepoBadges.CountHits_ = func(ctx context.Context, op *sourcegraph.RepoBadgesCountHitsOp) (*sourcegraph.RepoBadgesCountHitsResult, error) {
		calledCountHits = true
		return &sourcegraph.RepoBadgesCountHitsResult{Hits: 123}, nil
	}

	want := "https://img.shields.io/badge/views-123-blue.svg"

	u, err := apirouter.URL(apirouter.RepoCounter, map[string]string{"Repo": "my/repo", "Counter": "views", "Format": "svg"})
	if err != nil {
		t.Fatal(err)
	}
	urlStr := u.String()
	if !record {
		urlStr += "?no-record"
	}
	resp, err := c.GetNoFollowRedirects(urlStr)
	if err != nil {
		t.Fatal(err)
	}

	if want := http.StatusSeeOther; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
	if got := resp.Header.Get("location"); got != want {
		t.Errorf("got redirected to %q, want %q", got, want)
	}

	if !*calledGet {
		t.Error("!calledGet")
	}
	if calledRecordHit != record {
		t.Errorf("got calledRecordHit == %v, want %v", calledRecordHit, record)
	}
	if !calledCountHits {
		t.Error("!calledCountHits")
	}
}
