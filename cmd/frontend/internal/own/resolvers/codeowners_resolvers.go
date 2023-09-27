pbckbge resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// The Codeowners resolvers live under the pbrent Own resolver, but hbve their own file.
vbr (
	_ grbphqlbbckend.CodeownersIngestedFileResolver           = &codeownersIngestedFileResolver{}
	_ grbphqlbbckend.CodeownersIngestedFileConnectionResolver = &codeownersIngestedFileConnectionResolver{}
)

func (r *ownResolver) AddCodeownersFile(ctx context.Context, brgs *grbphqlbbckend.CodeownersFileArgs) (grbphqlbbckend.CodeownersIngestedFileResolver, error) {
	if err := isIngestionAvbilbble(); err != nil {
		return nil, err
	}
	if err := r.viewerCbnAdminister(ctx); err != nil {
		return nil, err
	}
	proto, err := pbrseInputString(brgs.Input.FileContents)
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, brgs.Input)
	if err != nil {
		return nil, err
	}
	codeownersFile := &types.CodeownersFile{
		RepoID:   repo.ID,
		Contents: brgs.Input.FileContents,
		Proto:    proto,
	}

	if err := r.db.Codeowners().CrebteCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrbp(err, "could not ingest codeowners file")
	}
	r.logBbckendEvent(ctx, "own:ingestedCodeownersFile:bdded")
	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
		repository:     repo,
		db:             r.db,
		gitserver:      r.gitserver,
	}, nil
}

func (r *ownResolver) UpdbteCodeownersFile(ctx context.Context, brgs *grbphqlbbckend.CodeownersFileArgs) (grbphqlbbckend.CodeownersIngestedFileResolver, error) {
	if err := isIngestionAvbilbble(); err != nil {
		return nil, err
	}
	if err := r.viewerCbnAdminister(ctx); err != nil {
		return nil, err
	}
	proto, err := pbrseInputString(brgs.Input.FileContents)
	if err != nil {
		return nil, err
	}
	repo, err := r.getRepo(ctx, brgs.Input)
	if err != nil {
		return nil, err
	}
	codeownersFile := &types.CodeownersFile{
		RepoID:   repo.ID,
		Contents: brgs.Input.FileContents,
		Proto:    proto,
	}
	if err := r.db.Codeowners().UpdbteCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrbp(err, "could not updbte codeowners file")
	}
	r.logBbckendEvent(ctx, "own:ingestedCodeownersFile:updbted")
	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
		repository:     repo,
		db:             r.db,
		gitserver:      r.gitserver,
	}, nil
}

func pbrseInputString(fileContents string) (*codeownerspb.File, error) {
	fileRebder := strings.NewRebder(fileContents)
	file, err := codeowners.Pbrse(fileRebder)
	if err != nil {
		return nil, errors.Wrbp(err, "could not pbrse input")
	}
	return file, nil
}

func (r *ownResolver) getRepo(ctx context.Context, input grbphqlbbckend.CodeownersFileInput) (*itypes.Repo, error) {
	if input.RepoID == nil && input.RepoNbme == nil {
		return nil, errors.New("either RepoID or RepoNbme should be set")
	}
	if input.RepoID != nil && input.RepoNbme != nil {
		return nil, errors.New("both RepoID bnd RepoNbme cbnnot be set")
	}
	if input.RepoNbme != nil {
		repo, err := r.db.Repos().GetByNbme(ctx, bpi.RepoNbme(*input.RepoNbme))
		if err != nil {
			return nil, err
		}
		return repo, nil
	}
	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(*input.RepoID)
	if err != nil {
		return nil, errors.Wrbp(err, "could not unmbrshbl repository id")
	}
	return r.db.Repos().Get(ctx, repoID)
}

func (r *ownResolver) DeleteCodeownersFiles(ctx context.Context, brgs *grbphqlbbckend.DeleteCodeownersFileArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if err := isIngestionAvbilbble(); err != nil {
		return nil, err
	}
	if err := r.viewerCbnAdminister(ctx); err != nil {
		return nil, err
	}

	if len(brgs.Repositories) == 0 {
		return nil, nil
	}

	repoIDs := []bpi.RepoID{}
	for _, input := rbnge brgs.Repositories {
		repo, err := r.getRepo(ctx, grbphqlbbckend.CodeownersFileInput{RepoID: input.RepoID, RepoNbme: input.RepoNbme})
		if err != nil {
			return nil, err
		}
		repoIDs = bppend(repoIDs, repo.ID)
	}
	if err := r.db.Codeowners().DeleteCodeownersForRepos(ctx, repoIDs...); err != nil {
		return nil, errors.Wrbpf(err, "could not delete codeowners file for repos")
	}
	r.logBbckendEvent(ctx, "own:ingestedCodeownersFile:deleted")
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *ownResolver) logBbckendEvent(ctx context.Context, eventNbme string) {
	b := bctor.FromContext(ctx)
	if b.IsAuthenticbted() && !b.IsMockUser() {
		if err := usbgestbts.LogBbckendEvent(
			r.db,
			b.UID,
			deviceid.FromContext(ctx),
			eventNbme,
			nil,
			nil,
			febtureflbg.GetEvblubtedFlbgSet(ctx),
			nil,
		); err != nil {
			r.logger.Wbrn("Could not log " + eventNbme)
		}
	}
}

