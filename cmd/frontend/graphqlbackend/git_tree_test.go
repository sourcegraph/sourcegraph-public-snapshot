package graphqlbackend

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGitTree_History(t *testing.T) {
	ctx := context.Background()
	gs := gitserver.NewMockClientFrom(gitserver.NewTestClient(t))
	gs.ResolveRevisionFunc.SetDefaultReturn("deadbeef", nil)
	gs.ReadDirFunc.SetDefaultReturn(gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
		&fileutil.FileInfo{Name_: "file1"},
	}), nil)
	gs.CommitsFunc.SetDefaultReturn([]*gitdomain.Commit{
		{ID: "deadbeef"},
	}, nil)
	db := dbmocks.NewMockDB()

	rr := NewRepositoryResolver(db, gs, &types.Repo{Name: "repo"})
	gcr := NewGitCommitResolver(db, gs, rr, "deadbeef", nil)

	tree, err := gcr.Tree(ctx, &TreeArgs{Path: ""})
	require.NoError(t, err)

	entries, err := tree.Entries(ctx, &gitTreeEntryConnectionArgs{})
	require.NoError(t, err)
	require.Len(t, entries, 1)

	historyNodes, err := entries[0].
		History(ctx, HistoryArgs{}).
		Nodes(ctx)
	require.NoError(t, err)
	require.Len(t, historyNodes, 1)
}

