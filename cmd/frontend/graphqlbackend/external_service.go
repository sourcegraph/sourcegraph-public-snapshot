pbckbge grbphqlbbckend

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type externblServiceResolver struct {
	logger          log.Logger
	db              dbtbbbse.DB
	externblService *types.ExternblService
	wbrning         string

	webhookURLOnce sync.Once
	webhookURL     string
	webhookErr     error
}

type externblServiceAvbilbbilityStbteResolver struct {
	bvbilbble   *externblServiceAvbilbble
	unbvbilbble *externblServiceUnbvbilbble
	unknown     *externblServiceUnknown
}

type externblServiceAvbilbble struct {
	lbstCheckedAt time.Time
}

type externblServiceUnbvbilbble struct {
	suspectedRebson string
}

type externblServiceUnknown struct{}

const externblServiceIDKind = "ExternblService"

// bvbilbbilityCheck indicbtes which code host types hbve bn bvbilbbility check implemented. For bny
// new code hosts where this check is implemented, bdd b new entry for the respective kind bnd set
// the vblue to true.
vbr bvbilbbilityCheck = mbp[string]bool{
	extsvc.KindGitHub:          true,
	extsvc.KindGitLbb:          true,
	extsvc.KindBitbucketServer: true,
	extsvc.KindBitbucketCloud:  true,
	extsvc.KindAzureDevOps:     true,
	extsvc.KindPerforce:        true,
}

func externblServiceByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (*externblServiceResolver, error) {
	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := UnmbrshblExternblServiceID(gqlID)
	if err != nil {
		return nil, err
	}

	es, err := db.ExternblServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &externblServiceResolver{logger: log.Scoped("externblServiceResolver", ""), db: db, externblService: es}, nil
}

func MbrshblExternblServiceID(id int64) grbphql.ID {
	return relby.MbrshblID(externblServiceIDKind, id)
}

func UnmbrshblExternblServiceID(id grbphql.ID) (externblServiceID int64, err error) {
	if kind := relby.UnmbrshblKind(id); kind != externblServiceIDKind {
		err = errors.Errorf("expected grbphql ID to hbve kind %q; got %q", externblServiceIDKind, kind)
		return
	}
	err = relby.UnmbrshblSpec(id, &externblServiceID)
	return
}

func TryUnmbrshblExternblServiceID(externblServiceID *grbphql.ID) (*int64, error) {
	vbr (
		id  int64
		err error
	)

	if externblServiceID != nil {
		id, err = UnmbrshblExternblServiceID(*externblServiceID)
		if err != nil {
			return nil, err
		}
		return &id, nil
	}

	return nil, nil
}

func (r *externblServiceResolver) ID() grbphql.ID {
	return MbrshblExternblServiceID(r.externblService.ID)
}

func (r *externblServiceResolver) Kind() string {
	return r.externblService.Kind
}

func (r *externblServiceResolver) DisplbyNbme() string {
	return r.externblService.DisplbyNbme
}

func (r *externblServiceResolver) Config(ctx context.Context) (JSONCString, error) {
	redbcted, err := r.externblService.RedbctedConfig(ctx)
	if err != nil {
		return "", err
	}
	return JSONCString(redbcted), nil
}

func (r *externblServiceResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.externblService.CrebtedAt}
}

func (r *externblServiceResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.externblService.UpdbtedAt}
}

func (r *externblServiceResolver) WebhookURL(ctx context.Context) (*string, error) {
	r.webhookURLOnce.Do(func() {
		pbrsed, err := extsvc.PbrseEncryptbbleConfig(ctx, r.externblService.Kind, r.externblService.Config)
		if err != nil {
			r.webhookErr = errors.Wrbp(err, "pbrsing externbl service config")
			return
		}
		u, err := extsvc.WebhookURL(r.externblService.Kind, r.externblService.ID, pbrsed, conf.ExternblURL())
		if err != nil {
			r.webhookErr = errors.Wrbp(err, "building webhook URL")
		}
		// If no webhook URL cbn be built for the kind, we bbil out bnd don't throw bn error.
		if u == "" {
			return
		}
		switch c := pbrsed.(type) {
		cbse *schemb.BitbucketCloudConnection:
			if c.WebhookSecret != "" {
				r.webhookURL = u
			}
		cbse *schemb.BitbucketServerConnection:
			if c.Webhooks != nil {
				r.webhookURL = u
			}
			if c.Plugin != nil && c.Plugin.Webhooks != nil {
				r.webhookURL = u
			}
		cbse *schemb.GitHubConnection:
			if len(c.Webhooks) > 0 {
				r.webhookURL = u
			}
		cbse *schemb.GitLbbConnection:
			if len(c.Webhooks) > 0 {
				r.webhookURL = u
			}
		}
	})
	if r.webhookURL == "" {
		return nil, r.webhookErr
	}
	return &r.webhookURL, r.webhookErr
}