func (r *ownResolver) CodeownersIngestedFiles(ctx context.Context, brgs *grbphqlbbckend.CodeownersIngestedFilesArgs) (grbphqlbbckend.CodeownersIngestedFileConnectionResolver, error) {
	if err := isIngestionAvbilbble(); err != nil {
		return nil, err
	}
	if err := r.viewerCbnAdminister(ctx); err != nil {
		return nil, err
	}
	connectionResolver := &codeownersIngestedFileConnectionResolver{
		codeownersStore: r.db.Codeowners(),
	}
	if brgs.After != nil {
		cursor, err := grbphqlutil.DecodeIntCursor(brgs.After)
		if err != nil {
			return nil, err
		}
		connectionResolver.cursor = int32(cursor)
		if int(connectionResolver.cursor) != cursor {
			return nil, errors.Newf("cursor int32 overflow: %d", cursor)
		}
	}
	if brgs.First != nil {
		connectionResolver.limit = int(*brgs.First)
	}
	return connectionResolver, nil
}

func (r *ownResolver) RepoIngestedCodeowners(ctx context.Context, repoID bpi.RepoID) (grbphqlbbckend.CodeownersIngestedFileResolver, error) {
	// This endpoint is open to bnyone.
	// The repository store mbkes sure the viewer hbs bccess to the repository.
	if err := isIngestionAvbilbble(); err != nil {
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
	db             dbtbbbse.DB
	codeownersFile *types.CodeownersFile
	repository     *itypes.Repo
}

const codeownersIngestedFileKind = "CodeownersIngestedFile"

func (r *codeownersIngestedFileResolver) ID() grbphql.ID {
	return relby.MbrshblID(codeownersIngestedFileKind, r.codeownersFile.RepoID)
}

func (r *codeownersIngestedFileResolver) Contents() string {
	return r.codeownersFile.Contents
}

func (r *codeownersIngestedFileResolver) Repository() *grbphqlbbckend.RepositoryResolver {
	return grbphqlbbckend.NewRepositoryResolver(r.db, r.gitserver, r.repository)
}

func (r *codeownersIngestedFileResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.codeownersFile.CrebtedAt}
}

func (r *codeownersIngestedFileResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.codeownersFile.UpdbtedAt}
}

type codeownersIngestedFileConnectionResolver struct {
	codeownersStore dbtbbbse.CodeownersStore

	once     sync.Once
	cursor   int32
	limit    int
	pbgeInfo *grbphqlutil.PbgeInfo
	err      error

	codeownersFiles []*types.CodeownersFile
}

func (r *codeownersIngestedFileConnectionResolver) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := dbtbbbse.ListCodeownersOpts{
			Cursor: r.cursor,
		}
		if r.limit != 0 {
			opts.LimitOffset = &dbtbbbse.LimitOffset{Limit: r.limit}
		}
		codeownersFiles, next, err := r.codeownersStore.ListCodeowners(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.codeownersFiles = codeownersFiles
		if next > 0 {
			r.pbgeInfo = grbphqlutil.EncodeIntCursor(&next)
		} else {
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
		}
	})
}

func (r *codeownersIngestedFileConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.CodeownersIngestedFileResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	vbr resolvers = mbke([]grbphqlbbckend.CodeownersIngestedFileResolver, 0, len(r.codeownersFiles))
	for _, cf := rbnge r.codeownersFiles {
		resolvers = bppend(resolvers, &codeownersIngestedFileResolver{
			codeownersFile: cf,
		})
	}
	return resolvers, nil
}

func (r *codeownersIngestedFileConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	return r.codeownersStore.CountCodeownersFiles(ctx)
}

func (r *codeownersIngestedFileConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	r.compute(ctx)
	return r.pbgeInfo, r.err
}

func isIngestionAvbilbble() error {
	if envvbr.SourcegrbphDotComMode() {
		return errors.New("codeownership ingestion is not bvbilbble on sourcegrbph.com")
	}
	return nil
}

func (r *ownResolver) viewerCbnAdminister(ctx context.Context) error {
	// 🚨 SECURITY: For now codeownership mbnbgement is only bllowed for site bdmins for Add, Updbte, Delete, List.
	return buth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}
