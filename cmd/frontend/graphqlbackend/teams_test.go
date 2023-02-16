package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	actorID = 54123451
)

type fakeTeamsDb struct {
	database.TeamStore
	list       []*types.Team
	members    orderedTeamMembers
	users      *fakeUsersDB
	lastUsedID int32
}

func (teams *fakeTeamsDb) CreateTeam(_ context.Context, t *types.Team) error {
	teams.lastUsedID++
	u := *t
	u.ID = teams.lastUsedID
	teams.list = append(teams.list, &u)
	return nil
}

func (teams *fakeTeamsDb) UpdateTeam(_ context.Context, t *types.Team) error {
	if t == nil {
		return errors.New("UpdateTeam: team cannot be nil")
	}
	if t.ID == 0 {
		return errors.New("UpdateTeam: team.ID must be set (not 0)")
	}
	for _, u := range teams.list {
		if u.ID == t.ID {
			*u = *t
			return nil
		}
	}
	return errors.Newf("UpdateTeam: cannot find team with ID=%d", t.ID)
}

func (teams *fakeTeamsDb) GetTeamByID(_ context.Context, id int32) (*types.Team, error) {
	for _, t := range teams.list {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, database.TeamNotFoundError{}
}

func (teams *fakeTeamsDb) GetTeamByName(_ context.Context, name string) (*types.Team, error) {
	for _, t := range teams.list {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, database.TeamNotFoundError{}
}

func (teams *fakeTeamsDb) DeleteTeam(_ context.Context, id int32) error {
	for i, t := range teams.list {
		if t.ID == id {
			maxI := len(teams.list) - 1
			teams.list[i], teams.list[maxI] = teams.list[maxI], teams.list[i]
			teams.list = teams.list[:maxI]
			return nil
		}
	}
	return database.TeamNotFoundError{}
}

func (teams *fakeTeamsDb) ListTeams(_ context.Context, opts database.ListTeamsOpts) (selected []*types.Team, next int32, err error) {
	for _, t := range teams.list {
		if matches(t, opts) {
			selected = append(selected, t)
		}
	}
	if opts.LimitOffset != nil {
		selected = selected[opts.LimitOffset.Offset:]
		if limit := opts.LimitOffset.Limit; limit != 0 && len(selected) > limit {
			next = selected[opts.LimitOffset.Limit].ID
			selected = selected[:opts.LimitOffset.Limit]
		}
	}
	return selected, next, nil
}

func (teams *fakeTeamsDb) CountTeams(ctx context.Context, opts database.ListTeamsOpts) (int32, error) {
	selected, _, err := teams.ListTeams(ctx, opts)
	return int32(len(selected)), err
}

func matches(team *types.Team, opts database.ListTeamsOpts) bool {
	if opts.Cursor != 0 && team.ID < opts.Cursor {
		return false
	}
	if opts.WithParentID != 0 && team.ParentTeamID != opts.WithParentID {
		return false
	}
	if opts.RootOnly && team.ParentTeamID != 0 {
		return false
	}
	if opts.Search != "" {
		search := strings.ToLower(opts.Search)
		name := strings.ToLower(team.Name)
		displayName := strings.ToLower(team.DisplayName)
		if !strings.Contains(name, search) && !strings.Contains(displayName, search) {
			return false
		}
	}
	// opts.ForUserMember is not supported yet as there is no membership fake.
	return true
}

type orderedTeamMembers []*types.TeamMember

func (o orderedTeamMembers) Len() int { return len(o) }
func (o orderedTeamMembers) Less(i, j int) bool {
	if o[i].TeamID < o[j].TeamID {
		return true
	}
	if o[i].TeamID == o[j].TeamID {
		return o[i].UserID < o[j].UserID
	}
	return false
}
func (o orderedTeamMembers) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (teams *fakeTeamsDb) CountTeamMembers(ctx context.Context, opts database.ListTeamMembersOpts) (int32, error) {
	ms, _, err := teams.ListTeamMembers(ctx, opts)
	return int32(len(ms)), err
}

func (teams *fakeTeamsDb) ListTeamMembers(ctx context.Context, opts database.ListTeamMembersOpts) (selected []*types.TeamMember, next *database.TeamMemberListCursor, err error) {
	sort.Sort(teams.members)
	for _, m := range teams.members {
		if opts.Cursor.TeamID > m.TeamID {
			continue
		}
		if opts.Cursor.TeamID == m.TeamID && opts.Cursor.UserID > m.UserID {
			continue
		}
		if opts.TeamID != 0 && opts.TeamID != m.TeamID {
			continue
		}
		if opts.Search != "" {
			if teams.users == nil {
				return nil, nil, errors.New("fakeTeamsDB needs reference to fakeUsersDB for ListTeamMembersOpts.Search")
			}
			u, err := teams.users.GetByID(ctx, m.UserID)
			if err != nil {
				return nil, nil, err
			}
			if u == nil {
				continue
			}
			search := strings.ToLower(opts.Search)
			username := strings.ToLower(u.Username)
			displayName := strings.ToLower(u.DisplayName)
			if !strings.Contains(username, search) && !strings.Contains(displayName, search) {
				continue
			}
		}
		selected = append(selected, m)
	}
	if opts.LimitOffset != nil {
		selected = selected[opts.LimitOffset.Offset:]
		if limit := opts.LimitOffset.Limit; limit != 0 && len(selected) > limit {
			next = &database.TeamMemberListCursor{
				TeamID: selected[opts.LimitOffset.Limit].TeamID,
				UserID: selected[opts.LimitOffset.Limit].UserID,
			}
			selected = selected[:opts.LimitOffset.Limit]
		}
	}
	return selected, next, nil
}

func (teams *fakeTeamsDb) CreateTeamMember(ctx context.Context, members ...*types.TeamMember) error {
	for _, existingMember := range teams.members {
		for _, newMember := range members {
			if *existingMember == *newMember {
				return errors.Newf("Member teamID=%d userID=%d already exists.", newMember.TeamID, newMember.UserID)
			}
		}
	}
	teams.members = append(teams.members, members...)
	return nil
}

type fakeUsersDB struct {
	database.UserStore
	lastUserID int32
	list       []types.User
}

func fakeContext(u types.User) context.Context {
	return actor.WithActor(context.Background(), &actor.Actor{UID: u.ID})
}

func (users *fakeUsersDB) GetByID(_ context.Context, id int32) (*types.User, error) {
	for _, u := range users.list {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

func (users *fakeUsersDB) GetByCurrentAuthUser(ctx context.Context) (*types.User, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, database.ErrNoCurrentUser
	}
	return a.User(ctx, users)
}

func (users *fakeUsersDB) newUser(u types.User) int32 {
	id := users.lastUserID + 1
	users.lastUserID = id
	u.ID = id
	users.list = append(users.list, u)
	return id
}

var (
	db        *database.MockDB
	fakeTeams *fakeTeamsDb
	fakeUsers *fakeUsersDB
)

func setupDB() {
	fakeTeams = &fakeTeamsDb{}
	fakeUsers = &fakeUsersDB{}
	fakeTeams.users = fakeUsers
	db = database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	db.UsersFunc.SetDefaultReturn(fakeUsers)
	db.WithTransactFunc.SetDefaultHook(func(_ context.Context, callback func(database.DB) error) error {
		return callback(db)
	})
}

func userCtx(userID int32) context.Context {
	a := &actor.Actor{
		UID: userID,
	}
	return actor.WithActor(context.Background(), a)
}

func TestTeamNode(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create fake team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("failed to get fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query TeamByID($id: ID!){
			node(id: $id) {
				__typename
				... on Team {
				  name
				}
			}
		}`,
		ExpectedResult: `{
			"node": {
				"__typename": "Team",
				"name": "team"
			}
		}`,
		Variables: map[string]any{
			"id": string(relay.MarshalID("Team", team.ID)),
		},
	})
}

func TestTeamNodeURL(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	team := &types.Team{
		Name: "team-刺身", // team-sashimi
	}
	if err := fakeTeams.CreateTeam(ctx, team); err != nil {
		t.Fatalf("failed to create fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `{
			team(name: "team-刺身") {
				... on Team {
					url
				}
			}
		}`,
		ExpectedResult: `{
			"team": {
				"url": "/teams/team-%E5%88%BA%E8%BA%AB"
			}
		}`,
	})
}

func TestTeamNodeSiteAdminCanAdminister(t *testing.T) {
	for _, isAdmin := range []bool{true, false} {
		t.Run(fmt.Sprintf("viewer is admin = %v", isAdmin), func(t *testing.T) {
			setupDB()
			ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: isAdmin}))
			if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
				t.Fatalf("failed to create fake team: %s", err)
			}
			team, err := fakeTeams.GetTeamByName(ctx, "team")
			if err != nil {
				t.Fatalf("failed to get fake team: %s", err)
			}
			RunTest(t, &Test{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `query TeamByID($id: ID!){
					node(id: $id) {
						__typename
						... on Team {
							viewerCanAdminister
						}
					}
				}`,
				ExpectedResult: fmt.Sprintf(`{
					"node": {
						"__typename": "Team",
						"viewerCanAdminister": %v
					}
				}`, isAdmin),
				Variables: map[string]any{
					"id": string(relay.MarshalID("Team", team.ID)),
				},
			})
		})
	}
}

func TestCreateTeamBare(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!) {
			createTeam(name: $name) {
				name
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"name": "team-name-testing"
			}
		}`,
		Variables: map[string]any{
			"name": "team-name-testing",
		},
	})
	expected := &types.Team{
		ID:        1,
		Name:      "team-name-testing",
		CreatorID: actor.FromContext(ctx).UID,
	}
	if diff := cmp.Diff([]*types.Team{expected}, fakeTeams.list); diff != "" {
		t.Errorf("unexpected teams in fake database (-want,+got):\n%s", diff)
	}
}

func TestCreateTeamDisplayName(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $displayName: String!) {
			createTeam(name: $name, displayName: $displayName) {
				displayName
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"displayName": "Team Display Name"
			}
		}`,
		Variables: map[string]any{
			"name":        "team-name-testing",
			"displayName": "Team Display Name",
		},
	})
}

func TestCreateTeamReadOnlyDefault(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!) {
			createTeam(name: $name) {
				readonly
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"readonly": false
			}
		}`,
		Variables: map[string]any{
			"name": "team-name-testing",
		},
	})
}