func TestGitTree_Entries(t *testing.T) {
	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	wantPath := ""

	gitserverClient.ReadDirFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recursive bool) (gitserver.ReadDirIterator, error) {
		switch path {
		case "", ".", "/":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo(".aspect/", true),
					CreateFileInfo(".aspect/rules/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
					CreateFileInfo(".aspect/cli/", true),
					CreateFileInfo(".aspect/cli/file1", false),
					CreateFileInfo(".aspect/cli/file2", false),
					CreateFileInfo("folder/", true),
					CreateFileInfo("folder/nestedfile", false),
					CreateFileInfo("folder/subfolder/", true),
					CreateFileInfo("folder/subfolder/deeplynestedfile", false),
					CreateFileInfo("folder/subfolder2/", true),
					CreateFileInfo("folder/subfolder2/file", false),
					CreateFileInfo("folder/subfolder2/file2", false),
					CreateFileInfo("folder2/", true),
					CreateFileInfo("folder2/file", false),
					CreateFileInfo("file", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo(".aspect/", true),
				CreateFileInfo("folder/", true),
				CreateFileInfo("folder2/", true),
				CreateFileInfo("file", false),
			}), nil
		case "folder/", "folder":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo("folder/", true),
					CreateFileInfo("folder/nestedfile", false),
					CreateFileInfo("folder/subfolder/", true),
					CreateFileInfo("folder/subfolder/deeplynestedfile", false),
					CreateFileInfo("folder/subfolder2/", true),
					CreateFileInfo("folder/subfolder2/file", false),
					CreateFileInfo("folder/subfolder2/file2", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo("folder/nestedfile", false),
				CreateFileInfo("folder/subfolder/", true),
				CreateFileInfo("folder/subfolder2/", true),
			}), nil
		case ".aspect/", ".aspect":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo(".aspect/", true),
					CreateFileInfo(".aspect/rules/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
					CreateFileInfo(".aspect/cli/", true),
					CreateFileInfo(".aspect/cli/file1", false),
					CreateFileInfo(".aspect/cli/file2", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo(".aspect/rules/", true),
				CreateFileInfo(".aspect/cli/", true),
			}), nil
		case ".aspect/rules/", ".aspect/rules":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo(".aspect/", true),
					CreateFileInfo(".aspect/rules/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
			}), nil
		case ".aspect/rules/external_repository_action_cache/", ".aspect/rules/external_repository_action_cache":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo(".aspect/", true),
					CreateFileInfo(".aspect/rules/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
					CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
			}), nil
		case ".aspect/cli/", ".aspect/cli":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo(".aspect/", true),
					CreateFileInfo(".aspect/cli/", true),
					CreateFileInfo(".aspect/cli/file1", false),
					CreateFileInfo(".aspect/cli/file2", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo(".aspect/cli/file1", false),
				CreateFileInfo(".aspect/cli/file2", false),
			}), nil
		case "folder/subfolder/", "folder/subfolder":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo("folder/", true),
					CreateFileInfo("folder/subfolder/", true),
					CreateFileInfo("folder/subfolder/deeplynestedfile", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo("folder/subfolder/deeplynestedfile", false),
			}), nil
		case "folder/subfolder2/", "folder/subfolder2":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo("folder/", true),
					CreateFileInfo("folder/subfolder2/", true),
					CreateFileInfo("folder/subfolder2/file", false),
					CreateFileInfo("folder/subfolder2/file2", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo("folder/subfolder2/file", false),
				CreateFileInfo("folder/subfolder2/file2", false),
			}), nil
		case "folder2/", "folder2":
			if recursive {
				return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
					CreateFileInfo("folder2/", true),
					CreateFileInfo("folder2/file", false),
				}), nil
			}
			return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
				CreateFileInfo("folder2/file", false),
			}), nil
		default:
			return nil, errors.Newf("bad argument %q", path)
		}
	})

	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		Stat: CreateFileInfo(wantPath, true),
	}
	gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

	// Plain list all root entries.
	t.Run("root", func(t *testing.T) {
		entries, err := gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
			CreateFileInfo("folder/", true),
			CreateFileInfo("folder2/", true),
			CreateFileInfo("file", false),
		}, entries)
		entries, err = gitTree.Files(context.Background(), &gitTreeEntryConnectionArgs{})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo("file", false),
		}, entries)
		entries, err = gitTree.Directories(context.Background(), &gitTreeEntryConnectionArgs{})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
			CreateFileInfo("folder/", true),
			CreateFileInfo("folder2/", true),
		}, entries)
	})

	t.Run("Subfolder", func(t *testing.T) {
		opts := GitTreeEntryResolverOpts{
			Commit: &GitCommitResolver{
				repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
			},
			Stat: CreateFileInfo("folder/", true),
		}
		gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

		entries, err := gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo("folder/subfolder/", true),
			CreateFileInfo("folder/subfolder2/", true),
			CreateFileInfo("folder/nestedfile", false),
		}, entries)
		entries, err = gitTree.Files(context.Background(), &gitTreeEntryConnectionArgs{})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo("folder/nestedfile", false),
		}, entries)
		entries, err = gitTree.Directories(context.Background(), &gitTreeEntryConnectionArgs{})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo("folder/subfolder/", true),
			CreateFileInfo("folder/subfolder2/", true),
		}, entries)
	})

	t.Run("Pagination", func(t *testing.T) {
		entries, err := gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{ConnectionArgs: gqlutil.ConnectionArgs{First: pointers.Ptr(int32(1))}})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
		}, entries)
		entries, err = gitTree.Files(context.Background(), &gitTreeEntryConnectionArgs{ConnectionArgs: gqlutil.ConnectionArgs{First: pointers.Ptr(int32(1))}})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo("file", false),
		}, entries)
		entries, err = gitTree.Directories(context.Background(), &gitTreeEntryConnectionArgs{ConnectionArgs: gqlutil.ConnectionArgs{First: pointers.Ptr(int32(1))}})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
		}, entries)

		// Invalid first.
		_, err = gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{ConnectionArgs: gqlutil.ConnectionArgs{First: pointers.Ptr(int32(-1))}})
		require.Error(t, err)

		// First is bigger than the number of entries.
		entries, err = gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{ConnectionArgs: gqlutil.ConnectionArgs{First: pointers.Ptr(int32(100))}})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
			CreateFileInfo("folder/", true),
			CreateFileInfo("folder2/", true),
			CreateFileInfo("file", false),
		}, entries)
	})

	t.Run("Recursive", func(t *testing.T) {
		entries, err := gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{Recursive: true})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
			CreateFileInfo(".aspect/cli/", true),
			CreateFileInfo(".aspect/rules/", true),
			CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
			CreateFileInfo("folder/", true),
			CreateFileInfo("folder/subfolder/", true),
			CreateFileInfo("folder/subfolder2/", true),
			CreateFileInfo("folder2/", true),
			CreateFileInfo(".aspect/cli/file1", false),
			CreateFileInfo(".aspect/cli/file2", false),
			CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
			CreateFileInfo("file", false),
			CreateFileInfo("folder/nestedfile", false),
			CreateFileInfo("folder/subfolder/deeplynestedfile", false),
			CreateFileInfo("folder/subfolder2/file", false),
			CreateFileInfo("folder/subfolder2/file2", false),
			CreateFileInfo("folder2/file", false),
		}, entries)
		entries, err = gitTree.Files(context.Background(), &gitTreeEntryConnectionArgs{Recursive: true})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/cli/file1", false),
			CreateFileInfo(".aspect/cli/file2", false),
			CreateFileInfo(".aspect/rules/external_repository_action_cache/file", false),
			CreateFileInfo("file", false),
			CreateFileInfo("folder/nestedfile", false),
			CreateFileInfo("folder/subfolder/deeplynestedfile", false),
			CreateFileInfo("folder/subfolder2/file", false),
			CreateFileInfo("folder/subfolder2/file2", false),
			CreateFileInfo("folder2/file", false),
		}, entries)
		entries, err = gitTree.Directories(context.Background(), &gitTreeEntryConnectionArgs{Recursive: true})
		require.NoError(t, err)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
			CreateFileInfo(".aspect/cli/", true),
			CreateFileInfo(".aspect/rules/", true),
			CreateFileInfo(".aspect/rules/external_repository_action_cache/", true),
			CreateFileInfo("folder/", true),
			CreateFileInfo("folder/subfolder/", true),
			CreateFileInfo("folder/subfolder2/", true),
			CreateFileInfo("folder2/", true),
		}, entries)
	})

	t.Run("Ancestors", func(t *testing.T) {
		opts := GitTreeEntryResolverOpts{
			Commit: &GitCommitResolver{
				repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
			},
			Stat: CreateFileInfo("folder/subfolder/", true),
		}
		gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

		entries, err := gitTree.Entries(context.Background(), &gitTreeEntryConnectionArgs{Ancestors: true})
		require.NoError(t, err)
		// TODO: This test is currently correct, but should we really return all
		// elements in the ancestors, or are we only interested in the parent
		// tree objects?
		// Also, the ordering here feels arbitrary(?)
		assertEntries(t, []fs.FileInfo{
			CreateFileInfo(".aspect/", true),
			CreateFileInfo("folder/", true),
			CreateFileInfo("folder2/", true),
			CreateFileInfo("file", false),
			CreateFileInfo("folder/subfolder/", true),
			CreateFileInfo("folder/subfolder2/", true),
			CreateFileInfo("folder/nestedfile", false),
			CreateFileInfo("folder/subfolder/deeplynestedfile", false),
		}, entries)
	})
}

