package resolvers

import (
	"context"
	"reflect"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestSavedSearches(t *testing.T) {
	t.Run("same user", func(t *testing.T) {
		userID := int32(1)
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

		ss := dbmocks.NewMockSavedSearchStore()
		ss.ListFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs, paginationArgs *database.PaginationArgs) ([]*types.SavedSearch, error) {
			return []*types.SavedSearch{{ID: 1, Description: "test query", Query: "test type:diff patternType:regexp", Owner: *args.Owner}}, nil
		})
		ss.CountFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs) (int, error) {
			return 1, nil
		})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		ownerID := graphqlbackend.MarshalUserID(userID)
		args := graphqlbackend.SavedSearchesArgs{
			Owner:                  &ownerID,
			OrderBy:                graphqlbackend.SavedSearchesOrderByUpdatedAt,
			ConnectionResolverArgs: dummyConnectionResolverArgs,
		}

		resolver, err := newTestResolver(t, db).SavedSearches(actor.WithActor(context.Background(), actor.FromUser(userID)), args)
		if err != nil {
			t.Fatal(err)
		}

		nodes, err := resolver.Nodes(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		wantNodes := []graphqlbackend.SavedSearchResolver{
			&savedSearchResolver{db, types.SavedSearch{
				ID:          1,
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				Owner:       types.NamespaceUser(userID),
			}},
		}
		if !reflect.DeepEqual(nodes, wantNodes) {
			t.Errorf("got %v+, want %v+", nodes[0], wantNodes[0])
		}
	})

	t.Run("different user", func(t *testing.T) {
		userID := int32(2)
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

		ss := dbmocks.NewMockSavedSearchStore()
		ss.ListFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs, paginationArgs *database.PaginationArgs) ([]*types.SavedSearch, error) {
			panic("should fail auth check and never be called")
		})
		ss.CountFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs) (int, error) {
			panic("should fail auth check and never be called")
		})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		ownerID := graphqlbackend.MarshalUserID(3)
		args := graphqlbackend.SavedSearchesArgs{
			Owner:                  &ownerID,
			OrderBy:                graphqlbackend.SavedSearchesOrderByUpdatedAt,
			ConnectionResolverArgs: dummyConnectionResolverArgs,
		}

		_, err := newTestResolver(t, db).SavedSearches(actor.WithActor(context.Background(), actor.FromUser(userID)), args)
		if err == nil {
			t.Error("got nil, want error to be returned for accessing saved searches of different user by non-site admin")
		}
	})

	t.Run("different org", func(t *testing.T) {
		userID := int32(1)
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(&types.User{ID: userID}, nil)
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

		orgID := int32(2)
		om := dbmocks.NewMockOrgMemberStore()
		om.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(ctx context.Context, oid, uid int32) (*types.OrgMembership, error) {
			return nil, &database.ErrOrgMemberNotFound{}
		})

		ss := dbmocks.NewMockSavedSearchStore()
		ss.ListFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs, paginationArgs *database.PaginationArgs) ([]*types.SavedSearch, error) {
			return []*types.SavedSearch{{ID: 1, Description: "test query", Query: "test type:diff patternType:regexp", Owner: types.NamespaceOrg(orgID)}}, nil
		})
		ss.CountFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs) (int, error) {
			return 1, nil
		})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgMembersFunc.SetDefaultReturn(om)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		ownerID := graphqlbackend.MarshalOrgID(orgID)
		args := graphqlbackend.SavedSearchesArgs{
			Owner:                  &ownerID,
			OrderBy:                graphqlbackend.SavedSearchesOrderByUpdatedAt,
			ConnectionResolverArgs: dummyConnectionResolverArgs,
		}

		if _, err := newTestResolver(t, db).SavedSearches(actor.WithActor(context.Background(), actor.FromUser(userID)), args); err != auth.ErrNotAnOrgMember {
			t.Errorf("got %v+, want %v+", err, auth.ErrNotAnOrgMember)
		}
	})

	t.Run("anonymous visitor", func(t *testing.T) {
		userID := int32(1)
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(nil, nil)

		ss := dbmocks.NewMockSavedSearchStore()
		ss.ListFunc.SetDefaultReturn([]*types.SavedSearch{{ID: 1, Description: "d", Owner: types.NamespaceUser(userID), VisibilitySecret: false}}, nil)
		ss.CountFunc.SetDefaultHook(func(_ context.Context, args database.SavedSearchListArgs) (int, error) {
			return 1, nil
		})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		args := graphqlbackend.SavedSearchesArgs{
			ViewerIsAffiliated:     pointers.Ptr(true),
			OrderBy:                graphqlbackend.SavedSearchesOrderByUpdatedAt,
			ConnectionResolverArgs: dummyConnectionResolverArgs,
		}
		ctx := actor.WithActor(context.Background(), actor.FromAnonymousUser(""))
		resolver, err := newTestResolver(t, db).SavedSearches(ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := resolver.Nodes(context.Background()); err != nil {
			t.Fatal(err)
		}
		mockrequire.CalledOnceWith(t, ss.ListFunc, mockrequire.Values(mockrequire.Skip,
			database.SavedSearchListArgs{
				PublicOnly: true,
				HideDrafts: true,
			},
		))
	})

	t.Run("forbid enumerating non-public results by non-site admin", func(t *testing.T) {
		userID := int32(1)
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

		ss := dbmocks.NewMockSavedSearchStore()

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		args := graphqlbackend.SavedSearchesArgs{
			OrderBy:                graphqlbackend.SavedSearchesOrderByUpdatedAt,
			ConnectionResolverArgs: dummyConnectionResolverArgs,
		}
		ctx := actor.WithActor(context.Background(), actor.FromUser(userID))
		resolver, err := newTestResolver(t, db).SavedSearches(ctx, args)
		if want := auth.ErrMustBeSiteAdmin; !errors.Is(err, want) {
			t.Fatalf("got %v+, want %v+", err, want)
		}
		if resolver != nil {
			t.Fatal("want nil resolver")
		}

		mockrequire.NotCalled(t, ss.ListFunc)
	})
}