func TestCreateTeamReadOnlyTrue(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $readonly: Boolean!) {
			createTeam(name: $name, readonly: $readonly) {
				readonly
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"readonly": true
			}
		}`,
		Variables: map[string]any{
			"name":     "team-name-testing",
			"readonly": true,
		},
	})
}

func TestCreateTeamParentByID(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	err := fakeTeams.CreateTeam(ctx, &types.Team{
		Name: "team-name-parent",
	})
	if err != nil {
		t.Fatal(err)
	}
	parentTeam, err := fakeTeams.GetTeamByName(ctx, "team-name-parent")
	if err != nil {
		t.Fatal(err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $parentTeamID: ID!) {
			createTeam(name: $name, parentTeam: $parentTeamID) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"parentTeam": {
					"name": "team-name-parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":         "team-name-testing",
			"parentTeamID": string(relay.MarshalID("Team", parentTeam.ID)),
		},
	})
}

func TestCreateTeamParentByName(t *testing.T) {
	setupDB()
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := fakeTeams.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $parentTeamName: String!) {
			createTeam(name: $name, parentTeamName: $parentTeamName) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"parentTeam": {
					"name": "team-name-parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":           "team-name-testing",
			"parentTeamName": "team-name-parent",
		},
	})
}

func TestUpdateTeamByID(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team-name-testing")
	if err != nil {
		t.Fatalf("failed to get fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($id: ID!, $newDisplayName: String!) {
			updateTeam(id: $id, displayName: $newDisplayName) {
				displayName
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"displayName": "Updated Display Name"
			}
		}`,
		Variables: map[string]any{
			"id":             string(relay.MarshalID("Team", team.ID)),
			"newDisplayName": "Updated Display Name",
		},
	})
	wantTeams := []*types.Team{
		{
			ID:          1,
			Name:        "team-name-testing",
			DisplayName: "Updated Display Name",
		},
	}
	if diff := cmp.Diff(wantTeams, fakeTeams.list); diff != "" {
		t.Errorf("fake teams storage (-want,+got):\n%s", diff)
	}
}

