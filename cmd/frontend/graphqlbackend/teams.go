package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

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

type teamConnectionResolver struct {
	db       database.DB
	parentID int32
	search   string
	cursor   int32
	limit    int
	once     sync.Once
	teams    []*types.Team
	pageInfo *graphqlutil.PageInfo
	err      error
}

// applyArgs unmarshals query conditions and limites set in `ListTeamsArgs`
// into `teamConnectionResolver` fields for convenient use in database query.
func (r *teamConnectionResolver) applyArgs(args *ListTeamsArgs) error {
	if args.After != nil {
		cursor, err := graphqlutil.DecodeIntCursor(args.After)
		if err != nil {
			return err
		}
		r.cursor = int32(cursor)
		if int(r.cursor) != cursor {
			return errors.Newf("cursor int32 overflow: %d", cursor)
		}
	}
	if args.Search != nil {
		r.search = *args.Search
	}
	if args.First != nil {
		r.limit = int(*args.First)
	}
	return nil
}

// compute resolves teams queried for this resolver.
// The result of running it is setting `teams`, `next` and `err`
// fields on the resolver. This ensures that resolving multiple
// graphQL attributes that require listing (like `pageInfo` and `nodes`)
// results in just one query.
func (r *teamConnectionResolver) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := database.ListTeamsOpts{
			Cursor:       r.cursor,
			WithParentID: r.parentID,
			Search:       r.search,
		}
		if r.limit != 0 {
			opts.LimitOffset = &database.LimitOffset{Limit: r.limit}
		}
		teams, next, err := r.db.Teams().ListTeams(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.teams = teams
		if next > 0 {
			r.pageInfo = graphqlutil.EncodeIntCursor(&next)
		} else {
			r.pageInfo = graphqlutil.HasNextPage(false)
		}
	})
}

func (r *teamConnectionResolver) TotalCount(ctx context.Context, args *struct{ CountDeeplyNestedTeams bool }) (int32, error) {
	if args != nil && args.CountDeeplyNestedTeams {
		return 0, errors.New("Not supported: counting deeply nested teams.")
	}
	// Not taking into account limit or cursor for count.
	opts := database.ListTeamsOpts{
		WithParentID: r.parentID,
		Search:       r.search,
	}
	return r.db.Teams().CountTeams(ctx, opts)
}

func (r *teamConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	r.compute(ctx)
	return r.pageInfo, r.err
}

func (r *teamConnectionResolver) Nodes(ctx context.Context) ([]*TeamResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	var rs []*TeamResolver
	for _, t := range r.teams {
		rs = append(rs, &TeamResolver{
			db:   r.db,
			team: t,
		})
	}
	return rs, nil
}

type TeamResolver struct {
	db   database.DB
	team *types.Team
}

func (r *TeamResolver) ID() graphql.ID {
	return relay.MarshalID("Team", r.team.ID)
}

func (r *TeamResolver) Name() string {
	return r.team.Name
}

func (r *TeamResolver) URL() string {
	absolutePath := fmt.Sprintf("/teams/%s", r.team.Name)
	u := &url.URL{Path: absolutePath}
	return u.String()
}

func (r *TeamResolver) DisplayName() *string {
	if r.team.DisplayName == "" {
		return nil
	}
	return &r.team.DisplayName
}

func (r *TeamResolver) Readonly() bool {
	return r.team.ReadOnly
}

func (r *TeamResolver) ParentTeam(ctx context.Context) (*TeamResolver, error) {
	if r.team.ParentTeamID == 0 {
		return nil, nil
	}
	parentTeam, err := r.db.Teams().GetTeamByID(ctx, r.team.ParentTeamID)
	if err != nil {
		return nil, err
	}
	return &TeamResolver{team: parentTeam, db: r.db}, nil
}

func (r *TeamResolver) ViewerCanAdminister(ctx context.Context) bool {
	// 🚨 SECURITY: For now administration is only allowed for site admins.
	err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	return err == nil
}

