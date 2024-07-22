package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/object/mocks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServeSearchJobDownload(t *testing.T) {
	enabled := true
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{SearchJobs: &enabled}}})
	defer conf.Mock(nil)

	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger

	mockUploadStore := mocks.NewMockStorage()
	mockUploadStore.ListFunc.SetDefaultHook(
		func(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
			return iterator.From([]string{}), nil
		})

	db := database.NewDB(logger, dbtest.NewDB(t))
	bs := basestore.NewWithHandle(db.Handle())
	s := store.New(db, observation.TestContextTB(t))
	svc := service.New(observationCtx, s, mockUploadStore, service.NewSearcherFake())

	router := mux.NewRouter()
	router.HandleFunc("/{id}.json", ServeSearchJobDownload(logger, svc))

	// no job
	{
		req, err := http.NewRequest(http.MethodGet, "/99.json", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	}

	// no blobs
	{
		// create job
		userID, err := createUser(bs, "bob")
		require.NoError(t, err)
		userCtx := actor.WithActor(context.Background(), &actor.Actor{
			UID: userID,
		})
		_, err = svc.CreateSearchJob(userCtx, "1@rev1")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/1.json", nil)
		require.NoError(t, err)

		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: userID}))
		w := httptest.NewRecorder()
		w.Body = &bytes.Buffer{}
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "", w.Body.String())
	}

	// wrong user
	{
		userID, err := createUser(bs, "alice")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/1.json", nil)
		require.NoError(t, err)

		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: userID}))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	}
}

func createUser(store *basestore.Store, username string) (int32, error) {
	admin := username == "admin"
	q := sqlf.Sprintf(`INSERT INTO users(username, site_admin) VALUES(%s, %s) RETURNING id`, username, admin)
	return basestore.ScanAny[int32](store.QueryRow(context.Background(), q))
}
