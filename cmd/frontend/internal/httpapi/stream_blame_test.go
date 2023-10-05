package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockGSClient(t *testing.T, wantRev api.CommitID, returnErr error, hunks []*gitserver.Hunk) gitserver.Client {
	hunkReader := gitserver.NewMockHunkReader(hunks, returnErr)
	gsClient := gitserver.NewMockClient()
	gsClient.GetCommitFunc.SetDefaultHook(
		func(_ context.Context,
			repoName api.RepoName,
			commit api.CommitID,
			opts gitserver.ResolveRevisionOptions,
		) (*gitdomain.Commit, error) {
			return &gitdomain.Commit{
				Parents: []api.CommitID{"xxx", "yyy"},
			}, nil
		})
	gsClient.StreamBlameFileFunc.SetDefaultHook(
		func(
			ctx context.Context,
			repo api.RepoName,
			path string,
			opts *gitserver.BlameOptions,
		) (gitserver.HunkReader, error) {
			if want, got := wantRev, opts.NewestCommit; want != got {
				t.Logf("want %s, got %s", want, got)
				t.Fail()
			}
			return hunkReader, nil
		})
	return gsClient
}

func TestStreamBlame(t *testing.T) {
	logger, _ := logtest.Captured(t)

	hunks := []*gitserver.Hunk{
		{
			StartLine: 1,
			EndLine:   2,
			CommitID:  api.CommitID("abcd"),
			Author: gitdomain.Signature{
				Name:  "Bob",
				Email: "bob@internet.com",
				Date:  time.Now(),
			},
			Message:  "one",
			Filename: "foo.c",
		},
		{
			StartLine: 2,
			EndLine:   3,
			CommitID:  api.CommitID("ijkl"),
			Author: gitdomain.Signature{
				Name:  "Bob",
				Email: "bob@internet.com",
				Date:  time.Now(),
			},
			Message:  "two",
			Filename: "foo.c",
		},
	}

	db := dbmocks.NewMockDB()
	backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		if name == "github.com/bob/foo" {
			return &types.Repo{Name: name}, nil
		}

		// A repo synced from src serve-git.
		if name == "foo" {
			return &types.Repo{
				Name: name,
				URI:  "repos/foo",
			}, nil
		}

		return nil, &database.RepoNotFoundErr{Name: name}
	}
	backend.Mocks.Repos.Get = func(ctx context.Context, repo api.RepoID) (*types.Repo, error) {
		return &types.Repo{Name: "github.com/bob/foo"}, nil
	}
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		switch rev {
		case "1234":
			return "efgh", nil
		case "":
			return "abcd", nil
		default:
			return "", &gitdomain.RevisionNotFoundError{Repo: repo.Name}
		}
	}
	usersStore := dbmocks.NewMockUserStore()
	errNotFound := &errcode.Mock{
		IsNotFound: true,
	}
	usersStore.GetByVerifiedEmailFunc.SetDefaultReturn(nil, errNotFound)
	db.UsersFunc.SetDefaultReturn(usersStore)

	t.Cleanup(func() {
		backend.Mocks.Repos = backend.MockRepos{}
	})

	ffs := featureflag.NewMemoryStore(nil, nil, map[string]bool{"enable-streaming-git-blame": true})
	ctx := featureflag.WithFlags(context.Background(), ffs)

	t.Run("NOK feature flag disabled", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/no-vars", nil)
		require.NoError(t, err)
		req = req.WithContext(context.Background()) // no feature flag there

		gsClient := setupMockGSClient(t, "abcd", nil, hunks)
		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("NOK no mux vars", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/no-vars", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		gsClient := setupMockGSClient(t, "abcd", nil, hunks)
		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("NOK repo not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVars(req, map[string]string{
			"Repo": "github.com/bob/bar",
			"path": "foo.c",
		})
		gsClient := setupMockGSClient(t, "abcd", nil, hunks)
		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("NOK rev not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVars(req, map[string]string{
			"Repo": "github.com/bob/foo",
			"path": "foo.c",
			"Rev":  "@void",
		})
		gsClient := setupMockGSClient(t, "abcd", nil, hunks)
		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("OK no rev", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVars(req, map[string]string{
			"Repo": "github.com/bob/foo",
			"path": "foo.c",
		})
		gsClient := setupMockGSClient(t, "abcd", nil, hunks)
		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		data := rec.Body.String()
		assert.Contains(t, data, `"commitID":"abcd"`)
		assert.Contains(t, data, `"commitID":"ijkl"`)
		assert.Contains(t, data, `done`)
	})

	t.Run("OK rev", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVars(req, map[string]string{
			"Rev":  "@1234",
			"Repo": "github.com/bob/foo",
			"path": "foo.c",
		})
		gsClient := setupMockGSClient(t, "efgh", nil, []*gitserver.Hunk{
			{
				StartLine: 1,
				EndLine:   2,
				CommitID:  api.CommitID("efgh"),
				Author: gitdomain.Signature{
					Name:  "Bob",
					Email: "bob@internet.com",
					Date:  time.Now(),
				},
				Message:  "one",
				Filename: "foo.c",
			},
		})

		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		data := rec.Body.String()
		assert.Contains(t, data, `"commitID":"efgh"`)
		assert.Contains(t, data, `done`)
		assert.Contains(t, data, `"url":"github.com/bob/foo/-/commit/efgh"`)
	})

	t.Run("NOK err reading hunks", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVars(req, map[string]string{
			"Repo": "github.com/bob/foo",
			"path": "foo.c",
		})
		gsClient := setupMockGSClient(t, "abcd", errors.New("foo"), hunks)
		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("src-serve OK rev", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVars(req, map[string]string{
			"Rev":  "@1234",
			"Repo": "foo",
			"path": "foo.c",
		})
		gsClient := setupMockGSClient(t, "efgh", nil, []*gitserver.Hunk{
			{
				StartLine: 1,
				EndLine:   2,
				CommitID:  api.CommitID("efgh"),
				Author: gitdomain.Signature{
					Name:  "Bob",
					Email: "bob@internet.com",
					Date:  time.Now(),
				},
				Message:  "one",
				Filename: "foo.c",
			},
		})

		handleStreamBlame(logger, db, gsClient).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		data := rec.Body.String()
		assert.Contains(t, data, `"commitID":"efgh"`)
		assert.Contains(t, data, `done`)
		assert.Contains(t, data, `"url":"foo/-/commit/efgh"`)
	})
}
