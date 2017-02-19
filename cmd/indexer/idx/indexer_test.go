package idx

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

func Test_queueWithoutDuplicates(t *testing.T) {
	enqueue, dequeue := queueWithoutDuplicates(prometheus.NewGauge(prometheus.GaugeOpts{}))
	doDequeue := func() string {
		c := make(chan string)
		dequeue <- c
		return <-c
	}

	enqueue <- "foo"
	enqueue <- "bar"
	enqueue <- "foo"
	enqueue <- "baz"

	if doDequeue() != "foo" {
		t.Fail()
	}
	if doDequeue() != "bar" {
		t.Fail()
	}
	if doDequeue() != "baz" {
		t.Fail()
	}
}

func Test_index_java(t *testing.T) {
	// test input data
	inputRepo := "github.com/sourcegraph/testname123"

	// mock data
	var inputRepoID int32 = 123
	depIDs := [2]string{"foo:bar", "blah:baz"}
	depRepos := [2]string{"github.com/foo/bar", "github.com/blah/baz"}
	depQueries := map[string]string{depIDs[0]: depRepos[0], depIDs[1]: depRepos[1]}
	repos := map[string]*sourcegraph.Repo{
		inputRepo: &sourcegraph.Repo{
			ID:            inputRepoID,
			URI:           inputRepo,
			DefaultBranch: "master",
		},
		depRepos[0]: &sourcegraph.Repo{
			ID:            234,
			URI:           depRepos[0],
			DefaultBranch: "master",
		},
		depRepos[1]: &sourcegraph.Repo{
			ID:            345,
			URI:           depRepos[1],
			DefaultBranch: "master",
		},
	}
	repoRevs := map[string]vcs.CommitID{
		inputRepo:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		depRepos[0]: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		depRepos[1]: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	// expected output data
	expEnqueuedRepos := depRepos[:]
	sort.Strings(expEnqueuedRepos)

	wq := NewQueue(nil)
	ctx := accesscontrol.WithInsecureSkip(context.Background(), true)
	testSvc := svcMock{}
	testSvc.ListPackages_ = func(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error) {
		return nil, nil
	}
	testSvc.GetInventoryUncached_ = func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
		if !reflect.DeepEqual(repoRev, &sourcegraph.RepoRevSpec{Repo: inputRepoID, CommitID: string(repoRevs[inputRepo])}) {
			t.Fatalf("GetInventoryUncached: unexpected repoRev param %+v", repoRev)
		}
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Java", Type: "programming"}}}, nil
	}
	updateCalled := false
	testSvc.Update_ = func(ctx context.Context, op *sourcegraph.ReposUpdateOp) (err error) {
		updateCalled = true
		return nil
	}
	testSvc.GetByURI_ = func(ctx context.Context, uri string) (res *sourcegraph.Repo, err error) {
		return &sourcegraph.Repo{
			URI: uri,
			ID:  inputRepoID,
		}, nil
	}
	testSvc.Dependencies_ = func(ctx context.Context, repoID int32, excludePrivate bool) ([]*sourcegraph.DependencyReference, error) {
		if repoID != inputRepoID {
			t.Fatalf("Dependencies: unexpected repoID param %d", repoID)
		}
		return []*sourcegraph.DependencyReference{{
			RepoID:  repoID,
			DepData: map[string]interface{}{"id": depIDs[0]},
		}, {
			RepoID:  repoID,
			DepData: map[string]interface{}{"id": depIDs[1]},
		}}, nil
	}
	testSvc.ResolveRepo_ = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		rp, ok := repos[uri]
		if !ok {
			t.Fatalf("ResolveRepo: uri %q not in mocked set", uri)
		}
		return rp, nil
	}
	testSvc.ResolveRevision_ = func(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error) {
		if spec != "HEAD" {
			t.Fatalf(`ResolveRevision: expected spec to be "HEAD", was %q`, spec)
		}
		return repoRevs[repo.URI], nil
	}
	defsRefreshIndexCalled := false
	testSvc.DefsRefreshIndex_ = func(ctx context.Context, repoURI, commit string) (err error) {
		if repoURI != inputRepo {
			t.Fatalf(`DefsRefreshIndex: expected repoURI to be %q, was %q`, inputRepo, repoURI)
		}
		defsRefreshIndexCalled = true
		return nil
	}
	pkgsRefreshIndexCalled := false
	testSvc.PkgsRefreshIndex_ = func(ctx context.Context, repo string, commit string) (err error) {
		if repo != inputRepo {
			t.Fatalf(`PkgsRefreshIndex: expected repoURI to be %q, was %q`, inputRepo, repo)
		}
		pkgsRefreshIndexCalled = true
		return nil
	}
	testSvc.GoogleGitHub_ = func(query string) (string, error) {
		rp, ok := depQueries[query]
		if !ok {
			t.Fatalf("GoogleGitHub: query %q unmocked", query)
		}
		return rp, nil
	}

	err := index(ctx, wq, testSvc, inputRepo)
	if err != nil {
		t.Fatal(err)
	}

	if !defsRefreshIndexCalled {
		t.Fatal("!defsRefreshIndexCalled")
	}
	if !updateCalled {
		t.Fatal("!updateCalled")
	}
	if !pkgsRefreshIndexCalled {
		t.Fatal("!pkgsRefreshIndexCalled")
	}

	// Collect queue
	var enqueued []string
	for range expEnqueuedRepos {
		c := make(chan string)
		wq.dequeue <- c
		select {
		case foundRepo := <-c:
			enqueued = append(enqueued, foundRepo)
		case <-time.After(time.Second * 5):
			t.Fatal("timed out waiting for enqueued repository")
		}
	}
	sort.Strings(enqueued)
	if !reflect.DeepEqual(expEnqueuedRepos, enqueued) {
		t.Errorf("after one indexing pass, expected queue to contain %+v, but found %+v", expEnqueuedRepos, enqueued)
	}
}