func (r *TeamResolver) Members(ctx context.Context, args *ListTeamMembersArgs) (*teamMemberConnection, error) {
	c := &teamMemberConnection{
		db:     r.db,
		teamID: r.team.ID,
	}
	if err := c.applyArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TeamResolver) ChildTeams(ctx context.Context, args *ListTeamsArgs) (*teamConnectionResolver, error) {
	c := &teamConnectionResolver{
		db:       r.db,
		parentID: r.team.ID,
	}
	if err := c.applyArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TeamResolver) OwnerField() string {
	return EnterpriseResolvers.ownResolver.TeamOwnerField(r)
}

type ListTeamMembersArgs struct {
	First  *int32
	After  *string
	Search *string
}

type teamMemberConnection struct {
	db       database.DB
	teamID   int32
	cursor   teamMemberListCursor
	search   string
	limit    int
	once     sync.Once
	nodes    []*types.TeamMember
	pageInfo *graphqlutil.PageInfo
	err      error
}

type teamMemberListCursor struct {
	TeamID int32 `json:"team,omitempty"`
	UserID int32 `json:"user,omitempty"`
}

// applyArgs unmarshals query conditions and limites set in `ListTeamMembersArgs`
// into `teamMemberConnection` fields for convenient use in database query.
func (r *teamMemberConnection) applyArgs(args *ListTeamMembersArgs) error {
	if args.After != nil && *args.After != "" {
		cursorText, err := graphqlutil.DecodeCursor(args.After)
		if err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(cursorText), &r.cursor); err != nil {
			return err
		}
	}
	if args.Search != nil {
		r.search = *args.Search
	}
	if args.First != nil {
		r.limit = int(*args.First)
	}
	return nil
}

// compute resolves team members queried for this resolver.
// The result of running it is setting `nodes`, `pageInfo` and `err`
// fields on the resolver. This ensures that resolving multiple
// graphQL attributes that require listing (like `pageInfo` and `nodes`)
// results in just one query.
func (r *teamMemberConnection) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := database.ListTeamMembersOpts{
			Cursor: database.TeamMemberListCursor{
				TeamID: r.cursor.TeamID,
				UserID: r.cursor.UserID,
			},
			TeamID: r.teamID,
			Search: r.search,
		}
		if r.limit != 0 {
			opts.LimitOffset = &database.LimitOffset{Limit: r.limit}
		}
		nodes, next, err := r.db.Teams().ListTeamMembers(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.nodes = nodes
		if next != nil {
			cursorStruct := teamMemberListCursor{
				TeamID: next.TeamID,
				UserID: next.UserID,
			}
			cursorBytes, err := json.Marshal(&cursorStruct)
			if err != nil {
				r.err = errors.Wrap(err, "error encoding pageInfo")
			}
			cursorString := string(cursorBytes)
			r.pageInfo = graphqlutil.EncodeCursor(&cursorString)
		} else {
			r.pageInfo = graphqlutil.HasNextPage(false)
		}
	})
}

func (r *teamMemberConnection) TotalCount(ctx context.Context, args *struct{ CountDeeplyNestedTeamMembers bool }) (int32, error) {
	if args != nil && args.CountDeeplyNestedTeamMembers {
		return 0, errors.New("Not supported: counting deeply nested team members.")
	}
	// Not taking into account limit or cursor for count.
	opts := database.ListTeamMembersOpts{
		TeamID: r.teamID,
		Search: r.search,
	}
	return r.db.Teams().CountTeamMembers(ctx, opts)
}

func (r *teamMemberConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	return r.pageInfo, nil
}

func (r *teamMemberConnection) Nodes(ctx context.Context) ([]*UserResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	var rs []*UserResolver
	// 🚨 Query in a loop is inefficient: Follow up with another pull request
	// to where team members query joins with users and fetches them in one go.
	for _, n := range r.nodes {
		if n.UserID == 0 {
			// 🚨 At this point only User can be a team member, so user ID should
			// always be present. If not, return a `null` team member.
			rs = append(rs, nil)
			continue
		}
		user, err := r.db.Users().GetByID(ctx, n.UserID)
		if err != nil {
			return nil, err
		}
		rs = append(rs, NewUserResolver(r.db, user))
	}
	return rs, nil
}