func TestSavedSearchByID(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		ctx := context.Background()

		userID := int32(1)
		fixtureID := int32(1)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

		ss := dbmocks.NewMockSavedSearchStore()
		ss.GetByIDFunc.SetDefaultReturn(
			&types.SavedSearch{
				ID:          fixtureID,
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				Owner:       types.NamespaceUser(userID),
			},
			nil,
		)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		ctx = actor.WithActor(ctx, &actor.Actor{UID: userID})

		result, err := newTestResolver(t, db).SavedSearchByID(ctx, marshalSavedSearchID(fixtureID))
		if err != nil {
			t.Fatal(err)
		}
		want := &savedSearchResolver{
			db: db,
			s: types.SavedSearch{
				ID:          fixtureID,
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				Owner:       types.NamespaceUser(userID),
			},
		}

		if !reflect.DeepEqual(result, want) {
			t.Errorf("got %v+, want %v+", result, want)
		}
	})

	t.Run("non-owner", func(t *testing.T) {
		// Non-owners cannot view a user's saved searches.
		userID := int32(1)
		otherUserID := int32(2)
		fixtureID := marshalSavedSearchID(1)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: otherUserID}, nil)

		ss := dbmocks.NewMockSavedSearchStore()
		ss.GetByIDFunc.SetDefaultReturn(
			&types.SavedSearch{
				Description:      "test query",
				Query:            "test type:diff patternType:regexp",
				Owner:            types.NamespaceUser(userID),
				VisibilitySecret: true,
			},
			nil,
		)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SavedSearchesFunc.SetDefaultReturn(ss)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: otherUserID})

		_, err := newTestResolver(t, db).SavedSearchByID(ctx, fixtureID)
		if err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestCreateSavedSearch(t *testing.T) {
	userID := int32(1)
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})

	ss := dbmocks.NewMockSavedSearchStore()
	ss.CreateFunc.SetDefaultHook(func(_ context.Context, newSavedSearch *types.SavedSearch) (*types.SavedSearch, error) {
		return &types.SavedSearch{
			ID:               1,
			Description:      newSavedSearch.Description,
			Query:            newSavedSearch.Query,
			Draft:            newSavedSearch.Draft,
			Owner:            newSavedSearch.Owner,
			VisibilitySecret: newSavedSearch.VisibilitySecret,
			CreatedByUser:    &userID,
			UpdatedByUser:    &userID,
		}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	t.Run("visibility secret", func(t *testing.T) {
		result, err := newTestResolver(t, db).CreateSavedSearch(ctx, &graphqlbackend.CreateSavedSearchArgs{
			Input: graphqlbackend.SavedSearchInput{
				Owner:       graphqlbackend.MarshalUserID(userID),
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				Draft:       true,
				Visibility:  graphqlbackend.SavedSearchVisibilitySecret,
			}})
		if err != nil {
			t.Fatal(err)
		}
		want := &savedSearchResolver{db, types.SavedSearch{
			ID:               1,
			Description:      "test query",
			Query:            "test type:diff patternType:regexp",
			Draft:            true,
			Owner:            types.NamespaceUser(userID),
			VisibilitySecret: true,
			CreatedByUser:    &userID,
			UpdatedByUser:    &userID,
		}}

		mockrequire.Called(t, ss.CreateFunc)

		if !reflect.DeepEqual(result, want) {
			t.Errorf("got %v+, want %v+", result, want)
		}

		// Ensure create saved search errors when patternType is not provided in the query.
		_, err = newTestResolver(t, db).CreateSavedSearch(ctx, &graphqlbackend.CreateSavedSearchArgs{
			Input: graphqlbackend.SavedSearchInput{
				Description: "test query",
				Query:       "test type:diff",
				Owner:       graphqlbackend.MarshalUserID(userID),
			}})
		if err == nil {
			t.Error("Expected error for createSavedSearch when query does not provide a patternType: field.")
		}
	})

	t.Run("visibility public", func(t *testing.T) {
		_, err := newTestResolver(t, db).CreateSavedSearch(ctx, &graphqlbackend.CreateSavedSearchArgs{
			Input: graphqlbackend.SavedSearchInput{
				Owner:       graphqlbackend.MarshalUserID(userID),
				Description: "test query",
				Query:       "test type:diff patternType:regexp",
				Visibility:  graphqlbackend.SavedSearchVisibilityPublic,
			}})
		if want := auth.ErrMustBeSiteAdmin; err != want {
			t.Fatalf("got %v, want %v", err, want)
		}
	})
}

func TestUpdateSavedSearch(t *testing.T) {
	fixtureID := int32(1)
	userID := int32(1)
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})

	ss := dbmocks.NewMockSavedSearchStore()
	ss.UpdateFunc.SetDefaultHook(func(ctx context.Context, savedSearch *types.SavedSearch) (*types.SavedSearch, error) {
		return &types.SavedSearch{
			ID:            fixtureID,
			Description:   savedSearch.Description,
			Query:         savedSearch.Query,
			Owner:         savedSearch.Owner,
			Draft:         savedSearch.Draft,
			CreatedByUser: &userID,
			UpdatedByUser: &userID,
		}, nil
	})
	ss.GetByIDFunc.SetDefaultReturn(&types.SavedSearch{ID: fixtureID, Owner: types.NamespaceUser(userID)}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	result, err := newTestResolver(t, db).UpdateSavedSearch(ctx, &graphqlbackend.UpdateSavedSearchArgs{
		ID: marshalSavedSearchID(fixtureID),
		Input: graphqlbackend.SavedSearchUpdateInput{
			Description: "updated query description",
			Query:       "test type:diff patternType:regexp",
			Draft:       true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &savedSearchResolver{db, types.SavedSearch{
		ID:            fixtureID,
		Description:   "updated query description",
		Query:         "test type:diff patternType:regexp",
		Draft:         true,
		Owner:         types.NamespaceUser(userID),
		CreatedByUser: &userID,
		UpdatedByUser: &userID,
	}}

	mockrequire.Called(t, ss.UpdateFunc)

	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v+, want %v+", result, want)
	}

	// Ensure update saved search errors when patternType is not provided in the query.
	_, err = newTestResolver(t, db).UpdateSavedSearch(ctx, &graphqlbackend.UpdateSavedSearchArgs{
		ID:    marshalSavedSearchID(fixtureID),
		Input: graphqlbackend.SavedSearchUpdateInput{Description: "updated query description", Query: "test type:diff"}},
	)
	if err == nil {
		t.Error("Expected error for updateSavedSearch when query does not provide a patternType: field.")
	}
}

func TestTransferSavedSearchOwnership(t *testing.T) {
	fixtureID := int32(1)
	userID := int32(1)
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})

	orgID := int32(2)
	om := dbmocks.NewMockOrgMemberStore()
	om.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(ctx context.Context, oid, uid int32) (*types.OrgMembership, error) {
		return &types.OrgMembership{OrgID: oid, UserID: uid}, nil
	})

	mockStore := dbmocks.NewMockSavedSearchStore()
	mockStore.UpdateOwnerFunc.SetDefaultHook(func(ctx context.Context, id int32, newOwner types.Namespace) (*types.SavedSearch, error) {
		return &types.SavedSearch{
			ID:    id,
			Owner: newOwner,
		}, nil
	})
	mockStore.GetByIDFunc.SetDefaultReturn(&types.SavedSearch{ID: fixtureID, Owner: types.NamespaceUser(userID)}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(om)
	db.SavedSearchesFunc.SetDefaultReturn(mockStore)

	result, err := newTestResolver(t, db).TransferSavedSearchOwnership(ctx, &graphqlbackend.TransferSavedSearchOwnershipArgs{
		ID:       marshalSavedSearchID(fixtureID),
		NewOwner: graphqlbackend.MarshalOrgID(orgID),
	})
	if err != nil {
		t.Fatal(err)
	}

	mockrequire.Called(t, mockStore.UpdateOwnerFunc)
	mockrequire.Called(t, om.GetByOrgIDAndUserIDFunc)
	want := &savedSearchResolver{db, types.SavedSearch{
		ID:    fixtureID,
		Owner: types.NamespaceOrg(orgID),
	}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v+, want %v+", result, want)
	}
}

