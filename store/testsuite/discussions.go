package testsuite

import (
	"log"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func Discussions_Create_ok(ctx context.Context, t *testing.T, store store.Discussions, repo string) {
	for i, d := range discussionsFixture(repo) {
		want := *d
		want.ID = int64(i + 1)

		before := time.Now()
		err := store.Create(ctx, d)
		after := time.Now()
		if err != nil {
			t.Fatal(err)
		}
		// We can't guess the timestamp beforehand, so just ensure the range
		// makes sense
		want.CreatedAt = d.CreatedAt
		if !reflect.DeepEqual(d, &want) {
			log.Fatalf("Discussion creation is unwant. %#v != %#v", d, want)
		}
		if !timeInRange(before, d.CreatedAt.Time(), after) {
			log.Fatalf("CreatedAt timestamp incorrect !(%s <= %s <= %s)", before, d.CreatedAt.Time(), after)
		}
	}
}

func Discussions_Get_ok(ctx context.Context, t *testing.T, store store.Discussions, repo string) {
	discussions := discussionsFixture(repo)
	for _, d := range discussions {
		// store.Create mutates the discussion, so use a copy
		copy := *d
		err := store.Create(ctx, &copy)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Get and check the 2nd discussion
	got, err := store.Get(ctx, sourcegraph.RepoSpec{URI: repo}, 2)
	if err != nil {
		log.Fatal(err)
	}
	want := discussions[1]
	want.ID = 2
	want.CreatedAt = got.CreatedAt
	if !reflect.DeepEqual(got, want) {
		log.Fatalf("Discussion Get is unexpected. %#v != %#v", got, want)
	}
}

func Discussions_List_DefKey(ctx context.Context, t *testing.T, store store.Discussions, repo string) {
	discussions := discussionsFixture(repo)
	for _, d := range discussions {
		err := store.Create(ctx, d)
		if err != nil {
			t.Fatal(err)
		}
	}

	dss, err := store.List(ctx, &sourcegraph.DiscussionListOp{DefKey: discussions[0].DefKey})
	if err != nil {
		t.Fatal(err)
	}
	// 0 and 2 have the same DefKey (except different commits)
	want := []*sourcegraph.Discussion{discussions[0], discussions[2]}
	if !reflect.DeepEqual(dss.Discussions, want) {
		log.Fatalf("Discussion List is unwant. %#v != %#v", dss.Discussions, want)
	}
}

func Discussions_List_Repo(ctx context.Context, t *testing.T, store store.Discussions, repo string) {
	discussions := discussionsFixture(repo)
	for i := 0; i < len(discussions); i++ {
		// We want to mutate to get the correct ID
		err := store.Create(ctx, discussions[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	dss, err := store.List(ctx, &sourcegraph.DiscussionListOp{Repo: sourcegraph.RepoSpec{URI: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dss.Discussions, discussions) {
		log.Fatalf("Discussion List is unexpected. %#v != %#v", dss.Discussions, discussions)
	}
}

func discussionsFixture(repo string) []*sourcegraph.Discussion {
	return []*sourcegraph.Discussion{
		&sourcegraph.Discussion{
			Title:       "TestTitle",
			Description: "TestDescription",
			Author:      sourcegraph.UserSpec{Login: "TestUser"},
			DefKey: graph.DefKey{
				Repo:     repo,
				CommitID: "ddc58f1f46b0186cf3db28ac9f03432ade9f0783",
				UnitType: "unit",
				Unit:     "TestUnit",
				Path:     "testpackage/testfile.go",
			},
		},
		&sourcegraph.Discussion{
			Title:       "TestTitle",
			Description: "TestDescription",
			Author:      sourcegraph.UserSpec{Login: "TestUser"},
			DefKey: graph.DefKey{
				Repo:     repo,
				CommitID: "e6145aad46b985cfb4572545085d7abc835ca477",
				UnitType: "unit",
				Unit:     "TestUnit2",
				Path:     "testpackage/testfile.go",
			},
		},
		&sourcegraph.Discussion{
			Title:       "TestTitle2",
			Description: "TestDescription2",
			Author:      sourcegraph.UserSpec{Login: "TestUser2"},
			DefKey: graph.DefKey{
				Repo:     repo,
				CommitID: "e6145aad46b985cfb4572545085d7abc835ca477",
				UnitType: "unit",
				Unit:     "TestUnit",
				Path:     "testpackage/testfile.go",
			},
		},
	}
}

func timeInRange(before, x, after time.Time) bool {
	return !(x.Before(before) || x.After(after))
}
