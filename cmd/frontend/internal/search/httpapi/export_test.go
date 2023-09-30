package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func TestServeSearchJobDownload(t *testing.T) {
	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger

	mockUploadStore := mocks.NewMockStore()
	mockUploadStore.ListFunc.SetDefaultHook(
		func(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
			return iterator.From([]string{}), nil
		})

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(db, observation.TestContextTB(t))
	svc := service.New(observationCtx, s, mockUploadStore)

	router := mux.NewRouter()
	router.HandleFunc("/{id}.csv", ServeSearchJobDownload(svc))

	// no job
	{
		req, err := http.NewRequest(http.MethodGet, "/99.csv", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	}

	// no blobs
	{
		// create job
		userID, err := createUser(db, "bob")
		require.NoError(t, err)
		userCtx := actor.WithActor(context.Background(), &actor.Actor{
			UID: userID,
		})
		_, err = svc.CreateSearchJob(userCtx, "foo")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/1.csv", nil)
		require.NoError(t, err)

		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: userID}))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
	}

	// wrong user
	{
		userID, err := createUser(db, "alice")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/1.csv", nil)
		require.NoError(t, err)

		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: userID}))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	}
}

func createUser(db database.DB, username string) (userID int32, err error) {
	admin := username == "admin"
	ctx := context.Background()

	q := sqlf.Sprintf("INSERT INTO users (username) VALUES (%s) RETURNING id", username)
	userID, err = basestore.ScanAny[int32](db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	if err != nil {
		return
	}

	roles := []types.SystemRole{types.UserSystemRole}
	if admin {
		roles = append(roles, types.SiteAdministratorSystemRole)
	}

	err = db.UserRoles().BulkAssignSystemRolesToUser(ctx, database.BulkAssignSystemRolesToUserOpts{
		UserID: userID,
		Roles:  roles,
	})
	return
}