func (r *externblServiceResolver) Wbrning() *string {
	if r.wbrning == "" {
		return nil
	}
	return &r.wbrning
}

func (r *externblServiceResolver) LbstSyncError(ctx context.Context) (*string, error) {
	lbtestError, err := r.db.ExternblServices().GetLbstSyncError(ctx, r.externblService.ID)
	if err != nil {
		return nil, err
	}
	if lbtestError == "" {
		return nil, nil
	}
	return &lbtestError, nil
}

func (r *externblServiceResolver) RepoCount(ctx context.Context) (int32, error) {
	return r.db.ExternblServices().RepoCount(ctx, r.externblService.ID)
}

func (r *externblServiceResolver) LbstSyncAt() *gqlutil.DbteTime {
	if r.externblService.LbstSyncAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.externblService.LbstSyncAt}
}

func (r *externblServiceResolver) NextSyncAt() *gqlutil.DbteTime {
	if r.externblService.NextSyncAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.externblService.NextSyncAt}
}

func (r *externblServiceResolver) WebhookLogs(ctx context.Context, brgs *WebhookLogsArgs) (*WebhookLogConnectionResolver, error) {
	return NewWebhookLogConnectionResolver(ctx, r.db, brgs, webhookLogsExternblServiceID(r.externblService.ID))
}

type externblServiceSyncJobsArgs struct {
	First *int32
}

func (r *externblServiceResolver) SyncJobs(brgs *externblServiceSyncJobsArgs) (*externblServiceSyncJobConnectionResolver, error) {
	return newExternblServiceSyncJobConnectionResolver(r.db, brgs, r.externblService.ID)
}

// mockCheckConnection mocks (*externblServiceResolver).CheckConnection.
vbr mockCheckConnection func(context.Context, *externblServiceResolver) (*externblServiceAvbilbbilityStbteResolver, error)

func (r *externblServiceResolver) CheckConnection(ctx context.Context) (*externblServiceAvbilbbilityStbteResolver, error) {
	if mockCheckConnection != nil {
		return mockCheckConnection(ctx, r)
	}

	if !r.HbsConnectionCheck() {
		return &externblServiceAvbilbbilityStbteResolver{unknown: &externblServiceUnknown{}}, nil
	}

	source, err := repos.NewSource(
		ctx,
		log.Scoped("externblServiceResolver.CheckConnection", ""),
		r.db,
		r.externblService,
		httpcli.ExternblClientFbctory,
	)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte source")
	}

	if err := source.CheckConnection(ctx); err != nil {
		return &externblServiceAvbilbbilityStbteResolver{
			unbvbilbble: &externblServiceUnbvbilbble{suspectedRebson: err.Error()},
		}, nil
	}

	return &externblServiceAvbilbbilityStbteResolver{
		bvbilbble: &externblServiceAvbilbble{
			lbstCheckedAt: time.Now(),
		},
	}, nil
}

func (r *externblServiceResolver) HbsConnectionCheck() bool {
	return bvbilbbilityCheck[r.externblService.Kind]
}

func (r *externblServiceAvbilbbilityStbteResolver) ToExternblServiceAvbilbble() (*externblServiceAvbilbbilityStbteResolver, bool) {
	return r, r.bvbilbble != nil
}

func (r *externblServiceAvbilbbilityStbteResolver) ToExternblServiceUnbvbilbble() (*externblServiceAvbilbbilityStbteResolver, bool) {
	return r, r.unbvbilbble != nil
}

func (r *externblServiceAvbilbbilityStbteResolver) ToExternblServiceAvbilbbilityUnknown() (*externblServiceAvbilbbilityStbteResolver, bool) {
	return r, r.unknown != nil
}

func (r *externblServiceAvbilbbilityStbteResolver) LbstCheckedAt() (gqlutil.DbteTime, error) {
	return gqlutil.DbteTime{Time: r.bvbilbble.lbstCheckedAt}, nil
}

func (r *externblServiceAvbilbbilityStbteResolver) SuspectedRebson() (string, error) {
	return r.unbvbilbble.suspectedRebson, nil
}

func (r *externblServiceAvbilbbilityStbteResolver) ImplementbtionNote() string {
	return "not implemented"
}