type CreateTeamArgs struct {
	Name           string
	DisplayName    *string
	ReadOnly       bool
	ParentTeam     *graphql.ID
	ParentTeamName *string
}

func (r *schemaResolver) CreateTeam(ctx context.Context, args *CreateTeamArgs) (*TeamResolver, error) {
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
	return &TeamResolver{team: &t, db: r.db}, nil
}

type UpdateTeamArgs struct {
	ID             *graphql.ID
	Name           *string
	DisplayName    *string
	ParentTeam     *graphql.ID
	ParentTeamName *string
}

func (r *schemaResolver) UpdateTeam(ctx context.Context, args *UpdateTeamArgs) (*TeamResolver, error) {
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
	return &TeamResolver{team: t, db: r.db}, nil
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

func (a *TeamMembersArgs) membersIDs() (map[int32]bool, error) {
	ids := map[int32]bool{}
	for i, memberID := range a.Members {
		if got, want := relay.UnmarshalKind(memberID), "TeamMember"; got != want {
			return nil, errors.Newf("Members[%d]=%q unexpected kind, got %q want %q", i, memberID, got, want)
		}
		var id int32
		if err := relay.UnmarshalSpec(memberID, &id); err != nil {
			return nil, errors.Wrapf(err, "Members[%d]=%q ID malformed", i, memberID)
		}
		ids[id] = true
	}
	return ids, nil
}

func (r *schemaResolver) AddTeamMembers(ctx context.Context, args *TeamMembersArgs) (*TeamResolver, error) {
	// 🚨 SECURITY: For now we only allow site admins to use teams.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("only site admins can modify team members")
	}
	if args.Team == nil && args.TeamName == nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), none specified")
	}
	if args.Team != nil && args.TeamName != nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), both specified")
	}
	memberIDs, err := args.membersIDs()
	if err != nil {
		return nil, err
	}
	team, err := findTeam(ctx, r.db.Teams(), args.Team, args.TeamName)
	if err != nil {
		return nil, err
	}
	listOpts := database.ListTeamMembersOpts{
		TeamID: team.ID,
	}
	for {
		existingMembers, cursor, err := r.db.Teams().ListTeamMembers(ctx, listOpts)
		if err != nil {
			return nil, err
		}
		for _, m := range existingMembers {
			delete(memberIDs, m.UserID)
		}
		if cursor == nil {
			break
		}
		listOpts.Cursor = *cursor
	}
	var membersToAdd []*types.TeamMember
	for userID := range memberIDs {
		membersToAdd = append(membersToAdd, &types.TeamMember{
			TeamID: team.ID,
			UserID: userID,
		})
	}
	if len(membersToAdd) > 0 {
		if err := r.db.Teams().CreateTeamMember(ctx, membersToAdd...); err != nil {
			return nil, err
		}
	}
	return &TeamResolver{
		db:   r.db,
		team: team,
	}, nil
}

func (r *schemaResolver) SetTeamMembers(args *TeamMembersArgs) *TeamResolver {
	return &TeamResolver{}
}

func (r *schemaResolver) RemoveTeamMembers(args *TeamMembersArgs) *TeamResolver {
	return &TeamResolver{}
}

func (r *schemaResolver) Teams(ctx context.Context, args *ListTeamsArgs) (*teamConnectionResolver, error) {
	// 🚨 SECURITY: For now we only allow site admins to use teams.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("only site admins can view teams")
	}
	c := &teamConnectionResolver{db: r.db}
	if err := c.applyArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

type TeamArgs struct {
	Name string
}

func (r *schemaResolver) Team(ctx context.Context, args *TeamArgs) (*TeamResolver, error) {
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

	return &TeamResolver{db: r.db, team: t}, nil
}