func TestChangeSavedSearchVisibility(t *testing.T) {
	fixtureID := int32(1)
	userID := int32(1)
	users := dbmocks.NewMockUserStore()

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})

	ss := dbmocks.NewMockSavedSearchStore()
	ss.UpdateVisibilityFunc.SetDefaultHook(func(ctx context.Context, id int32, secret bool) (*types.SavedSearch, error) {
		return &types.SavedSearch{
			ID:               id,
			Owner:            types.NamespaceUser(userID),
			VisibilitySecret: secret,
		}, nil
	})
	ss.GetByIDFunc.SetDefaultReturn(&types.SavedSearch{ID: fixtureID, Owner: types.NamespaceUser(userID)}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	t.Run("non-site admin", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

		_, err := newTestResolver(t, db).ChangeSavedSearchVisibility(ctx, &graphqlbackend.ChangeSavedSearchVisibilityArgs{
			ID:            marshalSavedSearchID(fixtureID),
			NewVisibility: graphqlbackend.SavedSearchVisibilitySecret,
		})
		if want := auth.ErrMustBeSiteAdmin; err != want {
			t.Fatalf("got err %v, want %v", err, want)
		}
	})

	t.Run("site admin", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID, SiteAdmin: true}, nil)

		result, err := newTestResolver(t, db).ChangeSavedSearchVisibility(ctx, &graphqlbackend.ChangeSavedSearchVisibilityArgs{
			ID:            marshalSavedSearchID(fixtureID),
			NewVisibility: graphqlbackend.SavedSearchVisibilitySecret,
		})
		if err != nil {
			t.Fatal(err)
		}

		mockrequire.Called(t, ss.UpdateVisibilityFunc)
		want := &savedSearchResolver{db, types.SavedSearch{
			ID:               fixtureID,
			Owner:            types.NamespaceUser(userID),
			VisibilitySecret: true,
		}}
		if !reflect.DeepEqual(result, want) {
			t.Errorf("got %v+, want %v+", result, want)
		}
	})
}