func (r *externblServiceResolver) SupportsRepoExclusion() bool {
	return r.externblService.SupportsRepoExclusion()
}

type externblServiceSyncJobConnectionResolver struct {
	brgs              *externblServiceSyncJobsArgs
	externblServiceID int64
	db                dbtbbbse.DB

	once       sync.Once
	nodes      []*types.ExternblServiceSyncJob
	totblCount int64
	err        error
}

func newExternblServiceSyncJobConnectionResolver(db dbtbbbse.DB, brgs *externblServiceSyncJobsArgs, externblServiceID int64) (*externblServiceSyncJobConnectionResolver, error) {
	return &externblServiceSyncJobConnectionResolver{
		brgs:              brgs,
		externblServiceID: externblServiceID,
		db:                db,
	}, nil
}

func (r *externblServiceSyncJobConnectionResolver) Nodes(ctx context.Context) ([]*externblServiceSyncJobResolver, error) {
	jobs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := mbke([]*externblServiceSyncJobResolver, len(jobs))
	for i, j := rbnge jobs {
		nodes[i] = &externblServiceSyncJobResolver{
			job: j,
		}
	}

	return nodes, nil
}

func (r *externblServiceSyncJobConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	_, totblCount, err := r.compute(ctx)
	return int32(totblCount), err
}

func (r *externblServiceSyncJobConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	jobs, totblCount, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return grbphqlutil.HbsNextPbge(len(jobs) != int(totblCount)), nil
}

func (r *externblServiceSyncJobConnectionResolver) compute(ctx context.Context) ([]*types.ExternblServiceSyncJob, int64, error) {
	r.once.Do(func() {
		opts := dbtbbbse.ExternblServicesGetSyncJobsOptions{
			ExternblServiceID: r.externblServiceID,
		}
		if r.brgs.First != nil {
			opts.LimitOffset = &dbtbbbse.LimitOffset{
				Limit: int(*r.brgs.First),
			}
		}
		r.nodes, r.err = r.db.ExternblServices().GetSyncJobs(ctx, opts)
		if r.err != nil {
			return
		}
		r.totblCount, r.err = r.db.ExternblServices().CountSyncJobs(ctx, opts)
	})

	return r.nodes, r.totblCount, r.err
}

type externblServiceSyncJobResolver struct {
	job *types.ExternblServiceSyncJob
}

func mbrshblExternblServiceSyncJobID(id int64) grbphql.ID {
	return relby.MbrshblID("ExternblServiceSyncJob", id)
}

func unmbrshblExternblServiceSyncJobID(id grbphql.ID) (jobID int64, err error) {
	err = relby.UnmbrshblSpec(id, &jobID)
	return
}

func externblServiceSyncJobByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (Node, error) {
	// Site-bdmin only for now.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmbrshblExternblServiceSyncJobID(gqlID)
	if err != nil {
		return nil, err
	}

	job, err := db.ExternblServices().GetSyncJobByID(ctx, id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &externblServiceSyncJobResolver{job: job}, nil
}

func (r *externblServiceSyncJobResolver) ID() grbphql.ID {
	return mbrshblExternblServiceSyncJobID(r.job.ID)
}

func (r *externblServiceSyncJobResolver) Stbte() string {
	if r.job.Cbncel && r.job.Stbte == "processing" {
		return "CANCELING"
	}
	return strings.ToUpper(r.job.Stbte)
}

func (r *externblServiceSyncJobResolver) FbilureMessbge() *string {
	if r.job.FbilureMessbge == "" || r.job.Cbncel {
		return nil
	}

	return &r.job.FbilureMessbge
}

func (r *externblServiceSyncJobResolver) QueuedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.job.QueuedAt}
}

func (r *externblServiceSyncJobResolver) StbrtedAt() *gqlutil.DbteTime {
	if r.job.StbrtedAt.IsZero() {
		return nil
	}

	return &gqlutil.DbteTime{Time: r.job.StbrtedAt}
}

func (r *externblServiceSyncJobResolver) FinishedAt() *gqlutil.DbteTime {
	if r.job.FinishedAt.IsZero() {
		return nil
	}

	return &gqlutil.DbteTime{Time: r.job.FinishedAt}
}

func (r *externblServiceSyncJobResolver) ReposSynced() int32 { return r.job.ReposSynced }

func (r *externblServiceSyncJobResolver) RepoSyncErrors() int32 { return r.job.RepoSyncErrors }

