package backend

import (
	"crypto/rand"
	"html"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestDeltasService_ListFiles(t *testing.T) {
	var s deltas
	ctx, _ := testContext()

	ds := sourcegraph.DeltaSpec{
		Base: sourcegraph.RepoRevSpec{Repo: 1, CommitID: "basecommit"},
		Head: sourcegraph.RepoRevSpec{Repo: 2, CommitID: "headcommit"},
	}

	fdiffs := []*diff.FileDiff{
		{
			OrigName: "f",
			NewName:  "f",
			Extended: []string{
				"diff --git f f",
				"index basecommit..headcommit 100644",
			},
			Hunks: []*diff.Hunk{
				{
					OrigStartLine: 1,
					OrigLines:     1,
					NewStartLine:  1,
					NewLines:      1,
					StartPosition: 1,
					Body:          []byte("-a\n+b\n"),
				},
			},
		},
	}

	var calledDiff bool
	s.mockDiffFunc = func(ctx context.Context, ds sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error) {
		calledDiff = true
		return fdiffs, nil, nil
	}

	dfs, err := s.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{Ds: ds})
	if err != nil {
		t.Fatal(err)
	}

	want := fdiffs[0]
	got := &dfs.FileDiffs[0].FileDiff
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\n\nwant\n%+v", got, want)
	}

	if !calledDiff {
		t.Error("!calledDiff")
	}
}

func TestDeltasService_ListFiles_Escaped(t *testing.T) {
	var s deltas
	ctx, _ := testContext()

	ds := sourcegraph.DeltaSpec{
		Base: sourcegraph.RepoRevSpec{Repo: 1, CommitID: "basecommit"},
		Head: sourcegraph.RepoRevSpec{Repo: 2, CommitID: "headcommit"},
	}

	fdiffs := []*diff.FileDiff{
		{
			OrigName: "f",
			NewName:  "f",
			Extended: []string{
				"diff --git f f",
				"index basecommit..headcommit 100644",
			},
			Hunks: []*diff.Hunk{
				{
					OrigStartLine: 1,
					OrigLines:     1,
					NewStartLine:  1,
					NewLines:      1,
					StartPosition: 1,
					Body:          []byte(html.EscapeString("-<div>what</div>\n+<div>no way</div>\n")),
				},
			},
		},
	}

	var calledDiff bool
	s.mockDiffFunc = func(ctx context.Context, ds sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error) {
		calledDiff = true
		return fdiffs, nil, nil
	}

	dfs, err := s.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{Ds: ds})
	if err != nil {
		t.Fatal(err)
	}

	want := fdiffs[0]
	got := &dfs.FileDiffs[0].FileDiff
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\n\nwant\n%+v", got, want)
	}

	if !calledDiff {
		t.Error("!calledDiff")
	}
}

func TestDeltasListFilesCacheKeyDeterministic(t *testing.T) {
	op := new(sourcegraph.DeltasListFilesOp)
	op.Ds.Base.Repo = 1
	op.Ds.Base.CommitID = "base-commit"
	op.Ds.Head.Repo = 2
	op.Ds.Head.CommitID = "head-commit"
	op.Opt = new(sourcegraph.DeltaListFilesOptions)
	op.Opt.Filter = "filter"
	op.Opt.Ignore = []string{"not", "this"}
	k, err := deltasListFileCacheKey(op)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		x, err := deltasListFileCacheKey(op)
		if err != nil {
			t.Fatal(err)
		}
		if k != x {
			t.Errorf("Deltas list files cache key is not determinstic got %q, expected %q", x, k)
		}
	}
}

func TestDeltaListFilesCache(t *testing.T) {
	cache := newDeltasListFilesCache(1, 10*1024)
	files := new(sourcegraph.DeltaFiles)
	files.Stats.Changed = 1111

	op := new(sourcegraph.DeltasListFilesOp)
	op.Ds.Base.Repo = 1
	cache.Add(op, files)
	hit, ok := cache.Get(op)
	if !ok {
		t.Fatal("Delta files not cached")
	}
	if hit.Stats.Changed != 1111 {
		t.Errorf("Delta files cached entry does not match, got %d lines changed", hit.Stats.Changed)
	}

	// Test eviction.
	other := new(sourcegraph.DeltasListFilesOp)
	other.Ds.Base.Repo = 2
	cache.Add(other, files)
	_, ok = cache.Get(op)
	if ok {
		t.Error("Delta files should not be cached")
	}

	// Test over max entry size.
	fluff := make([]byte, 10*1024)
	_, err := rand.Read(fluff)
	if err != nil {
		t.Fatal(err)
	}
	large := new(sourcegraph.DeltaFiles)
	large.FileDiffs = []*sourcegraph.FileDiff{&sourcegraph.FileDiff{PreImage: string(fluff)}}
	cache.Add(op, large)
	_, ok = cache.Get(op)
	if ok {
		t.Error("Delta files over the maximum entry size should not be cached")
	}
}