func TestUpdateTeamByName(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newDisplayName: String!) {
			updateTeam(name: $name, displayName: $newDisplayName) {
				displayName
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"displayName": "Updated Display Name"
			}
		}`,
		Variables: map[string]any{
			"name":           "team-name-testing",
			"newDisplayName": "Updated Display Name",
		},
	})
	wantTeams := []*types.Team{
		{
			ID:          1,
			Name:        "team-name-testing",
			DisplayName: "Updated Display Name",
		},
	}
	if diff := cmp.Diff(wantTeams, fakeTeams.list); diff != "" {
		t.Errorf("fake teams storage (-want,+got):\n%s", diff)
	}
}

func TestUpdateTeamErrorBothNameAndID(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team-name-testing")
	if err != nil {
		t.Fatalf("failed to get fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $id: ID!, $newDisplayName: String!) {
			updateTeam(name: $name, id: $id, displayName: $newDisplayName) {
				displayName
			}
		}`,
		ExpectedResult: "null",
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "team to update is identified by either id or name, but both were specified",
				Path:    []any{"updateTeam"},
			},
		},
		Variables: map[string]any{
			"id":             string(relay.MarshalID("Team", team.ID)),
			"name":           "team-name-testing",
			"newDisplayName": "Updated Display Name",
		},
	})
}

