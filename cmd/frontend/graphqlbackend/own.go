package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type ListOwnershipArgs struct {
	First   *int32
	After   *string
	Reasons *[]string
}

type OwnResolver interface {
	GitBlobOwnership(ctx context.Context, blob *GitTreeEntryResolver, args ListOwnershipArgs) (OwnershipConnectionResolver, error)

	PersonOwnerField(person *PersonResolver) string
	UserOwnerField(user *UserResolver) string
	TeamOwnerField(team *TeamResolver) string

	NodeResolvers() map[string]NodeByIDFunc
}

type OwnershipConnectionResolver interface {
	TotalCount() int32
	PageInfo() *graphqlutil.PageInfo
	Nodes() []OwnershipResolver
}

type Ownable interface {
	ToGitBlob() (*GitTreeEntryResolver, bool)
}

type OwnershipResolver interface {
	Owner() OwnerResolver
	Reasons() []OwnershipReasonResolver
}

type OwnerResolver interface {
	OwnerField() string

	ToPerson() (*PersonResolver, bool)
	ToUser() (*UserResolver, bool)
	ToTeam() (*TeamResolver, bool)
}

type OwnershipReasonResolver interface {
	ToCodeownersFileEntry() (CodeownersFileEntryResolver, bool)
	ToRecentContributor() (RecentContributorResolver, bool)
}

type CodeownersFileEntryResolver interface {
	Title() string
	Description() string
	Ingested() bool
	CodeownersFile() FileResolver
	RuleLineMatch() int32
}

type RecentContributorResolver interface {
	Title() string
	Description() string
	Since() gqlutil.DateTime
	Count() int32
}
