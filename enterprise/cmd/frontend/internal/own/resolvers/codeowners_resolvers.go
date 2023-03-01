package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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

func (r *ownResolver) ViewerCanAdminister(ctx context.Context) error {
	// 🚨 SECURITY: For now codeownership management is only allowed for site admins for Add, Update, Delete, List.
	// Eventually we should allow users to access the Get method, but check that they have view permissions on the repository.
	return auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}

func (r *ownResolver) AddCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	if err := r.ViewerCanAdminister(ctx); err != nil {
		return nil, err
	}
	proto, err := parseInputString(args.Input.FileContents)
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, args.Input.RepositoryID, args.Input.RepositoryName)
	if err != nil {
		return nil, err
	}
	codeownersFile := &types.CodeownersFile{
		RepoID:   repo.ID,
		Contents: args.Input.FileContents,
		Proto:    proto,
	}

	if err := r.codeownersStore.CreateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not ingest codeowners file")
	}

	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
		repository:     repo,
		db:             r.db,
		gitserver:      r.gitserver,
	}, nil
}

func (r *ownResolver) UpdateCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	if err := r.ViewerCanAdminister(ctx); err != nil {
		return nil, err
	}
	proto, err := parseInputString(args.Input.FileContents)
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, args.Input.RepositoryID, args.Input.RepositoryName)
	if err != nil {
		return nil, err
	}
	codeownersFile := &types.CodeownersFile{
		RepoID:   repo.ID,
		Contents: args.Input.FileContents,
		Proto:    proto,
	}
	if err := r.codeownersStore.UpdateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not update codeowners file")
	}

	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
		repository:     repo,
		db:             r.db,
		gitserver:      r.gitserver,
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

func (r *ownResolver) getRepo(ctx context.Context, repositoryID *int32, repositoryName *string) (*types.Repo, error) {
	if repositoryID == nil && repositoryName == nil {
		return nil, errors.New("either RepositoryID or RepositoryName should be set")
	}
	if repositoryID != nil && repositoryName != nil {
		return nil, errors.New("both RepositoryID and RepositoryName cannot be set")
	}
	if repositoryName != nil {
		repo, err := r.db.Repos().GetByName(ctx, api.RepoName(*repositoryName))
		if err != nil {
			return nil, errors.Wrapf(err, "could not fetch repository for name %v", repositoryName)
		}
		return repo, nil
	}
	repo, err := r.db.Repos().GetByIDs(ctx, api.RepoID(*repositoryID))
	if err != nil {
		return nil, errors.Wrapf(err, "could not fetch repository for ID %v", repositoryID)
	}
	if len(repo) != 1 {
		return nil, errors.New("could not fetch repository")
	}
	return repo[0], nil
}

func (r *ownResolver) DeleteCodeownersFiles(ctx context.Context, args *graphqlbackend.DeleteCodeownersFileArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.ViewerCanAdminister(ctx); err != nil {
		return nil, err
	}
	if err := r.codeownersStore.DeleteCodeownersForRepos(ctx, args.RepositoryIDs...); err != nil {
		return nil, errors.Wrapf(err, "could not delete codeowners file for repos +%d", args.RepositoryIDs)
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *ownResolver) CodeownersIngestedFiles(ctx context.Context, args *graphqlbackend.CodeownersIngestedFilesArgs) (graphqlbackend.CodeownersIngestedFileConnectionResolver, error) {
	if err := r.ViewerCanAdminister(ctx); err != nil {
		return nil, err
	}
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

func (r *ownResolver) CodeownersIngestedFile(ctx context.Context, args *graphqlbackend.CodeownersIngestedFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	// TODO: do we need to check repo permissions here?
	codeownersFile, err := r.codeownersStore.GetCodeownersForRepo(ctx, api.RepoID(args.RepositoryID))
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, &args.RepositoryID, nil)
	if err != nil {
		return nil, err
	}
	return &codeownersIngestedFileResolver{
		gitserver:      r.gitserver,
		db:             r.db,
		codeownersFile: codeownersFile,
		repository:     repo,
	}, nil
}

type codeownersIngestedFileResolver struct {
	gitserver      gitserver.Client
	db             edb.EnterpriseDB
	codeownersFile *types.CodeownersFile
	repository     *types.Repo
}

func (r *codeownersIngestedFileResolver) Contents() string {
	return r.codeownersFile.Contents
}

func (r *codeownersIngestedFileResolver) Repository() *graphqlbackend.RepositoryResolver {
	return graphqlbackend.NewRepositoryResolver(r.db, r.gitserver, r.repository)
}

func (r *codeownersIngestedFileResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.codeownersFile.CreatedAt}
}

func (r *codeownersIngestedFileResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.codeownersFile.UpdatedAt}
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