func TestUpdateParentByID(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parentTeam, err := fakeTeams.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("failed to fetch fake parent team: %s", err)
	}
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newParentID: ID!) {
			updateTeam(name: $name, parentTeam: $newParentID) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"parentTeam": {
					"name": "parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":        "team",
			"newParentID": string(relay.MarshalID("Team", parentTeam.ID)),
		},
	})
}

func TestUpdateParentByName(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newParentName: String!) {
			updateTeam(name: $name, parentTeamName: $newParentName) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"parentTeam": {
					"name": "parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":          "team",
			"newParentName": "parent",
		},
	})
}

func TestUpdateParentErrorBothNameAndID(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parentTeam, err := fakeTeams.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("failed to fetch fake parent team: %s", err)
	}
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newParentID: ID!, $newParentName: String!) {
			updateTeam(name: $name, parentTeam: $newParentID, parentTeamName: $newParentName) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: "null",
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "parent team is identified by either id or name, but both were specified",
				Path:    []any{"updateTeam"},
			},
		},
		Variables: map[string]any{
			"name":          "team",
			"newParentID":   string(relay.MarshalID("Team", parentTeam.ID)),
			"newParentName": parentTeam.Name,
		},
	})
}

func TestDeleteTeamByID(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("cannot find fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($id: ID!) {
			deleteTeam(id: $id) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": {
				"alwaysNil": null
			}
		}`,
		Variables: map[string]any{
			"id": string(relay.MarshalID("Team", team.ID)),
		},
	})
	if diff := cmp.Diff([]*types.Team{}, fakeTeams.list); diff != "" {
		t.Errorf("expected no teams in fake db after deleting, (-want,+got):\n%s", diff)
	}
}

func TestDeleteTeamByName(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($name: String!) {
			deleteTeam(name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": {
				"alwaysNil": null
			}
		}`,
		Variables: map[string]any{
			"name": "team",
		},
	})
	if diff := cmp.Diff([]*types.Team{}, fakeTeams.list); diff != "" {
		t.Errorf("expected no teams in fake db after deleting, (-want,+got):\n%s", diff)
	}
}

