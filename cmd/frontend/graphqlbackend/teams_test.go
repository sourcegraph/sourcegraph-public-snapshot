package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	t.ID = teams.lastUsedID
	teams.list = append(teams.list, t)
	return nil
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

func TestCreateTeamBare(t *testing.T) {
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	ctx, user, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!) {
			createTeam(name: $name) {
				id
				name
			}
		}`,
		ExpectedResult: fmt.Sprintf(`{
			"createTeam": {
				"id": %q,
				"name": "team-name-testing"
			}
		}`, relay.MarshalID("Team", 1)),
		Variables: map[string]any{
			"name": "team-name-testing",
		},
	})
	expected := &types.Team{
		ID:        1,
		Name:      "team-name-testing",
		CreatorID: user.ID,
	}
	if diff := cmp.Diff([]*types.Team{expected}, fakeTeams.list); diff != "" {
		t.Errorf("unexpected teams in fake database (-want,+got):\n%s", diff)
	}
}

func TestCreateTeamDisplayName(t *testing.T) {
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
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
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
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
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
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
	fakeTeams := &fakeTeamsDb{}
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := fakeTeams.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
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
	fakeTeams := &fakeTeamsDb{}
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := fakeTeams.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
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
