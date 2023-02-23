package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	_ graphqlbackend.CodeownersIngestedFileResolver           = &codeownersIngestedFileResolver{}
	_ graphqlbackend.IngestedCodeownersResolver               = &codeownersResolver{}
	_ graphqlbackend.CodeownersIngestedFileConnectionResolver = &codeownersIngestedFileConnectionResolver{}
)

type codeownersResolver struct{}

func (c *codeownersResolver) AddCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	return &codeownersIngestedFileResolver{}, nil
}

func (c *codeownersResolver) UpdateCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	return &codeownersIngestedFileResolver{}, nil
}

func (c *codeownersResolver) DeleteCodeownersFile(ctx context.Context, args *graphqlbackend.DeleteCodeownersFileArgs) error {
	return nil
}

func (c *codeownersResolver) CodeownersIngestedFiles(ctx context.Context, args *graphqlbackend.CodeownersIngestedFilesArgs) ([]graphqlbackend.CodeownersIngestedFileConnectionResolver, error) {
	return nil, nil
}

type codeownersIngestedFileResolver struct {
	codeownersFile types.CodeownersFile
}

func (c *codeownersIngestedFileResolver) Contents() string {
	return c.codeownersFile.Contents
}

func (c *codeownersIngestedFileResolver) RepoID() int32 {
	return int32(c.codeownersFile.RepoID)
}

func (c *codeownersIngestedFileResolver) CreatedAt() time.Time {
	return c.codeownersFile.CreatedAt
}

func (c *codeownersIngestedFileResolver) UpdatedAt() time.Time {
	return c.codeownersFile.UpdatedAt
}

type codeownersIngestedFileConnectionResolver struct{}

func (c *codeownersIngestedFileConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CodeownersIngestedFileResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (c *codeownersIngestedFileConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	//TODO implement me
	panic("implement me")
}

func (c *codeownersIngestedFileConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}
