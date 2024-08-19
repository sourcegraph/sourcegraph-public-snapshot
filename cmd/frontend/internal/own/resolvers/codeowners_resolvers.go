package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/own/types"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// The Codeowners resolvers live under the parent Own resolver, but have their own file.
var (
	_ graphqlbackend.CodeownersIngestedFileResolver           = &codeownersIngestedFileResolver{}
	_ graphqlbackend.CodeownersIngestedFileConnectionResolver = &codeownersIngestedFileConnectionResolver{}
)

func (r *ownResolver) AddCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	if err := isIngestionAvailable(); err != nil {
		return nil, err
	}
	if err := r.viewerCanAdminister(ctx); err != nil {
		return nil, err
	}
	proto, err := parseInputString(args.Input.FileContents)
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, args.Input)
	if err != nil {
		return nil, err
	}
	codeownersFile := &types.CodeownersFile{
		RepoID:   repo.ID,
		Contents: args.Input.FileContents,
		Proto:    proto,
	}

	if err := r.db.Codeowners().CreateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not ingest codeowners file")
	}
	r.logBackendEvent(ctx, "own:ingestedCodeownersFile:added")
	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
		repository:     repo,
		db:             r.db,
		gitserver:      r.gitserver,
	}, nil
}

func (r *ownResolver) UpdateCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	if err := isIngestionAvailable(); err != nil {
		return nil, err
	}
	if err := r.viewerCanAdminister(ctx); err != nil {
		return nil, err
	}
	proto, err := parseInputString(args.Input.FileContents)
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, args.Input)
	if err != nil {
		return nil, err
	}
	codeownersFile := &types.CodeownersFile{
		RepoID:   repo.ID,
		Contents: args.Input.FileContents,
		Proto:    proto,
	}
	if err := r.db.Codeowners().UpdateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not update codeowners file")
	}
	r.logBackendEvent(ctx, "own:ingestedCodeownersFile:updated")
	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
		repository:     repo,
		db:             r.db,
		gitserver:      r.gitserver,
	}, nil
}

func parseInputString(fileContents string) (*codeownerspb.File, error) {
	fileReader := strings.NewReader(fileContents)
	file, err := codeowners.Parse(fileReader)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse input")
	}
	return file, nil
}

func (r *ownResolver) getRepo(ctx context.Context, input graphqlbackend.CodeownersFileInput) (*itypes.Repo, error) {
	if input.RepoID == nil && input.RepoName == nil {
		return nil, errors.New("either RepoID or RepoName should be set")
	}
	if input.RepoID != nil && input.RepoName != nil {
		return nil, errors.New("both RepoID and RepoName cannot be set")
	}
	if input.RepoName != nil {
		repo, err := r.db.Repos().GetByName(ctx, api.RepoName(*input.RepoName))
		if err != nil {
			return nil, err
		}
		return repo, nil
	}
	repoID, err := graphqlbackend.UnmarshalRepositoryID(*input.RepoID)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal repository id")
	}
	return r.db.Repos().Get(ctx, repoID)
}

func (r *ownResolver) DeleteCodeownersFiles(ctx context.Context, args *graphqlbackend.DeleteCodeownersFileArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := isIngestionAvailable(); err != nil {
		return nil, err
	}
	if err := r.viewerCanAdminister(ctx); err != nil {
		return nil, err
	}

	if len(args.Repositories) == 0 {
		return nil, nil
	}

	repoIDs := []api.RepoID{}
	for _, input := range args.Repositories {
		repo, err := r.getRepo(ctx, graphqlbackend.CodeownersFileInput{RepoID: input.RepoID, RepoName: input.RepoName})
		if err != nil {
			return nil, err
		}
		repoIDs = append(repoIDs, repo.ID)
	}
	if err := r.db.Codeowners().DeleteCodeownersForRepos(ctx, repoIDs...); err != nil {
		return nil, errors.Wrapf(err, "could not delete codeowners file for repos")
	}
	r.logBackendEvent(ctx, "own:ingestedCodeownersFile:deleted")
	return &graphqlbackend.EmptyResponse{}, nil
}

// TODO: Use EventRecorder from internal/telemetryrecorder instead.
func (r *ownResolver) logBackendEvent(ctx context.Context, eventName string) {
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() && !a.IsMockUser() {
		//lint:ignore SA1019 existing usage of deprecated functionality.
		if err := usagestats.LogBackendEvent(
			r.db,
			a.UID,
			deviceid.FromContext(ctx),
			eventName,
			nil,
			nil,
			featureflag.GetEvaluatedFlagSet(ctx),
			nil,
		); err != nil {
			r.logger.Warn("Could not log " + eventName)
		}
	}
}

func (r *ownResolver) CodeownersIngestedFiles(ctx context.Context, args *graphqlbackend.CodeownersIngestedFilesArgs) (graphqlbackend.CodeownersIngestedFileConnectionResolver, error) {
	if err := isIngestionAvailable(); err != nil {
		return nil, err
	}
	if err := r.viewerCanAdminister(ctx); err != nil {
		return nil, err
	}
	connectionResolver := &codeownersIngestedFileConnectionResolver{
		codeownersStore: r.db.Codeowners(),
	}
	if args.After != nil {
		cursor, err := gqlutil.DecodeIntCursor(args.After)
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

func (r *ownResolver) RepoIngestedCodeowners(ctx context.Context, repoID api.RepoID) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	// This endpoint is open to anyone.
	// The repository store makes sure the viewer has access to the repository.
	if err := isIngestionAvailable(); err != nil {
		return nil, err
	}
	repo, err := r.db.Repos().Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	codeownersFile, err := r.db.Codeowners().GetCodeownersForRepo(ctx, repoID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
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
	db             database.DB
	codeownersFile *types.CodeownersFile
	repository     *itypes.Repo
}

const codeownersIngestedFileKind = "CodeownersIngestedFile"

func (r *codeownersIngestedFileResolver) ID() graphql.ID {
	return relay.MarshalID(codeownersIngestedFileKind, r.codeownersFile.RepoID)
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
	codeownersStore database.CodeownersStore

	once     sync.Once
	cursor   int32
	limit    int
	pageInfo *gqlutil.PageInfo
	err      error

	codeownersFiles []*types.CodeownersFile
}

func (r *codeownersIngestedFileConnectionResolver) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := database.ListCodeownersOpts{
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
			r.pageInfo = gqlutil.EncodeIntCursor(&next)
		} else {
			r.pageInfo = gqlutil.HasNextPage(false)
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

func (r *codeownersIngestedFileConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	r.compute(ctx)
	return r.pageInfo, r.err
}

func isIngestionAvailable() error {
	if dotcom.SourcegraphDotComMode() {
		return errors.New("codeownership ingestion is not available on sourcegraph.com")
	}
	return nil
}

func (r *ownResolver) viewerCanAdminister(ctx context.Context) error {
	// 🚨 SECURITY: For now codeownership management is only allowed for site admins for Add, Update, Delete, List.
	return auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}