func TestSavedSearchPermissions(t *testing.T) {
	user1 := &types.User{ID: 42}
	user2 := &types.User{ID: 43}
	admin := &types.User{ID: 44, SiteAdmin: true}
	org1 := &types.Org{ID: 42}
	org2 := &types.Org{ID: 43}

	cases := []struct {
		execUser            *types.User
		ownerUserID         *int32
		ownerOrgID          *int32
		visibilitySecret    bool
		viewerCanView       bool
		viewerCanAdminister bool
		opErrIs             error
	}{
		{
			execUser:            user1,
			ownerUserID:         &user1.ID,
			visibilitySecret:    true,
			viewerCanView:       true,
			viewerCanAdminister: true,
			opErrIs:             nil,
		},
		{
			execUser:            user1,
			ownerUserID:         &user2.ID,
			visibilitySecret:    true,
			viewerCanView:       false,
			viewerCanAdminister: false,
			opErrIs:             &auth.InsufficientAuthorizationError{},
		},
		{
			execUser:            user1,
			ownerUserID:         &user2.ID,
			visibilitySecret:    false,
			viewerCanView:       true,
			viewerCanAdminister: false,
			opErrIs:             &auth.InsufficientAuthorizationError{},
		},
		{
			execUser:            user1,
			ownerOrgID:          &org1.ID,
			visibilitySecret:    true,
			viewerCanView:       true,
			viewerCanAdminister: true,
			opErrIs:             nil,
		},
		{
			execUser:            user1,
			ownerOrgID:          &org2.ID,
			visibilitySecret:    true,
			viewerCanView:       false,
			viewerCanAdminister: false,
			opErrIs:             auth.ErrNotAnOrgMember,
		},
		{
			execUser:            user1,
			ownerOrgID:          &org2.ID,
			visibilitySecret:    false,
			viewerCanView:       true,
			viewerCanAdminister: false,
			opErrIs:             auth.ErrNotAnOrgMember,
		},
		{
			execUser:            admin,
			ownerOrgID:          &user1.ID,
			viewerCanView:       true,
			viewerCanAdminister: true,
			opErrIs:             nil,
		},
		{
			execUser:            admin,
			ownerOrgID:          &org1.ID,
			viewerCanView:       true,
			viewerCanAdminister: true,
			opErrIs:             nil,
		}}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), actor.FromUser(tt.execUser.ID))
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
				switch actor.FromContext(ctx).UID {
				case user1.ID:
					return user1, nil
				case user2.ID:
					return user2, nil
				case admin.ID:
					return admin, nil
				default:
					panic("bad actor")
				}
			})

			owner := types.Namespace{
				User: tt.ownerUserID,
				Org:  tt.ownerOrgID,
			}

			savedSearches := dbmocks.NewMockSavedSearchStore()
			savedSearches.CreateFunc.SetDefaultHook(func(_ context.Context, ss *types.SavedSearch) (*types.SavedSearch, error) {
				ss.Owner = owner
				ss.VisibilitySecret = tt.visibilitySecret
				return ss, nil
			})
			savedSearches.UpdateFunc.SetDefaultHook(func(_ context.Context, ss *types.SavedSearch) (*types.SavedSearch, error) {
				ss.Owner = owner
				ss.VisibilitySecret = tt.visibilitySecret
				return ss, nil
			})
			savedSearches.GetByIDFunc.SetDefaultReturn(&types.SavedSearch{
				Owner:            owner,
				VisibilitySecret: tt.visibilitySecret,
			}, nil)

			orgMembers := dbmocks.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(_ context.Context, orgID int32, userID int32) (*types.OrgMembership, error) {
				if orgID == userID {
					return &types.OrgMembership{}, nil
				}
				return nil, &database.ErrOrgMemberNotFound{}
			})

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.SavedSearchesFunc.SetDefaultReturn(savedSearches)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)

			{
				// Get
				result, err := newTestResolver(t, db).SavedSearchByID(ctx, marshalSavedSearchID(1))
				if couldView := err == nil; couldView != tt.viewerCanView {
					t.Fatalf("got couldView %v (error %v), want %v", couldView, err, tt.viewerCanView)
				}
				if result != nil {
					gotCanAdminister := result.ViewerCanAdminister(ctx)
					if gotCanAdminister != tt.viewerCanAdminister {
						t.Errorf("got %v, want %v", gotCanAdminister, tt.viewerCanAdminister)
					}
				}
			}

			{
				// Create
				var ownerID graphql.ID
				if owner.User != nil {
					ownerID = graphqlbackend.MarshalUserID(*owner.User)
				} else if owner.Org != nil {
					ownerID = graphqlbackend.MarshalOrgID(*owner.Org)
				} else {
					panic("bad owner")
				}

				var visibility graphqlbackend.SavedSearchVisibility
				if tt.visibilitySecret {
					visibility = graphqlbackend.SavedSearchVisibilitySecret
				} else {
					visibility = graphqlbackend.SavedSearchVisibilityPublic
				}

				_, err := newTestResolver(t, db).CreateSavedSearch(ctx, &graphqlbackend.CreateSavedSearchArgs{
					Input: graphqlbackend.SavedSearchInput{
						Owner:       ownerID,
						Description: "d",
						Query:       "patterntype:literal",
						Visibility:  visibility,
					},
				})
				if tt.opErrIs == nil {
					require.NoError(t, err)
				} else {
					require.ErrorAs(t, err, &tt.opErrIs)
				}
			}

			{
				// Update
				_, err := newTestResolver(t, db).UpdateSavedSearch(ctx, &graphqlbackend.UpdateSavedSearchArgs{
					ID: marshalSavedSearchID(1),
					Input: graphqlbackend.SavedSearchUpdateInput{
						Description: "d",
						Query:       "patterntype:literal",
					},
				})
				if tt.opErrIs == nil {
					require.NoError(t, err)
				} else {
					require.ErrorAs(t, err, &tt.opErrIs)
				}
			}

			{
				// Delete
				_, err := newTestResolver(t, db).DeleteSavedSearch(ctx, &graphqlbackend.DeleteSavedSearchArgs{
					ID: marshalSavedSearchID(1),
				})
				if tt.opErrIs == nil {
					require.NoError(t, err)
				} else {
					require.ErrorAs(t, err, &tt.opErrIs)
				}
			}
		})
	}
}

func TestDeleteSavedSearch(t *testing.T) {
	userID := int32(1)
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultReturn(&types.User{ID: userID}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})

	ss := dbmocks.NewMockSavedSearchStore()
	ss.GetByIDFunc.SetDefaultReturn(&types.SavedSearch{
		ID:          1,
		Description: "test query",
		Query:       "test type:diff",
		Owner:       types.NamespaceUser(userID),
	}, nil)

	ss.DeleteFunc.SetDefaultReturn(nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SavedSearchesFunc.SetDefaultReturn(ss)

	firstSavedSearchGraphqlID := graphql.ID("U2F2ZWRTZWFyY2g6NTI=")
	_, err := newTestResolver(t, db).DeleteSavedSearch(ctx, &graphqlbackend.DeleteSavedSearchArgs{ID: firstSavedSearchGraphqlID})
	if err != nil {
		t.Fatal(err)
	}

	mockrequire.Called(t, ss.DeleteFunc)
}

func newTestResolver(t *testing.T, db database.DB) *Resolver {
	t.Helper()
	return &Resolver{db: db, logger: logtest.Scoped(t)}
}