func TestDeleteTeamErrorBothIDAndNameGiven(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("cannot find fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($id: ID!, $name: String!) {
			deleteTeam(id: $id, name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "team to delete is identified by either id or name, but both were specified",
				Path:    []any{"deleteTeam"},
			},
		},
		Variables: map[string]any{
			"id":   string(relay.MarshalID("Team", team.ID)),
			"name": "team",
		},
	})
}

func TestDeleteTeamNoIdentifierGiven(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam() {
			deleteTeam() {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "team to delete is identified by either id or name, but neither was specified",
				Path:    []any{"deleteTeam"},
			},
		},
	})
}

func TestDeleteTeamNotFound(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($name: String!) {
			deleteTeam(name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: `team name="does-not-exist" not found: team not found: <nil>`,
				Path:    []any{"deleteTeam"},
			},
		},
		Variables: map[string]any{
			"name": "does-not-exist",
		},
	})
}

func TestDeleteTeamUnauthorized(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: false}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($name: String!) {
			deleteTeam(name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "only site admins can delete teams",
				Path:    []any{"deleteTeam"},
			},
		},
		Variables: map[string]any{
			"name": "team",
		},
	})
}

func TestTeamByName(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query Team($name: String!) {
			team(name: $name) {
				name
			}
		}`,
		ExpectedResult: `{
			"team": {
				"name": "team"
			}
		}`,
		Variables: map[string]any{
			"name": "team",
		},
	})
}

func TestTeamByNameNotFound(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query Team($name: String!) {
			team(name: $name) {
				name
			}
		}`,
		ExpectedResult: `{
			"team": null
		}`,
		Variables: map[string]any{
			"name": "does-not-exist",
		},
	})
}

func TestTeamByNameUnauthorized(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: false}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query Team($name: String!) {
			team(name: $name) {
				id
			}
		}`,
		ExpectedResult: `{
			"team": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "only site admins can view teams",
				Path:    []any{"team"},
			},
		},
		Variables: map[string]any{
			"name": "team",
		},
	})
}

func TestTeamsPaginated(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	for i := 1; i <= 25; i++ {
		name := fmt.Sprintf("team-%d", i)
		if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
			t.Fatalf("failed to create a team: %s", err)
		}
	}
	var (
		hasNextPage bool = true
		cursor      string
	)
	query := `query Teams($cursor: String!) {
		teams(after: $cursor, first: 10) {
			pageInfo {
				endCursor
				hasNextPage
			}
			nodes {
				name
			}
		}
	}`
	operationName := ""
	var gotNames []string
	for hasNextPage {
		variables := map[string]any{
			"cursor": cursor,
		}
		r := mustParseGraphQLSchema(t, db).Exec(ctx, query, operationName, variables)
		var wantErrors []*gqlerrors.QueryError
		checkErrors(t, wantErrors, r.Errors)
		var result struct {
			Teams *struct {
				PageInfo *struct {
					EndCursor   string
					HasNextPage bool
				}
				Nodes []struct {
					Name string
				}
			}
		}
		if err := json.Unmarshal(r.Data, &result); err != nil {
			t.Fatalf("cannot interpret graphQL query result: %s", err)
		}
		hasNextPage = result.Teams.PageInfo.HasNextPage
		cursor = result.Teams.PageInfo.EndCursor
		for _, node := range result.Teams.Nodes {
			gotNames = append(gotNames, node.Name)
		}
	}
	var wantNames []string
	for _, team := range fakeTeams.list {
		wantNames = append(wantNames, team.Name)
	}
	if diff := cmp.Diff(wantNames, gotNames); diff != "" {
		t.Errorf("unexpected team names (-want,+got):\n%s", diff)
	}
}

