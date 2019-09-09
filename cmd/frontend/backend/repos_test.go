package backend

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/util"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepo := &types.Repo{ID: 1, Name: "github.com/u/r"}

	calledGet := db.Mocks.Repos.MockGet_Return(t, wantRepo)

	repo, err := s.Get(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	// Should not be called because mock GitHub has same data as mock DB.
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_List(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepos := []*types.Repo{
		{Name: "r1"},
		{Name: "r2"},
	}

	calledList := db.Mocks.Repos.MockList(t, "r1", "r2")

	repos, err := s.List(ctx, db.ReposListOptions{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !reflect.DeepEqual(repos, wantRepos) {
		t.Errorf("got %+v, want %+v", repos, wantRepos)
	}
}

func TestRepos_Add(t *testing.T) {
	var s repos
	ctx := testContext()

	const repoName = "my/repo"

	calledRepoLookup := false
	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		calledRepoLookup = true
		if args.Repo != repoName {
			t.Errorf("got %q, want %q", args.Repo, repoName)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Name: repoName, Description: "d"},
		}, nil
	}
	defer func() { repoupdater.MockRepoLookup = nil }()

	calledUpsert := false
	db.Mocks.Repos.Upsert = func(op api.InsertRepoOp) error {
		calledUpsert = true
		if want := (api.InsertRepoOp{Name: repoName, Description: "d"}); !reflect.DeepEqual(op, want) {
			t.Errorf("got %+v, want %+v", op, want)
		}
		return nil
	}

	if err := s.AddGitHubDotComRepository(ctx, repoName); err != nil {
		t.Fatal(err)
	}
	if !calledRepoLookup {
		t.Error("!calledRepoLookup")
	}
	if !calledUpsert {
		t.Error("!calledUpsert")
	}
}

type gitObjectInfo string

func (oid gitObjectInfo) OID() git.OID {
	var v git.OID
	copy(v[:], []byte(oid))
	return v
}

func TestReposGetInventory(t *testing.T) {
	var s repos
	ctx := testContext()

	const (
		wantRepo     = "a"
		wantCommitID = "cccccccccccccccccccccccccccccccccccccccc"
		wantRootOID  = "oid-root"
	)
	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		if args.Repo != wantRepo {
			t.Errorf("got %q, want %q", args.Repo, wantRepo)
		}
		return &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{Name: wantRepo}}, nil
	}
	defer func() { repoupdater.MockRepoLookup = nil }()
	git.Mocks.Stat = func(commit api.CommitID, path string) (os.FileInfo, error) {
		if commit != wantCommitID {
			t.Errorf("got commit %q, want %q", commit, wantCommitID)
		}
		return &util.FileInfo{Name_: path, Mode_: os.ModeDir, Sys_: gitObjectInfo(wantRootOID)}, nil
	}
	git.Mocks.ReadDir = func(commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
		if commit != wantCommitID {
			t.Errorf("got commit %q, want %q", commit, wantCommitID)
		}
		switch name {
		case "":
			return []os.FileInfo{
				&util.FileInfo{Name_: "a", Mode_: os.ModeDir, Sys_: gitObjectInfo("oid-a")},
				&util.FileInfo{Name_: "b.go", Size_: 12},
			}, nil
		case "a":
			return []os.FileInfo{&util.FileInfo{Name_: "a/c.m", Size_: 24}}, nil
		default:
			panic("unhandled mock ReadDir " + name)
		}
	}
	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if commit != wantCommitID {
			t.Errorf("got commit %q, want %q", commit, wantCommitID)
		}
		switch name {
		case "b.go":
			return []byte("package main"), nil
		case "a/c.m":
			return []byte("@interface X:NSObject {}"), nil
		default:
			panic("unhandled mock ReadFile " + name)
		}
	}
	defer git.ResetMocks()

	tests := []struct {
		useEnhancedLanguageDetection bool
		want                         *inventory.Inventory
	}{
		{
			useEnhancedLanguageDetection: false,
			want: &inventory.Inventory{
				Languages: []inventory.Lang{
					{Name: "MATLAB", TotalBytes: 24}, // obviously incorrect, but this is how the pre-enhanced lang detection worked
					{Name: "Go", TotalBytes: 12},
				},
			},
		},
		{
			useEnhancedLanguageDetection: true,
			want: &inventory.Inventory{
				Languages: []inventory.Lang{
					{Name: "Objective-C", TotalBytes: 24},
					{Name: "Go", TotalBytes: 12},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("useEnhancedLanguageDetection=%v", test.useEnhancedLanguageDetection), func(t *testing.T) {
			rcache.SetupForTest(t)
			orig := useEnhancedLanguageDetection
			useEnhancedLanguageDetection = test.useEnhancedLanguageDetection
			defer func() { useEnhancedLanguageDetection = orig }() // reset

			inv, err := s.GetInventory(ctx, &types.Repo{Name: wantRepo}, wantCommitID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(inv, test.want) {
				t.Errorf("got  %#v\nwant %#v", inv, test.want)
			}
		})
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}