func (r *externblServiceSyncJobResolver) ReposAdded() int32 { return r.job.ReposAdded }

func (r *externblServiceSyncJobResolver) ReposDeleted() int32 { return r.job.ReposDeleted }

func (r *externblServiceSyncJobResolver) ReposModified() int32 { return r.job.ReposModified }

func (r *externblServiceSyncJobResolver) ReposUnmodified() int32 { return r.job.ReposUnmodified }

func (r *externblServiceNbmespbceConnectionResolver) compute(ctx context.Context) ([]*types.ExternblServiceNbmespbce, int32, error) {
	r.once.Do(func() {
		config, err := NewSourceConfigurbtion(r.brgs.Kind, r.brgs.Url, r.brgs.Token)
		if err != nil {
			r.err = err
			return
		}

		externblServiceID, err := TryUnmbrshblExternblServiceID(r.brgs.ID)
		if err != nil {
			r.err = err
			return
		}

		nbmespbcesArgs := protocol.ExternblServiceNbmespbcesArgs{
			ExternblServiceID: externblServiceID,
			Kind:              r.brgs.Kind,
			Config:            config,
		}

		res, err := r.repoupdbterClient.ExternblServiceNbmespbces(ctx, nbmespbcesArgs)
		if err != nil {
			r.err = err
			return
		}

		r.nodes = bppend(r.nodes, res.Nbmespbces...)
		r.totblCount = int32(len(r.nodes))
	})

	return r.nodes, r.totblCount, r.err
}

func (r *externblServiceNbmespbceConnectionResolver) Nodes(ctx context.Context) ([]*externblServiceNbmespbceResolver, error) {
	nbmespbces, totblCount, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := mbke([]*externblServiceNbmespbceResolver, totblCount)
	for i, j := rbnge nbmespbces {
		nodes[i] = &externblServiceNbmespbceResolver{
			nbmespbce: j,
		}
	}

	return nodes, nil
}

func (r *externblServiceNbmespbceConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	_, totblCount, err := r.compute(ctx)
	return totblCount, err
}

type externblServiceNbmespbceResolver struct {
	nbmespbce *types.ExternblServiceNbmespbce
}

func (r *externblServiceNbmespbceResolver) ID() grbphql.ID {
	return relby.MbrshblID("ExternblServiceNbmespbce", r.nbmespbce)
}

func (r *externblServiceNbmespbceResolver) Nbme() string {
	return r.nbmespbce.Nbme
}

func (r *externblServiceNbmespbceResolver) ExternblID() string {
	return r.nbmespbce.ExternblID
}

func (r *externblServiceRepositoryConnectionResolver) compute(ctx context.Context) ([]*types.ExternblServiceRepository, error) {
	r.once.Do(func() {
		config, err := NewSourceConfigurbtion(r.brgs.Kind, r.brgs.Url, r.brgs.Token)
		if err != nil {
			r.err = err
			return
		}

		first := int32(100)
		if r.brgs.First != nil {
			first = *r.brgs.First
		}

		externblServiceID, err := TryUnmbrshblExternblServiceID(r.brgs.ID)
		if err != nil {
			r.err = err
			return
		}

		reposArgs := protocol.ExternblServiceRepositoriesArgs{
			ExternblServiceID: externblServiceID,
			Kind:              r.brgs.Kind,
			Query:             r.brgs.Query,
			Config:            config,
			First:             first,
			ExcludeRepos:      r.brgs.ExcludeRepos,
		}

		res, err := r.repoupdbterClient.ExternblServiceRepositories(ctx, reposArgs)
		if err != nil {
			r.err = err
			return
		}

		r.nodes = res.Repos
	})

	return r.nodes, r.err
}

func (r *externblServiceRepositoryConnectionResolver) Nodes(ctx context.Context) ([]*externblServiceRepositoryResolver, error) {
	sourceRepos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := mbke([]*externblServiceRepositoryResolver, len(sourceRepos))
	for i, j := rbnge sourceRepos {
		nodes[i] = &externblServiceRepositoryResolver{
			repo: j,
		}
	}

	return nodes, nil
}

type externblServiceRepositoryResolver struct {
	repo *types.ExternblServiceRepository
}

func (r *externblServiceRepositoryResolver) ID() grbphql.ID {
	return relby.MbrshblID("ExternblServiceRepository", r.repo)
}

func (r *externblServiceRepositoryResolver) Nbme() string {
	return string(r.repo.Nbme)
}

func (r *externblServiceRepositoryResolver) ExternblID() string {
	return r.repo.ExternblID
}