// Skip testing DisplayName search as this is the same except the fake behavior.
func TestTeamsNameSearch(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	for _, name := range []string{"hit-1", "Hit-2", "HIT-3", "miss-4", "mIss-5", "MISS-6"} {
		if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
			t.Fatalf("failed to create a team: %s", err)
		}
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `{
			teams(search: "hit") {
				nodes {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"teams": {
				"nodes": [
					{"name": "hit-1"},
					{"name": "Hit-2"},
					{"name": "HIT-3"}
				]
			}
		}`,
	})
}

func TestTeamsCount(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	for i := 1; i <= 25; i++ {
		name := fmt.Sprintf("team-%d", i)
		if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
			t.Fatalf("failed to create a team: %s", err)
		}
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `{
			teams(first: 5) {
				totalCount
				nodes {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"teams": {
				"totalCount": 25,
				"nodes": [
					{"name": "team-1"},
					{"name": "team-2"},
					{"name": "team-3"},
					{"name": "team-4"},
					{"name": "team-5"}
				]
			}
		}`,
	})
}

func TestChildTeams(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parent, err := fakeTeams.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("cannot fetch parent team: %s", err)
	}
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("child-%d", i)
		if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: name, ParentTeamID: parent.ID}); err != nil {
			t.Fatalf("cannot create child team: %s", err)
		}
	}
	for i := 6; i <= 10; i++ {
		name := fmt.Sprintf("not-child-%d", i)
		if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
			t.Fatalf("cannot create a team: %s", err)
		}
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `{
			team(name: "parent") {
				childTeams {
					nodes {
						name
					}
				}
			}
		}`,
		ExpectedResult: `{
			"team": {
				"childTeams": {
					"nodes": [
						{"name": "child-1"},
						{"name": "child-2"},
						{"name": "child-3"},
						{"name": "child-4"},
						{"name": "child-5"}
					]
				}
			}
		}`,
	})
}

func TestMembersPaginated(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team-with-members"}); err != nil {
		t.Fatalf("failed to create team: %s", err)
	}
	teamWithMembers, err := fakeTeams.GetTeamByName(ctx, "team-with-members")
	if err != nil {
		t.Fatalf("failed to featch fake team: %s", err)
	}
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "different-team"}); err != nil {
		t.Fatalf("failed to create team: %s", err)
	}
	differentTeam, err := fakeTeams.GetTeamByName(ctx, "different-team")
	if err != nil {
		t.Fatalf("failed to featch fake team: %s", err)
	}
	for _, team := range []*types.Team{teamWithMembers, differentTeam} {
		for i := 1; i <= 25; i++ {
			id := fakeUsers.newUser(types.User{Username: fmt.Sprintf("user-%d-%d", team.ID, i)})
			m := &types.TeamMember{
				TeamID: team.ID,
				UserID: id,
			}
			fakeTeams.members = append(fakeTeams.members, m)
		}
	}
	var (
		hasNextPage bool = true
		cursor      string
	)
	query := `query Members($cursor: String!) {
		team(name: "team-with-members") {
			members(after: $cursor, first: 10) {
				totalCount
				pageInfo {
					endCursor
					hasNextPage
				}
				nodes {
					... on User {
						username
					}
				}
			}
		}
	}`
	operationName := ""
	var gotUsernames []string
	for hasNextPage {
		variables := map[string]any{
			"cursor": cursor,
		}
		r := mustParseGraphQLSchema(t, db).Exec(ctx, query, operationName, variables)
		var wantErrors []*gqlerrors.QueryError
		checkErrors(t, wantErrors, r.Errors)
		var result struct {
			Team *struct {
				Members *struct {
					TotalCount int
					PageInfo   *struct {
						EndCursor   string
						HasNextPage bool
					}
					Nodes []struct {
						Username string
					}
				}
			}
		}
		if err := json.Unmarshal(r.Data, &result); err != nil {
			t.Fatalf("cannot interpret graphQL query result: %s", err)
		}
		if got, want := result.Team.Members.TotalCount, 25; got != want {
			t.Errorf("totalCount, got %d, want %d", got, want)
		}
		if got, want := len(result.Team.Members.Nodes), 10; got > want {
			t.Errorf("#nodes, got %d, want at most %d", got, want)
		}
		hasNextPage = result.Team.Members.PageInfo.HasNextPage
		cursor = result.Team.Members.PageInfo.EndCursor
		for _, node := range result.Team.Members.Nodes {
			gotUsernames = append(gotUsernames, node.Username)
		}
	}
	var wantUsernames []string
	for i := 1; i <= 25; i++ {
		wantUsernames = append(wantUsernames, fmt.Sprintf("user-%d-%d", teamWithMembers.ID, i))
	}
	if diff := cmp.Diff(wantUsernames, gotUsernames); diff != "" {
		t.Errorf("unexpected member usernames (-want,+got):\n%s", diff)
	}
}

