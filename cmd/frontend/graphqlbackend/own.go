package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
	Nodes(context.Context) ([]OwnershipResolver, error)
}

type Ownable interface {
	ToGitBlob(context.Context) (*GitTreeEntryResolver, bool)
}

type OwnershipResolver interface {
	Owner(context.Context) (OwnerResolver, error)
	Reasons(context.Context) ([]OwnershipReasonResolver, error)
}

type OwnerResolver interface {
	OwnerField(context.Context) (string, error)

	ToPerson() (*PersonResolver, bool)
	ToTeam() (*TeamResolver, bool)
}

type OwnershipReasonResolver interface {
	ToCodeownersFileEntry() (CodeownersFileEntryResolver, bool)
}

type CodeownersFileEntryResolver interface {
	Title(context.Context) (string, error)
	Description(context.Context) (string, error)
	CodeownersFile(context.Context) (FileResolver, error)
	RuleLineMatch(context.Context) (int32, error)
}
