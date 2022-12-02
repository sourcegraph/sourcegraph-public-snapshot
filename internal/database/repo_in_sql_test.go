package database

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoFiles_PutRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	repo := mustCreate(ctx, t, db, &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "name",
		Private:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	})
	// vv := types.RepoVersion{
	// 	RepoID:     repo.ID,
	// 	ExternalID: "pretend this is a git sha",
	// 	PathCoverage: types.RepoVersionPathCoverage{
	// 		PathColor: 1,
	// 		PathIndex: 1,
	// 	},
	// 	Reachability: map[int]int{1: 1},
	// }

	// v, err := db.RepoVersions().CreateIfNotExists(ctx, vv)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// d, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "dir")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// cID, err := db.RepoFileContents().Create(ctx, "content")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// ff := types.RepoFile{
	// 	DirectoryID:      d.ID,
	// 	VersionID:        v.ID,
	// 	TopologicalOrder: 1, // we need to compute this
	// 	BaseName:         "file",
	// 	ContentID:        cID,
	// }
	// _, err = db.RepoFiles().CreateIfNotExists(ctx, ff)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// Git objects storer based on memory
	s := memory.NewStorage()

	// Clones the repository into the worktree (fs) and stores all the .git
	// content into the storer
	gitRepo, err := git.Clone(s, nil, &git.CloneOptions{
		URL: "https://github.com/git-fixtures/basic.git",
	})
	if err != nil {
		t.Fatal(err)
	}
	gitBranch, err := gitRepo.Branch("master")
	if err != nil {
		t.Fatal(err)
	}
	gitBranchRef, err := storer.ResolveReference(s, gitBranch.Merge)
	if err != nil {
		t.Fatal(err)
	}
	iter, err := gitRepo.Log(&git.LogOptions{From: gitBranchRef.Hash()})
	if err != nil {
		t.Fatal(err)
	}
	children := map[plumbing.Hash][]plumbing.Hash{}
	var roots []plumbing.Hash
	err = iter.ForEach(func(c *object.Commit) error {
		for _, h := range c.ParentHashes {
			ch := children[h]
			ch = append(ch, c.Hash)
			children[h] = ch
		}
		if len(c.ParentHashes) == 0 {
			roots = append(roots, c.Hash)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) == 0 {
		t.Error("want at least one root")
	}
	if len(children) == 0 {
		t.Error("want at least one children entry")
	}
	// At this point we have a reversed parent relationship
	// and to insert commits in order we are going to:
	// 1. start with the root nodes
	// 2. process element from the stack
	// 3. if all parents of a child were processed, push the child onto a stack
	// 4. otherwise this parent was processed so we'll get to this child through another route
	var stack []*object.Commit
	for _, r := range roots {
		gitCommit, err := gitRepo.CommitObject(r)
		if err != nil {
			t.Fatal(err)
		}
		stack = append(stack, gitCommit)
	}
	processed := map[plumbing.Hash]bool{}
	for len(stack) != 0 {
		// pop
		gitCommit := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		vv := types.RepoVersion{
			RepoID:     repo.ID,
			ExternalID: hex.EncodeToString(gitCommit.Hash[:]),
			// TODO
			PathCoverage: types.RepoVersionPathCoverage{
				PathColor: 1,
				PathIndex: 1,
			},
			Reachability: map[int]int{1: 1},
		}
		_, err = db.RepoVersions().CreateIfNotExists(ctx, vv)
		if err != nil {
			t.Fatal(err)
		}
		if gitCommit.NumParents() == 0 {
			func() {
				files, err := gitCommit.Files()
				if err != nil {
					t.Fatal(err)
				}
				defer files.Close()
				err = files.ForEach(func(f *object.File) error {
					t.Logf("%s + %s", hex.EncodeToString(gitCommit.Hash[:]), f.Name)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}()
		} else {
			t.Logf("%s need to compute diff", hex.EncodeToString(gitCommit.Hash[:]))
		}
		processed[gitCommit.Hash] = true
		for _, ch := range children[gitCommit.Hash] {
			if processed[ch] {
				continue
			}
			chCommit, err := gitRepo.CommitObject(ch)
			if err != nil {
				t.Fatal(err)
			}
			chParentsProcessed := true
			for _, ph := range chCommit.ParentHashes {
				if !processed[ph] {
					chParentsProcessed = false
				}
			}
			if chParentsProcessed {
				stack = append(stack, chCommit)
			}
		}
	}
}