func TestMembersSearch(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("failed to fetch fake team by ID: %s", err)
	}
	for _, u := range []types.User{
		{
			Username: "username-hit",
		},
		{
			Username: "username-miss",
		},
		{
			Username:    "look-at-displayname",
			DisplayName: "Display Name Hit",
		},
	} {
		userID := fakeUsers.newUser(u)
		fakeTeams.members = append(fakeTeams.members, &types.TeamMember{
			TeamID: team.ID,
			UserID: userID,
		})
	}
	idOfMissingUser := -7
	fakeTeams.members = append(fakeTeams.members, &types.TeamMember{
		TeamID: team.ID,
		UserID: int32(idOfMissingUser),
	})
	fakeUsers.newUser(types.User{Username: "search-hit-but-not-team-member"})
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `{
			team(name: "team") {
				members(search: "hit") {
					nodes {
						... on User {
							username
						}
					}
				}
			}
		}`,
		ExpectedResult: `{
			"team": {
				"members": {
					"nodes": [
						{"username": "username-hit"},
						{"username": "look-at-displayname"}
					]
				}
			}
		}`,
	})
}

func TestMembersAdd(t *testing.T) {
	setupDB()
	ctx := userCtx(fakeUsers.newUser(types.User{SiteAdmin: true}))
	if err := fakeTeams.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	team, err := fakeTeams.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("cannot fetch parent team: %s", err)
	}
	userExistingID := fakeUsers.newUser(types.User{Username: "existing"})
	userExistingAndAddedID := fakeUsers.newUser(types.User{Username: "existingAndAdded"})
	userAddedID := fakeUsers.newUser(types.User{Username: "added"})
	fakeTeams.members = append(fakeTeams.members,
		&types.TeamMember{TeamID: team.ID, UserID: userExistingID},
		&types.TeamMember{TeamID: team.ID, UserID: userExistingAndAddedID},
	)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation AddTeamMembers($existingAndAddedId: ID!, $addedId: ID!) {
			addTeamMembers(teamName: "team", members: [
				$existingAndAddedId,
				$addedId
			]) {
				members {
					nodes {
						... on User {
							username
						}
					}
				}
			}
		}`,
		ExpectedResult: `{
			"addTeamMembers": {
				"members": {
					"nodes": [
						{"username": "existing"},
						{"username": "existingAndAdded"},
						{"username": "added"}
					]
				}
			}
		}`,
		Variables: map[string]any{
			"existingAndAddedId": string(relay.MarshalID("TeamMember", userExistingAndAddedID)),
			"addedId":            string(relay.MarshalID("TeamMember", userAddedID)),
		},
	})
}
