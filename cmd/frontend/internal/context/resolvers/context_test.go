package resolvers

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codycontext "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/codycontext"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestContextResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	repo1 := types.Repo{Name: "repo1"}
	repo2 := types.Repo{Name: "repo2"}
	truePtr := true
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			CodyEnabled: &truePtr,
			LicenseKey:  "asdf",
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				CodyContextIgnore: pointers.Ptr(true),
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	oldMock := licensing.MockCheckFeature
	defer func() {
		licensing.MockCheckFeature = oldMock
	}()

	licensing.MockCheckFeature = func(feature licensing.Feature) error {
		if feature == licensing.FeatureCody {
			return nil
		}
		return errors.New("error")
	}

	// Create a normal user role with Cody access permission
	normalUserRole, err := db.Roles().Create(ctx, "normal user role", false)
	require.NoError(t, err)
	codyAccessPermission, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: rtypes.CodyNamespace,
		Action:    rtypes.CodyAccessAction,
	})
	require.NoError(t, err)
	err = db.RolePermissions().Assign(ctx, database.AssignRolePermissionOpts{
		PermissionID: codyAccessPermission.ID,
		RoleID:       normalUserRole.ID,
	})
	require.NoError(t, err)

	// Create an admin user, give them the normal user role, and authenticate our actor.
	newAdminUser, err := db.Users().Create(ctx, database.NewUser{
		Email:                 "test@example.com",
		Username:              "test",
		DisplayName:           "Test User",
		Password:              "hunter123",
		EmailIsVerified:       true,
		FailIfNotInitialUser:  true, // initial site admin account
		EnforcePasswordLength: false,
		TosAccepted:           true,
	})
	require.NoError(t, err)
	db.UserRoles().SetRolesForUser(ctx, database.SetRolesForUserOpts{
		UserID: newAdminUser.ID,
		Roles:  []int32{normalUserRole.ID},
	})
	require.NoError(t, err)
	ctx = actor.WithActor(ctx, actor.FromMockUser(newAdminUser.ID))

	// Create populates the IDs in the passed in types.Repo
	err = db.Repos().Create(ctx, &repo1, &repo2)
	require.NoError(t, err)

	files := map[api.RepoName]map[string][]byte{
		"repo1": {
			"testcode1.go": []byte("testcode1"),
			"ignore_me.go": []byte("secret"),
			"ignore_me.md": []byte("secret"),
			"testtext1.md": []byte("testtext1"),
			".cody/ignore": []byte("ignore_me.go\nignore_me.md"),
		},
		"repo2": {
			"testcode2.go": []byte("testcode2"),
			"ignore_me.go": []byte("secret"),
			"ignore_me.md": []byte("secret"),
			"testtext2.md": []byte("testtext2"),
			".cody/ignore": []byte("ignore_me.go\nignore_me.md"),
		},
	}

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.GetDefaultBranchFunc.SetDefaultReturn("main", api.CommitID("abc123"), nil)
	mockGitserver.StatFunc.SetDefaultHook(func(_ context.Context, repo api.RepoName, _ api.CommitID, fileName string) (fs.FileInfo, error) {
		return fakeFileInfo{path: fileName}, nil
	})
	mockGitserver.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, ci api.CommitID, fileName string) (io.ReadCloser, error) {
		if content, ok := files[repo][fileName]; ok {
			return io.NopCloser(bytes.NewReader(content)), nil
		}
		return nil, os.ErrNotExist
	})

	mockEmbeddingsClient := embeddings.NewMockClient()
	mockEmbeddingsClient.SearchFunc.SetDefaultReturn(nil, errors.New("embeddings should be disabled"))

	lineRange := func(start, end int) result.ChunkMatches {
		return result.ChunkMatches{{
			Ranges: result.Ranges{{
				Start: result.Location{Line: start},
				End:   result.Location{Line: end},
			}},
		}}
	}

	mockSearchClient := client.NewMockSearchClient()
	mockSearchClient.PlanFunc.SetDefaultHook(func(_ context.Context, _ string, _ *string, query string, _ search.Mode, _ search.Protocol, _ *int32) (*search.Inputs, error) {
		return &search.Inputs{OriginalQuery: query}, nil
	})
	mockSearchClient.ExecuteFunc.SetDefaultHook(func(_ context.Context, stream streaming.Sender, inputs *search.Inputs) (*search.Alert, error) {
		if strings.Contains(inputs.OriginalQuery, "-file") {
			stream.Send(streaming.SearchEvent{
				Results: result.Matches{&result.FileMatch{
					File: result.File{
						Path: "ignore_me.go",
						Repo: types.MinimalRepo{ID: repo1.ID, Name: repo1.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "ignore_me.go",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "testcode1.go",
						Repo: types.MinimalRepo{ID: repo1.ID, Name: repo1.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "testcode2.go",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}},
			})
		} else {
			stream.Send(streaming.SearchEvent{
				Results: result.Matches{&result.FileMatch{
					File: result.File{
						Path: "ignore_me.md",
						Repo: types.MinimalRepo{ID: repo1.ID, Name: repo1.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "ignore_me.md",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "testtext1.md",
						Repo: types.MinimalRepo{ID: repo1.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "testtext2.md",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}},
			})
		}
		return nil, nil
	})

	contextClient := codycontext.NewCodyContextClient(
		observation.NewContext(logger),
		db,
		mockEmbeddingsClient,
		mockSearchClient,
		nil,
		codycontext.NewCodyIgnoreFilter(mockGitserver),
	)

	resolver := NewResolver(
		db,
		mockGitserver,
		contextClient,
	)

	results, err := resolver.GetCodyContext(ctx, graphqlbackend.GetContextArgs{
		Repos:            graphqlbackend.MarshalRepositoryIDs([]api.RepoID{1, 2}),
		Query:            "my test query",
		TextResultsCount: 2,
		CodeResultsCount: 2,
	})
	require.NoError(t, err)

	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.(*graphqlbackend.FileChunkContextResolver).Blob().Path()
	}
	// One code result and text result from each repo
	expected := []string{"testcode1.go", "testcode2.go", "testtext1.md", "testtext2.md"}
	sort.Strings(expected)
	sort.Strings(paths)
	require.Equal(t, expected, paths)
}

type fakeFileInfo struct {
	path string
	fs.FileInfo
}

func (f fakeFileInfo) Name() string {
	return f.path
}
