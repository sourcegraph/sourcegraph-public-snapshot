package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func (r *settingsSubjectResolver) mockCheckedAccessForTest() {
	// 🚨 SECURITY: For use in test mock values only.
	r.checkedAccess_DO_NOT_SET_THIS_MANUALLY_OR_YOU_WILL_LEAK_SECRETS = true
}

func TestSettingsSubjectForNodeAndCheckAccess(t *testing.T) {
	userID := int32(1)
	otherUserID := int32(2)
	orgID := int32(2)

	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	orgMembers := dbmocks.NewMockOrgMemberStore()

	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)

	orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(ctx context.Context, orgID, userID int32) (*types.OrgMembership, error) {
		if orgID == 2 && userID == 1 {
			return &types.OrgMembership{}, nil
		}
		return nil, &database.ErrOrgMemberNotFound{}
	})

	cases := []struct {
		name        string
		node        Node
		actor       *actor.Actor
		isDotcom    bool
		wantError   error
		wantSubject *settingsSubjectResolver
	}{
		{
			name:        "site settings",
			node:        &siteResolver{db: db},
			actor:       actor.FromActualUser(&types.User{ID: userID}),
			wantSubject: &settingsSubjectResolver{site: &siteResolver{db: db}},
		},
		{
			name:        "user settings - same user",
			node:        &UserResolver{user: &types.User{ID: userID}, db: db},
			actor:       actor.FromActualUser(&types.User{ID: userID}),
			wantSubject: &settingsSubjectResolver{user: &UserResolver{user: &types.User{ID: userID}, db: db}},
		},
		{
			name:        "user settings - site admin",
			node:        &UserResolver{user: &types.User{ID: userID}, db: db},
			actor:       actor.FromActualUser(&types.User{ID: otherUserID, SiteAdmin: true}),
			wantSubject: &settingsSubjectResolver{user: &UserResolver{user: &types.User{ID: userID}, db: db}},
		},
		{
			name:        "user settings - site admin on dotcom",
			node:        &UserResolver{user: &types.User{ID: userID}, db: db},
			actor:       actor.FromActualUser(&types.User{ID: otherUserID, SiteAdmin: true}),
			isDotcom:    true,
			wantSubject: &settingsSubjectResolver{user: &UserResolver{user: &types.User{ID: userID}, db: db}},
		},
		{
			name:      "user settings - different user",
			node:      &UserResolver{user: &types.User{ID: otherUserID}, db: db},
			actor:     actor.FromActualUser(&types.User{ID: userID}),
			wantError: &auth.InsufficientAuthorizationError{},
		},
		{
			name:        "org settings - member",
			node:        &OrgResolver{org: &types.Org{ID: orgID}, db: db},
			actor:       actor.FromActualUser(&types.User{ID: userID}),
			wantSubject: &settingsSubjectResolver{org: &OrgResolver{org: &types.Org{ID: orgID}, db: db}},
		},
		{
			name:      "org settings - non-member",
			node:      &OrgResolver{org: &types.Org{ID: orgID}, db: db},
			actor:     actor.FromActualUser(&types.User{ID: otherUserID}),
			wantError: auth.ErrNotAnOrgMember,
		},
		{
			name:        "org settings - non-member site admin",
			node:        &OrgResolver{org: &types.Org{ID: orgID}, db: db},
			actor:       actor.FromActualUser(&types.User{ID: otherUserID, SiteAdmin: true}),
			wantSubject: &settingsSubjectResolver{org: &OrgResolver{org: &types.Org{ID: orgID}, db: db}},
		},
		{
			name:        "org settings - non-member site admin on dotcom",
			node:        &OrgResolver{org: &types.Org{ID: orgID}, db: db},
			actor:       actor.FromActualUser(&types.User{ID: otherUserID, SiteAdmin: true}),
			isDotcom:    true,
			wantSubject: &settingsSubjectResolver{org: &OrgResolver{org: &types.Org{ID: orgID}, db: db}},
		},
		{
			name:      "unknown node type",
			node:      &mockNode{},
			actor:     actor.FromActualUser(&types.User{ID: userID}),
			wantError: errUnknownSettingsSubject,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dotcom.MockSourcegraphDotComMode(t, tc.isDotcom)

			actorUser, err := tc.actor.User(context.Background(), nil)
			if err != nil {
				panic(err)
			}
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(actorUser, nil)

			ctx := actor.WithActor(context.Background(), tc.actor)

			if tc.wantSubject != nil {
				tc.wantSubject.mockCheckedAccessForTest()
			}

			subject, err := settingsSubjectForNodeAndCheckAccess(ctx, tc.node)
			if tc.wantError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantError)
				}
				require.Error(t, err)
				require.IsType(t, tc.wantError, err)
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(subject, tc.wantSubject) {
					t.Fatalf("got %#v, want %#v", subject, tc.wantSubject)
				}
			}
		})
	}
}

type mockNode struct{}

func (m *mockNode) ID() graphql.ID {
	return "mock"
}
