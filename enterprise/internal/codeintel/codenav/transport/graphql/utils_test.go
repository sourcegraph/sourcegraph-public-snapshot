package graphql

import (
	"context"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCursor(t *testing.T) {
	expected := "test"
	pageInfo := EncodeCursor(&expected)

	if !pageInfo.HasNextPage() {
		t.Fatalf("expected next page")
	}
	if pageInfo.EndCursor() == nil {
		t.Fatalf("unexpected nil cursor")
	}

	value, err := DecodeCursor(pageInfo.EndCursor())
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected decoded cursor. want=%s have=%s", expected, value)
	}
}

func TestCursorEmpty(t *testing.T) {
	pageInfo := EncodeCursor(nil)

	if pageInfo.HasNextPage() {
		t.Errorf("unexpected next page")
	}
	if pageInfo.EndCursor() != nil {
		t.Errorf("unexpected encoded cursor: %s", *pageInfo.EndCursor())
	}

	value, err := DecodeCursor(nil)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != "" {
		t.Errorf("unexpected decoded cursor: %s", value)
	}
}

func TestResolveLocations(t *testing.T) {
	repos := database.NewStrictMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*sgtypes.Repo, error) {
		return &sgtypes.Repo{ID: id, Name: api.RepoName(fmt.Sprintf("repo%d", id))}, nil
	})

	db := database.NewStrictMockDB()
	db.ReposFunc.SetDefaultReturn(repos)

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if spec == "deadbeef3" {
			return "", &gitdomain.RevisionNotFoundError{}
		}
		return api.CommitID(spec), nil
	})

	r1 := types.Range{Start: types.Position{Line: 11, Character: 12}, End: types.Position{Line: 13, Character: 14}}
	r2 := types.Range{Start: types.Position{Line: 21, Character: 22}, End: types.Position{Line: 23, Character: 24}}
	r3 := types.Range{Start: types.Position{Line: 31, Character: 32}, End: types.Position{Line: 33, Character: 34}}
	r4 := types.Range{Start: types.Position{Line: 41, Character: 42}, End: types.Position{Line: 43, Character: 44}}

	locations, err := resolveLocations(context.Background(), sharedresolvers.NewCachedLocationResolver(db, gsClient), []types.UploadLocation{
		{Dump: types.Dump{RepositoryID: 50}, TargetCommit: "deadbeef1", TargetRange: r1, Path: "p1"},
		{Dump: types.Dump{RepositoryID: 51}, TargetCommit: "deadbeef2", TargetRange: r2, Path: "p2"},
		{Dump: types.Dump{RepositoryID: 52}, TargetCommit: "deadbeef3", TargetRange: r3, Path: "p3"},
		{Dump: types.Dump{RepositoryID: 53}, TargetCommit: "deadbeef4", TargetRange: r4, Path: "p4"},
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	mockrequire.Called(t, repos.GetFunc)

	if len(locations) != 3 {
		t.Fatalf("unexpected length. want=%d have=%d", 3, len(locations))
	}
	if url := locations[0].CanonicalURL(); url != "/repo50@deadbeef1/-/blob/p1?L12:13-14:15" {
		t.Errorf("unexpected canonical url. want=%s have=%s", "/repo50@deadbeef1/-/blob/p1?L12:13-14:15", url)
	}
	if url := locations[1].CanonicalURL(); url != "/repo51@deadbeef2/-/blob/p2?L22:23-24:25" {
		t.Errorf("unexpected canonical url. want=%s have=%s", "/repo51@deadbeef2/-/blob/p2?L22:23-24:25", url)
	}
	if url := locations[2].CanonicalURL(); url != "/repo53@deadbeef4/-/blob/p4?L42:43-44:45" {
		t.Errorf("unexpected canonical url. want=%s have=%s", "/repo53@deadbeef4/-/blob/p4?L42:43-44:45", url)
	}
}
