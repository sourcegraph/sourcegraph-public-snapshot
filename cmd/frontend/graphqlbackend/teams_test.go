package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"

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

func setupDB() (*database.MockDB, *fakeTeamsDb) {
	ts := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(ts)
	db.WithTransactFunc.SetDefaultHook(func(_ context.Context, callback func(database.DB) error) error {
		return callback(db)
	})
	return db, ts
}

func TestTeamNode(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create fake team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team")
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

func TestTeamNodeSiteAdminCanAdminister(t *testing.T) {
	for _, isAdmin := range []bool{true, false} {
		t.Run(fmt.Sprintf("viewer is admin = %v", isAdmin), func(t *testing.T) {
			db, ts := setupDB()
			ctx, _, _ := fakeUser(t, context.Background(), db, isAdmin)
			if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
				t.Fatalf("failed to create fake team: %s", err)
			}
			team, err := ts.GetTeamByName(ctx, "team")
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
	db, ts := setupDB()
	ctx, user, _ := fakeUser(t, context.Background(), db, true)
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
		CreatorID: user.ID,
	}
	if diff := cmp.Diff([]*types.Team{expected}, ts.list); diff != "" {
		t.Errorf("unexpected teams in fake database (-want,+got):\n%s", diff)
	}
}

func TestCreateTeamDisplayName(t *testing.T) {
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	err := ts.CreateTeam(ctx, &types.Team{
		Name: "team-name-parent",
	})
	if err != nil {
		t.Fatal(err)
	}
	parentTeam, err := ts.GetTeamByName(ctx, "team-name-parent")
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
	db, ts := setupDB()
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := ts.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team-name-testing")
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
	if diff := cmp.Diff(wantTeams, ts.list); diff != "" {
		t.Errorf("fake teams storage (-want,+got):\n%s", diff)
	}
}

func TestUpdateTeamByName(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{
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
	if diff := cmp.Diff(wantTeams, ts.list); diff != "" {
		t.Errorf("fake teams storage (-want,+got):\n%s", diff)
	}
}

func TestUpdateTeamErrorBothNameAndID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team-name-testing")
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parentTeam, err := ts.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("failed to fetch fake parent team: %s", err)
	}
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parentTeam, err := ts.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("failed to fetch fake parent team: %s", err)
	}
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team")
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
	if diff := cmp.Diff([]*types.Team{}, ts.list); diff != "" {
		t.Errorf("expected no teams in fake db after deleting, (-want,+got):\n%s", diff)
	}
}

func TestDeleteTeamByName(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	if diff := cmp.Diff([]*types.Team{}, ts.list); diff != "" {
		t.Errorf("expected no teams in fake db after deleting, (-want,+got):\n%s", diff)
	}
}

func TestDeleteTeamErrorBothIDAndNameGiven(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team")
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
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, ts := setupDB()
	// false in the next line indicates not-site-admin
	ctx, _, _ := fakeUser(t, context.Background(), db, false)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	db, ts := setupDB()
	// false in the next line indicates not-site-admin
	ctx, _, _ := fakeUser(t, context.Background(), db, false)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	for i := 1; i <= 25; i++ {
		name := fmt.Sprintf("team-%d", i)
		if err := ts.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
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
	for _, team := range ts.list {
		wantNames = append(wantNames, team.Name)
	}
	if diff := cmp.Diff(wantNames, gotNames); diff != "" {
		t.Errorf("unexpected team names (-want,+got):\n%s", diff)
	}
}

// Skip testing DisplayName search as this is the same except the fake behavior.
func TestTeamsNameSearch(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	for _, name := range []string{"hit-1", "Hit-2", "HIT-3", "miss-4", "mIss-5", "MISS-6"} {
		if err := ts.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	for i := 1; i <= 25; i++ {
		name := fmt.Sprintf("team-%d", i)
		if err := ts.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
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
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parent, err := ts.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("cannot fetch parent team: %s", err)
	}
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("child-%d", i)
		if err := ts.CreateTeam(ctx, &types.Team{Name: name, ParentTeamID: parent.ID}); err != nil {
			t.Fatalf("cannot create child team: %s", err)
		}
	}
	for i := 6; i <= 10; i++ {
		name := fmt.Sprintf("not-child-%d", i)
		if err := ts.CreateTeam(ctx, &types.Team{Name: name}); err != nil {
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
