package graphqlbackend

import (
	"context"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/orgs"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type organizationResolver struct {
	organization *sourcegraph.Org
}

func (o *organizationResolver) Login() string {
	return o.organization.Login
}

func (o *organizationResolver) ID() int32 {
	return int32(o.organization.ID)
}

func (o *organizationResolver) Email() string {
	return o.organization.Email
}

func (o *organizationResolver) Name() string {
	return o.organization.Name
}

func (o *organizationResolver) AvatarURL() string {
	return o.organization.AvatarURL
}

func (o *organizationResolver) Description() string {
	return o.organization.Description
}

func (o *organizationResolver) Collaborators() int32 {
	return o.organization.Collaborators
}

type organizationMemberResolver struct {
	member *sourcegraph.OrgMember
}

func (o *organizationResolver) Members(ctx context.Context) ([]*organizationMemberResolver, error) {
	opts := &sourcegraph.OrgListOptions{
		OrgID:   strconv.Itoa(int(o.organization.ID)),
		OrgName: o.organization.Login,
	}

	// TODO(Dan): this method currently only returns a single page of results
	membersList, err := orgs.ListOrgMembersForInvites(ctx, opts)
	if err != nil {
		return nil, err
	}

	res := make([]*organizationMemberResolver, len(membersList.OrgMembers))
	for i, member := range membersList.OrgMembers {
		res[i] = &organizationMemberResolver{member}
	}
	return res, nil
}

func (m *organizationMemberResolver) Login() string {
	return m.member.Login
}

func (m *organizationMemberResolver) ID() int32 {
	return int32(m.member.ID)
}

func (m *organizationMemberResolver) Email() string {
	return m.member.Email
}

func (m *organizationMemberResolver) AvatarURL() string {
	return m.member.AvatarURL
}

func (m *organizationMemberResolver) IsSourcegraphUser() bool {
	return m.member.SourcegraphUser
}

func (m *organizationMemberResolver) CanInvite() bool {
	return m.member.CanInvite
}

type inviteResolver struct {
	invite *sourcegraph.UserInvite
}

func (m *organizationMemberResolver) Invite() *inviteResolver {
	if m.member.Invite == nil {
		return nil
	}
	return &inviteResolver{invite: m.member.Invite}
}

func (i *inviteResolver) UserLogin() string {
	return i.invite.UserID
}

func (i *inviteResolver) UserEmail() string {
	return i.invite.UserEmail
}

func (i *inviteResolver) OrgLogin() string {
	return i.invite.OrgName
}

func (i *inviteResolver) OrgID() (int32, error) {
	v, err := strconv.Atoi(i.invite.OrgID)
	if err != nil {
		return int32(v), nil
	}
	return 0, err
}

func (i *inviteResolver) SentAt() int32 {
	return int32(i.invite.SentAt.Unix())
}

func (i *inviteResolver) URI() string {
	return i.invite.URI
}

func (*schemaResolver) InviteOrgMemberToSourcegraph(ctx context.Context, args *struct {
	OrgLogin  string
	OrgID     int32
	UserLogin string
	UserEmail string
}) (bool, error) {
	res, err := orgs.InviteUser(ctx, &sourcegraph.UserInvite{
		OrgName:   args.OrgLogin,
		OrgID:     strconv.Itoa(int(args.OrgID)),
		UserID:    args.UserLogin,
		UserEmail: args.UserEmail,
	})
	if err != nil {
		return false, err
	}
	if res == sourcegraph.InviteMissingEmail {
		return false, nil
	}
	return true, nil
}
