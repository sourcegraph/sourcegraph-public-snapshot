package graphql

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	resolvermocks "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/mocks"
	transportmocks "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/mocks/transport"
	uploadmocks "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/mocks/transport/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	autoIndexingEnabled = func() bool { return true }
}

func TestDeleteLSIFUpload(t *testing.T) {
	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFUpload:42")))
	mockUploadResolver := uploadmocks.NewMockResolver()
	mockResolver := resolvermocks.NewMockResolver()
	mockResolver.UploadsResolverFunc.SetDefaultReturn(mockUploadResolver)

	if _, err := NewResolver(db, nil, mockResolver, &observation.TestContext).DeleteLSIFUpload(context.Background(), &struct{ ID graphql.ID }{id}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockUploadResolver.DeleteUploadByIDFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockUploadResolver.DeleteUploadByIDFunc.History()))
	}
	if val := mockUploadResolver.DeleteUploadByIDFunc.History()[0].Arg1; val != 42 {
		t.Fatalf("unexpected upload id. want=%d have=%d", 42, val)
	}
}

func TestDeleteLSIFUploadUnauthenticated(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, nil)

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFUpload:42")))
	mockResolver := resolvermocks.NewMockResolver()

	if _, err := NewResolver(db, nil, mockResolver, &observation.TestContext).DeleteLSIFUpload(context.Background(), &struct{ ID graphql.ID }{id}); err != backend.ErrNotAuthenticated {
		t.Errorf("unexpected error. want=%q have=%q", backend.ErrNotAuthenticated, err)
	}
}

func TestDeleteLSIFIndex(t *testing.T) {
	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := database.NewStrictMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFIndex:42")))
	mockResolver := resolvermocks.NewMockResolver()
	mockAutoIndexingResolver := transportmocks.NewMockResolver()
	mockResolver.AutoIndexingResolverFunc.PushReturn(mockAutoIndexingResolver)

	if _, err := NewResolver(db, nil, mockResolver, &observation.TestContext).DeleteLSIFIndex(context.Background(), &struct{ ID graphql.ID }{id}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockAutoIndexingResolver.DeleteIndexByIDFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockAutoIndexingResolver.DeleteIndexByIDFunc.History()))
	}
	if val := mockAutoIndexingResolver.DeleteIndexByIDFunc.History()[0].Arg1; val != 42 {
		t.Fatalf("unexpected index id. want=%d have=%d", 42, val)
	}
}

func TestDeleteLSIFIndexUnauthenticated(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, nil)

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFIndex:42")))
	mockResolver := resolvermocks.NewMockResolver()

	if _, err := NewResolver(db, nil, mockResolver, &observation.TestContext).DeleteLSIFIndex(context.Background(), &struct{ ID graphql.ID }{id}); err != backend.ErrNotAuthenticated {
		t.Errorf("unexpected error. want=%q have=%q", backend.ErrNotAuthenticated, err)
	}
}
