package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	autoIndexingEnabled = func() bool { return true }
}

func TestDeleteLSIFIndex(t *testing.T) {
	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := database.NewStrictMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	repos := database.NewStrictMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*sgtypes.Repo, error) {
		return &sgtypes.Repo{ID: id, Name: api.RepoName(fmt.Sprintf("repo%d", id))}, nil
	})

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFIndex:42")))
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockAutoIndexingService := NewMockAutoIndexingService()

	siteAdminChecker := sharedresolvers.NewSiteAdminChecker(db)
	locationResolverFactory := sharedresolvers.NewCachedLocationResolverFactory(nil, repos, gitserver.NewClient())
	rootResolver := NewRootResolver(&observation.TestContext, mockAutoIndexingService, mockUploadsService, mockPolicyService, siteAdminChecker, nil, nil, locationResolverFactory)

	if _, err := rootResolver.DeleteLSIFIndex(context.Background(), &struct{ ID graphql.ID }{id}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockAutoIndexingService.DeleteIndexByIDFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockAutoIndexingService.DeleteIndexByIDFunc.History()))
	}
	if val := mockAutoIndexingService.DeleteIndexByIDFunc.History()[0].Arg1; val != 42 {
		t.Fatalf("unexpected upload id. want=%d have=%d", 42, val)
	}
}

func TestDeleteLSIFIndexUnauthenticated(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, nil)
	repos := database.NewStrictMockRepoStore()

	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte("LSIFIndex:42")))
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockAutoIndexingService := NewMockAutoIndexingService()

	siteAdminChecker := sharedresolvers.NewSiteAdminChecker(db)
	locationResolverFactory := sharedresolvers.NewCachedLocationResolverFactory(nil, repos, gitserver.NewClient())
	rootResolver := NewRootResolver(&observation.TestContext, mockAutoIndexingService, mockUploadsService, mockPolicyService, siteAdminChecker, nil, nil, locationResolverFactory)

	if _, err := rootResolver.DeleteLSIFIndex(context.Background(), &struct{ ID graphql.ID }{id}); err != auth.ErrNotAuthenticated {
		t.Errorf("unexpected error. want=%q have=%q", auth.ErrNotAuthenticated, err)
	}
}
