package graphqlbackend

import (
	"context"
	"time"

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
	ToUser() (*UserResolver, bool)
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

type CodeownersResolver interface {
	AddCodeownersFile(context.Context, *CodeownersFileArgs) (CodeownersIngestedFileResolver, error)
	UpdateCodeownersFile(context.Context, *CodeownersFileArgs) (CodeownersIngestedFileResolver, error)
	DeleteCodeownersFile(context.Context, *DeleteCodeownersFileArgs) error

	CodeownersIngestedFiles(context.Context, *CodeownersIngestedFilesArgs) ([]CodeownersIngestedFileConnectionResolver, error)
}

type CodeownersFileArgs struct {
	FileContents string
	RepoID       int32
}

type DeleteCodeownersFileArgs struct {
	RepoID int32
}

type CodeownersIngestedFilesArgs struct {
	First  *int32
	After  *string
	RepoID *int32
}

type CodeownersIngestedFileResolver interface {
	Contents() string
	RepoID() int32
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type CodeownersIngestedFileConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeownersIngestedFileResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}
