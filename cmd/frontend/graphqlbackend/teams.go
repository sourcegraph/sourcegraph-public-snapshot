package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
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
	team *types.Team
	db   database.DB
}

func (r *teamResolver) ID() graphql.ID {
	return relay.MarshalID("Team", r.team.ID)
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
		return nil, nil
	}
	parentTeam, err := r.db.Teams().GetTeamByID(ctx, r.team.ParentTeamID)
	if err != nil {
		return nil, err
	}
	return &teamResolver{team: parentTeam, db: r.db}, nil
}
func (r *teamResolver) ViewerCanAdminister(ctx context.Context) bool {
	// 🚨 SECURITY: For now administration is only allowed for site admins.
	err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	return err == nil
}
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
	// 🚨 SECURITY: For now we only allow site admins to create teams.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("only site admins can create teams")
	}
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
	parentTeam, err := findTeam(ctx, teams, args.ParentTeam, args.ParentTeamName)
	if err != nil {
		return nil, errors.Wrap(err, "parent team")
	}
	if parentTeam != nil {
		t.ParentTeamID = parentTeam.ID
	}
	t.CreatorID = actor.FromContext(ctx).UID
	if err := teams.CreateTeam(ctx, &t); err != nil {
		return nil, err
	}
	return &teamResolver{team: &t, db: r.db}, nil
}

type UpdateTeamArgs struct {
	ID             *graphql.ID
	Name           *string
	DisplayName    *string
	ParentTeam     *graphql.ID
	ParentTeamName *string
}

func (r *schemaResolver) UpdateTeam(ctx context.Context, args *UpdateTeamArgs) (*teamResolver, error) {
	// 🚨 SECURITY: For now we only allow site admins to create teams.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("only site admins can update teams")
	}
	if args.ID == nil && args.Name == nil {
		return nil, errors.New("team to update is identified by either id or name, but neither was specified")
	}
	if args.ID != nil && args.Name != nil {
		return nil, errors.New("team to update is identified by either id or name, but both were specified")
	}
	if args.ParentTeam != nil && args.ParentTeamName != nil {
		return nil, errors.New("parent team is identified by either id or name, but both were specified")
	}
	var t *types.Team
	err := r.db.WithTransact(ctx, func(tx database.DB) (err error) {
		t, err = findTeam(ctx, tx.Teams(), args.ID, args.Name)
		if err != nil {
			return err
		}
		var needsUpdate bool
		if args.DisplayName != nil && *args.DisplayName != t.DisplayName {
			needsUpdate = true
			t.DisplayName = *args.DisplayName
		}
		if args.ParentTeam != nil || args.ParentTeamName != nil {
			parentTeam, err := findTeam(ctx, tx.Teams(), args.ParentTeam, args.ParentTeamName)
			if err != nil {
				return errors.Wrap(err, "cannot find parent team")
			}
			if parentTeam.ID != t.ParentTeamID {
				needsUpdate = true
				t.ParentTeamID = parentTeam.ID
			}
		}
		if needsUpdate {
			return tx.Teams().UpdateTeam(ctx, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &teamResolver{team: t, db: r.db}, nil
}

// findTeam returns a team by either GraphQL ID or name.
// If both parameters are nil, the result is nil.
func findTeam(ctx context.Context, teams database.TeamStore, graphqlID *graphql.ID, name *string) (*types.Team, error) {
	if graphqlID != nil {
		var id int32
		err := relay.UnmarshalSpec(*graphqlID, &id)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot interpret team id: %q", *graphqlID)
		}
		team, err := teams.GetTeamByID(ctx, id)
		if errcode.IsNotFound(err) {
			return nil, errors.Wrapf(err, "team id=%d not found", id)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching team id=%d", id)
		}
		return team, nil
	}
	if name != nil {
		team, err := teams.GetTeamByName(ctx, *name)
		if errcode.IsNotFound(err) {
			return nil, errors.Wrapf(err, "team name=%q not found", *name)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "could not fetch team name=%q", *name)
		}
		return team, nil
	}
	return nil, nil
}

type DeleteTeamArgs struct {
	ID   *graphql.ID
	Name *string
}

func (r *schemaResolver) DeleteTeam(ctx context.Context, args *DeleteTeamArgs) (*EmptyResponse, error) {
	// 🚨 SECURITY: For now we only allow site admins to create teams.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("only site admins can delete teams")
	}
	if args.ID == nil && args.Name == nil {
		return nil, errors.New("team to delete is identified by either id or name, but neither was specified")
	}
	if args.ID != nil && args.Name != nil {
		return nil, errors.New("team to delete is identified by either id or name, but both were specified")
	}
	t, err := findTeam(ctx, r.db.Teams(), args.ID, args.Name)
	if err != nil {
		return nil, err
	}
	if err := r.db.Teams().DeleteTeam(ctx, t.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
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

func (r *schemaResolver) Teams(ctx context.Context, args *ListTeamsArgs) (*teamConnectionResolver, error) {
	return &teamConnectionResolver{}, nil
}

type TeamArgs struct {
	Name string
}

func (r *schemaResolver) Team(ctx context.Context, args *TeamArgs) (*teamResolver, error) {
	// 🚨 SECURITY: For now we only allow site admins to use teams.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("only site admins can view teams")
	}

	t, err := r.db.Teams().GetTeamByName(ctx, args.Name)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &teamResolver{db: r.db, team: t}, nil
}
