package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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

type teamResolver struct{}

func (r *teamResolver) ID() graphql.ID            { return "" }
func (r *teamResolver) Name() string              { return "" }
func (r *teamResolver) URL() string               { return "" }
func (r *teamResolver) DisplayName() *string      { return nil }
func (r *teamResolver) Readonly() bool            { return false }
func (r *teamResolver) ParentTeam() *teamResolver { return nil }
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

func (r *schemaResolver) CreateTeam(args *CreateTeamArgs) *teamResolver {
	return &teamResolver{}
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
