package graphql

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestDeleteLSIFUpload(t *testing.T) {
	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFUpload:42")))

	mockUploadService := NewMockUploadService()
	mockPolicyService := NewMockPolicyService()
	mockAutoIndexingService := NewMockAutoIndexingService()
	mockAutoIndexingService.GetUnsafeDBFunc.SetDefaultReturn(db)

	rootResolver := NewRootResolver(&observation.TestContext, mockUploadService, mockAutoIndexingService, mockPolicyService)

	if _, err := rootResolver.DeleteLSIFUpload(context.Background(), &struct{ ID graphql.ID }{id}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockUploadService.DeleteUploadByIDFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockUploadService.DeleteUploadByIDFunc.History()))
	}

	if val := mockUploadService.DeleteUploadByIDFunc.History()[0].Arg1; val != 42 {
		t.Fatalf("unexpected upload id. want=%d have=%d", 42, val)
	}
}

func TestDeleteLSIFUploadUnauthenticated(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, nil)

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFUpload:42")))
	mockUploadService := NewMockUploadService()
	mockPolicyService := NewMockPolicyService()
	mockAutoIndexingService := NewMockAutoIndexingService()
	mockAutoIndexingService.GetUnsafeDBFunc.SetDefaultReturn(db)

	rootResolver := NewRootResolver(&observation.TestContext, mockUploadService, mockAutoIndexingService, mockPolicyService)

	if _, err := rootResolver.DeleteLSIFUpload(context.Background(), &struct{ ID graphql.ID }{id}); err != auth.ErrNotAuthenticated {
		t.Errorf("unexpected error. want=%q have=%q", auth.ErrNotAuthenticated, err)
	}
}
