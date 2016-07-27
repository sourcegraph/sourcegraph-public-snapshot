package backend

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestBuildsService_Get(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	wantBuild := &sourcegraph.Build{ID: 1, Repo: 1}

	calledGet := mock.stores.Builds.MockGet(t, sourcegraph.BuildSpec{ID: 1, Repo: 1})

	build, err := s.Get(ctx, &sourcegraph.BuildSpec{ID: 1, Repo: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !reflect.DeepEqual(build, wantBuild) {
		t.Errorf("got %+v, want %+v", build, wantBuild)
	}
}

func TestBuildsService_List(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	wantBuilds := &sourcegraph.BuildList{Builds: []*sourcegraph.Build{{ID: 1, Repo: 1}}}

	calledList := mock.stores.Builds.MockList(t, sourcegraph.BuildSpec{ID: 1, Repo: 1})

	builds, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !reflect.DeepEqual(builds, wantBuilds) {
		t.Errorf("got %+v, want %+v", builds, wantBuilds)
	}
}

func TestBuildsService_List_pagination(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	var buildList = []*sourcegraph.Build{
		{ID: 1, Repo: 1, CreatedAt: pbtypes.NewTimestamp(time.Unix(2, 0))},
		{ID: 12, Repo: 1, CreatedAt: pbtypes.NewTimestamp(time.Unix(1, 0))},
		{ID: 123, Repo: 1, CreatedAt: pbtypes.NewTimestamp(time.Unix(0, 0))},
	}

	tests := map[string]struct {
		opt        sourcegraph.BuildListOptions
		wantBuilds *sourcegraph.BuildList
	}{
		"default pagination options, all pages, asc sort, items 1-3 of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{}},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{buildList[2], buildList[1], buildList[0]},
				StreamResponse: sourcegraph.StreamResponse{HasMore: false},
			},
		},
		"first page, 2 items/page, items 1-2 of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 1, PerPage: 2}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{buildList[0], buildList[1]},
				StreamResponse: sourcegraph.StreamResponse{HasMore: true},
			},
		},
		"second page, 2 items/page, item 3 of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 2, PerPage: 2}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{buildList[2]},
				StreamResponse: sourcegraph.StreamResponse{HasMore: false},
			},
		},
		"third page, 2 items/page, no items of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 3, PerPage: 2}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{},
				StreamResponse: sourcegraph.StreamResponse{HasMore: false},
			},
		},
		"first page, 1 item/page, item 1 of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 1, PerPage: 1}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{buildList[0]},
				StreamResponse: sourcegraph.StreamResponse{HasMore: true},
			},
		},
		"second page, 1 item/page, item 2 of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 2, PerPage: 1}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{buildList[1]},
				StreamResponse: sourcegraph.StreamResponse{HasMore: true},
			},
		},
		"third page, 1 item/page, item 3 of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 3, PerPage: 1}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{buildList[2]},
				StreamResponse: sourcegraph.StreamResponse{HasMore: false},
			},
		},
		"fourth page, 1 item/page, no items of 3": {
			opt: sourcegraph.BuildListOptions{ListOptions: sourcegraph.ListOptions{Page: 4, PerPage: 1}, Sort: "created_at", Direction: "desc"},
			wantBuilds: &sourcegraph.BuildList{
				Builds:         []*sourcegraph.Build{},
				StreamResponse: sourcegraph.StreamResponse{HasMore: false},
			},
		},
	}

	for label, test := range tests {
		calledList := mock.stores.Builds.MockList(t,
			sourcegraph.BuildSpec{ID: 1, Repo: 1},
			sourcegraph.BuildSpec{ID: 12, Repo: 1},
			sourcegraph.BuildSpec{ID: 123, Repo: 1},
		)

		opt := test.opt
		builds, err := s.List(ctx, &opt)
		if err != nil {
			t.Fatalf("%s: BuildsService List: %v", label, err)
		}
		if !*calledList {
			t.Errorf("%s: !calledList", label)
		}
		if !reflect.DeepEqual(builds, test.wantBuilds) {
			t.Errorf("%s: got %+v, want %+v", label, builds, test.wantBuilds)
		}
	}
}

func TestBuildsService_Create_MissingRevision(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	wantRepo := int32(3)
	wantCommitID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mock.servers.Repos.MockGet(t, wantRepo)
	mock.servers.Repos.MockResolveRev_NotFound(t, wantRepo, wantCommitID)

	_, err := s.Create(ctx, &sourcegraph.BuildsCreateOp{Repo: wantRepo, CommitID: wantCommitID})
	if grpc.Code(err) != codes.NotFound {
		t.Fatalf("wanted NotFound err, got %v", err)
	}
}
