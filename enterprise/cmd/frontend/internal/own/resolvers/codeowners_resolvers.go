package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// The Codeowners resolvers live under the parent Own resolver, but have their own file.
var (
	_ graphqlbackend.CodeownersIngestedFileResolver           = &codeownersIngestedFileResolver{}
	_ graphqlbackend.CodeownersIngestedFileConnectionResolver = &codeownersIngestedFileConnectionResolver{}
)

func (r *ownResolver) AddCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	proto, err := parseInputString(args.FileContents)
	if err != nil {
		return nil, err
	}

	codeownersFile := &types.CodeownersFile{
		RepoID:   api.RepoID(args.RepoID),
		Contents: args.FileContents,
		Proto:    proto,
	}
	if err := r.codeownersStore.CreateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not ingest codeowners file")
	}

	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
	}, nil
}

func (r *ownResolver) UpdateCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	proto, err := parseInputString(args.FileContents)
	if err != nil {
		return nil, err
	}

	codeownersFile := &types.CodeownersFile{
		RepoID:   api.RepoID(args.RepoID),
		Contents: args.FileContents,
		Proto:    proto,
	}
	if err := r.codeownersStore.UpdateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not update codeowners file")
	}

	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
	}, nil
}

func parseInputString(fileContents string) (*codeownerspb.File, error) {
	fileReader := strings.NewReader(fileContents)
	proto, err := codeowners.Parse(fileReader)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse input")
	}
	return proto, nil
}

func (r *ownResolver) DeleteCodeownersFile(ctx context.Context, args *graphqlbackend.DeleteCodeownersFileArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.codeownersStore.DeleteCodeownersForRepo(ctx, api.RepoID(args.RepoID)); err != nil {
		return nil, errors.Wrapf(err, "could not delete codeowners file for repo %d", args.RepoID)
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *ownResolver) CodeownersIngestedFiles(ctx context.Context, args *graphqlbackend.CodeownersIngestedFilesArgs) (graphqlbackend.CodeownersIngestedFileConnectionResolver, error) {
	connectionResolver := &codeownersIngestedFileConnectionResolver{
		codeownersStore: r.codeownersStore,
	}
	if args.After != nil {
		cursor, err := graphqlutil.DecodeIntCursor(args.After)
		if err != nil {
			return nil, err
		}
		connectionResolver.cursor = int32(cursor)
		if int(connectionResolver.cursor) != cursor {
			return nil, errors.Newf("cursor int32 overflow: %d", cursor)
		}
	}
	if args.First != nil {
		connectionResolver.limit = int(*args.First)
	}
	return connectionResolver, nil
}

type codeownersIngestedFileResolver struct {
	codeownersFile *types.CodeownersFile
}

func (c *codeownersIngestedFileResolver) Contents() string {
	return c.codeownersFile.Contents
}

func (c *codeownersIngestedFileResolver) RepoID() int32 {
	return int32(c.codeownersFile.RepoID)
}

func (c *codeownersIngestedFileResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: c.codeownersFile.CreatedAt}
}

func (c *codeownersIngestedFileResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: c.codeownersFile.UpdatedAt}
}

type codeownersIngestedFileConnectionResolver struct {
	codeownersStore edb.CodeownersStore

	once     sync.Once
	cursor   int32
	limit    int
	pageInfo *graphqlutil.PageInfo
	err      error

	codeownersFiles []*types.CodeownersFile
}

func (r *codeownersIngestedFileConnectionResolver) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := edb.ListCodeownersOpts{
			Cursor: r.cursor,
		}
		if r.limit != 0 {
			opts.LimitOffset = &database.LimitOffset{Limit: r.limit}
		}
		codeownersFiles, next, err := r.codeownersStore.ListCodeowners(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.codeownersFiles = codeownersFiles
		if next > 0 {
			r.pageInfo = graphqlutil.EncodeIntCursor(&next)
		} else {
			r.pageInfo = graphqlutil.HasNextPage(false)
		}
	})
}

func (r *codeownersIngestedFileConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CodeownersIngestedFileResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	var resolvers = make([]graphqlbackend.CodeownersIngestedFileResolver, 0, len(r.codeownersFiles))
	for _, cf := range r.codeownersFiles {
		resolvers = append(resolvers, &codeownersIngestedFileResolver{
			codeownersFile: cf,
		})
	}
	return resolvers, nil
}

func (r *codeownersIngestedFileConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.codeownersStore.CountCodeownersFiles(ctx)
}

func (r *codeownersIngestedFileConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	r.compute(ctx)
	return r.pageInfo, r.err
}
