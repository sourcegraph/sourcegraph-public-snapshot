package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
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

	// Codeowners queries
	CodeownersIngestedFiles(context.Context, *CodeownersIngestedFilesArgs) (CodeownersIngestedFileConnectionResolver, error)
	RepoIngestedCodeowners(context.Context, api.RepoID) (CodeownersIngestedFileResolver, error)

	// Codeowners mutations
	AddCodeownersFile(context.Context, *CodeownersFileArgs) (CodeownersIngestedFileResolver, error)
	UpdateCodeownersFile(context.Context, *CodeownersFileArgs) (CodeownersIngestedFileResolver, error)
	DeleteCodeownersFiles(context.Context, *DeleteCodeownersFileArgs) (*EmptyResponse, error)
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

type CodeownersFileArgs struct {
	Input CodeownersFileInput
}

type CodeownersFileInput struct {
	FileContents string
	RepoID       *graphql.ID
	RepoName     *string
}

type DeleteCodeownersFilesInput struct {
	RepoID   *graphql.ID
	RepoName *string
}

type DeleteCodeownersFileArgs struct {
	Repositories []DeleteCodeownersFilesInput
}

type CodeownersIngestedFilesArgs struct {
	First *int32
	After *string
}

type CodeownersIngestedFileResolver interface {
	ID() graphql.ID
	Contents() string
	Repository() *RepositoryResolver
	CreatedAt() gqlutil.DateTime
	UpdatedAt() gqlutil.DateTime
}

type CodeownersIngestedFileConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeownersIngestedFileResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}
