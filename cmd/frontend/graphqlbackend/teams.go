package graphqlbackend

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ListTeamsArgs struct {
	First  *int32
	After  *string
	Search *string
}

type teamConnectionResolver struct{}

func (r *teamConnectionResolver) TotalCount(args *struct{ CountDeeplyNestedTeams bool }) int32 {
	return 0
}
func (r *teamConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}
func (r *teamConnectionResolver) Nodes() []*teamResolver { return nil }

type teamResolver struct {
	team    *types.Team
	teamsDb database.TeamStore
}

// TODO: handle nil teamResolver gracefully in case there is no team
// like in querying parent of a hierarnchy root.
func (r *teamResolver) ID() graphql.ID {
	intID := int(r.team.ID)
	stringID := strconv.Itoa(intID)
	return graphql.ID(stringID)
}
func (r *teamResolver) Name() string { return r.team.Name }
func (r *teamResolver) URL() string  { return "" }
func (r *teamResolver) DisplayName() *string {
	if r.team.DisplayName == "" {
		return nil
	}
	return &r.team.DisplayName
}
func (r *teamResolver) Readonly() bool { return r.team.ReadOnly }
func (r *teamResolver) ParentTeam(ctx context.Context) (*teamResolver, error) {
	if r.team.ParentTeamID == 0 {
		// TODO: How to handle this gracefully for query?
		return nil, errors.New("no parent team")
	}
	parentTeam, err := r.teamsDb.GetTeamByID(ctx, r.team.ParentTeamID)
	if err != nil {
		return nil, err
	}
	return &teamResolver{team: parentTeam, teamsDb: r.teamsDb}, nil
}
func (r *teamResolver) ViewerCanAdminister() bool { return false }
func (r *teamResolver) Members(args *ListTeamsArgs) *teamMemberConnection {
	return &teamMemberConnection{}
}
func (r *teamResolver) ChildTeams(args *ListTeamsArgs) *teamConnectionResolver {
	return &teamConnectionResolver{}
}

type teamMemberConnection struct{}

func (r *teamMemberConnection) TotalCount(args *struct{ CountDeeplyNestedTeamMembers bool }) int32 {
	return 0
}
func (r *teamMemberConnection) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}
func (r *teamMemberConnection) Nodes() []*UserResolver {
	return nil
}

type CreateTeamArgs struct {
	Name           string
	DisplayName    *string
	ReadOnly       bool
	ParentTeam     *graphql.ID
	ParentTeamName *string
}

func (r *schemaResolver) CreateTeam(ctx context.Context, args *CreateTeamArgs) (*teamResolver, error) {
	teams := r.db.Teams()
	var t types.Team
	t.Name = args.Name
	if args.DisplayName != nil {
		t.DisplayName = *args.DisplayName
	}
	t.ReadOnly = args.ReadOnly
	if args.ParentTeam != nil && args.ParentTeamName != nil {
		return nil, errors.New("must specify at most one: ParentTeam or ParentTeamName")
	}
	if args.ParentTeam != nil {
		id, err := strconv.Atoi(string(*args.ParentTeam))
		if err != nil {
			return nil, errors.Wrapf(err, "ParentTeam ID does not conform to integer: %q", *args.ParentTeam)
		}
		t.ParentTeamID = int32(id)
		if id != int(t.ParentTeamID) {
			return nil, errors.Wrapf(err, "ParentTeam ID does not conform to int32: %q", *args.ParentTeam)
		}
	}
	if args.ParentTeamName != nil {
		parentTeam, err := teams.GetTeamByName(ctx, *args.ParentTeamName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot locate parent team %q", *args.ParentTeamName)
		}
		t.ParentTeamID = parentTeam.ID
	}
	t.CreatorID = actor.FromContext(ctx).UID
	if err := teams.CreateTeam(ctx, &t); err != nil {
		return nil, err
	}
	return &teamResolver{team: &t, teamsDb: teams}, nil
}

type UpdateTeamArgs struct {
	ID             *graphql.ID
	Name           *string
	DisplayName    *string
	ParentTeam     *graphql.ID
	ParentTeamName *string
}

func (r *schemaResolver) UpdateTeam(args *UpdateTeamArgs) *teamResolver {
	return &teamResolver{}
}

type DeleteTeamArgs struct {
	ID   *graphql.ID
	Name *string
}

func (r *schemaResolver) DeleteTeam(args *DeleteTeamArgs) *EmptyResponse {
	return &EmptyResponse{}
}

type TeamMembersArgs struct {
	Team     *graphql.ID
	TeamName *string
	Members  []graphql.ID
}

func (r *schemaResolver) AddTeamMembers(args *TeamMembersArgs) *teamResolver {
	return &teamResolver{}
}

func (r *schemaResolver) SetTeamMembers(args *TeamMembersArgs) *teamResolver {
	return &teamResolver{}
}

func (r *schemaResolver) RemoveTeamMembers(args *TeamMembersArgs) *teamResolver {
	return &teamResolver{}
}