func assertEntries(t *testing.T, expected []fs.FileInfo, entries []*GitTreeEntryResolver) {
	t.Helper()
	have := []fs.FileInfo{}
	for _, e := range entries {
		have = append(have, CreateFileInfo(e.Path(), e.IsDirectory()))
	}

	require.Equal(t, expected, have, "entries do not match expected")
}

func TestGitTree(t *testing.T) {
	db := dbmocks.NewMockDB()
	gsClient := setupGitserverClient(t)
	tests := []*Test{
		{
			Schema: mustParseGraphQLSchemaWithClient(t, db, gsClient),
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						commit(rev: "` + exampleCommitSHA1 + `") {
							tree(path: "foo bar") {
								directories {
									name
									path
									url
								}
								files {
									name
									path
									url
								}
							}
						}
					}
				}
			`,
			ExpectedResult: `
{
  "repository": {
    "commit": {
      "tree": {
        "directories": [
          {
            "name": "Geoffrey's random queries.32r242442bf",
            "path": "foo bar/Geoffrey's random queries.32r242442bf",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/tree/foo%20bar/Geoffrey%27s%20random%20queries.32r242442bf"
          },
          {
            "name": "testDirectory",
            "path": "foo bar/testDirectory",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/tree/foo%20bar/testDirectory"
          }
        ],
        "files": [
          {
            "name": "% token.4288249258.sql",
            "path": "foo bar/% token.4288249258.sql",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/blob/foo%20bar/%25%20token.4288249258.sql"
          },
          {
            "name": "testFile",
            "path": "foo bar/testFile",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/blob/foo%20bar/testFile"
          }
        ]
      }
    }
  }
}
			`,
		},
	}
	testGitTree(t, db, tests)
}

func setupGitserverClient(t *testing.T) gitserver.Client {
	t.Helper()
	gsClient := gitserver.NewMockClient()
	gsClient.ReadDirFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, commit api.CommitID, name string, recurse bool) (gitserver.ReadDirIterator, error) {
		assert.Equal(t, api.CommitID(exampleCommitSHA1), commit)
		assert.Equal(t, "foo bar", name)
		assert.False(t, recurse)
		return gitserver.NewReadDirIteratorFromSlice([]fs.FileInfo{
			&fileutil.FileInfo{Name_: name + "/testDirectory", Mode_: os.ModeDir},
			&fileutil.FileInfo{Name_: name + "/Geoffrey's random queries.32r242442bf", Mode_: os.ModeDir},
			&fileutil.FileInfo{Name_: name + "/testFile", Mode_: 0},
			&fileutil.FileInfo{Name_: name + "/% token.4288249258.sql", Mode_: 0},
		}), nil
	})
	gsClient.StatFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
		assert.Equal(t, api.CommitID(exampleCommitSHA1), commit)
		assert.Equal(t, "foo bar", path)
		return &fileutil.FileInfo{Name_: path, Mode_: os.ModeDir}, nil
	})
	return gsClient
}

func testGitTree(t *testing.T, db *dbmocks.MockDB, tests []*Test) {
	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(nil, nil)

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)
	repos.GetByNameFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)

	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo api.RepoName, rev string) (api.CommitID, error) {
		assert.Equal(t, api.RepoName("github.com/gorilla/mux"), repo)
		assert.Equal(t, exampleCommitSHA1, rev)
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &gitdomain.Commit{ID: exampleCommitSHA1})
	defer func() {
		backend.Mocks = backend.MockServices{}
	}()

	RunTests(t, tests)
}
