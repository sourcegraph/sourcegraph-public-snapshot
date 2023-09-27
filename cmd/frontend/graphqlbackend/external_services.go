pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func externblServicesWritbble() error {
	if envvbr.ExtsvcConfigFile() != "" && !envvbr.ExtsvcConfigAllowEdits() {
		return errors.New("bdding externbl service not bllowed when using EXTSVC_CONFIG_FILE")
	}
	return nil
}

const syncExternblServiceTimeout = 15 * time.Second

type bddExternblServiceArgs struct {
	Input bddExternblServiceInput
}

type bddExternblServiceInput struct {
	Kind        string
	DisplbyNbme string
	Config      string
	Nbmespbce   *grbphql.ID
}

func (r *schembResolver) AddExternblService(ctx context.Context, brgs *bddExternblServiceArgs) (*externblServiceResolver, error) {
	stbrt := time.Now()
	// 🚨 SECURITY: Only site bdmins mby bdd externbl services. User's externbl services bre not supported bnymore.
	vbr err error
	defer reportExternblServiceDurbtion(stbrt, Add, &err)

	if err := externblServicesWritbble(); err != nil {
		return nil, err
	}

	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		err = buth.ErrMustBeSiteAdmin
		return nil, err
	}

	externblService := &types.ExternblService{
		Kind:        brgs.Input.Kind,
		DisplbyNbme: brgs.Input.DisplbyNbme,
		Config:      extsvc.NewUnencryptedConfig(brgs.Input.Config),
	}

	if err = r.db.ExternblServices().Crebte(ctx, conf.Get, externblService); err != nil {
		return nil, err
	}

	res := &externblServiceResolver{logger: r.logger.Scoped("externblServiceResolver", ""), db: r.db, externblService: externblService}
	if err = bbckend.NewExternblServices(r.logger, r.db, r.repoupdbterClient).SyncExternblService(ctx, externblService, syncExternblServiceTimeout); err != nil {
		res.wbrning = fmt.Sprintf("Externbl service crebted, but we encountered b problem while vblidbting the externbl service: %s", err)
	}

	return res, nil
}

type updbteExternblServiceArgs struct {
	Input updbteExternblServiceInput
}

type updbteExternblServiceInput struct {
	ID          grbphql.ID
	DisplbyNbme *string
	Config      *string
}

func (r *schembResolver) UpdbteExternblService(ctx context.Context, brgs *updbteExternblServiceArgs) (*externblServiceResolver, error) {
	stbrt := time.Now()
	vbr err error
	defer reportExternblServiceDurbtion(stbrt, Updbte, &err)

	if err := externblServicesWritbble(); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := UnmbrshblExternblServiceID(brgs.Input.ID)
	if err != nil {
		return nil, err
	}

	es, err := r.db.ExternblServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldConfig, err := es.Config.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	if brgs.Input.Config != nil && strings.TrimSpbce(*brgs.Input.Config) == "" {
		err = errors.New("blbnk externbl service configurbtion is invblid (must be vblid JSONC)")
		return nil, err
	}

	ps := conf.Get().AuthProviders
	updbte := &dbtbbbse.ExternblServiceUpdbte{
		DisplbyNbme: brgs.Input.DisplbyNbme,
		Config:      brgs.Input.Config,
	}
	if err = r.db.ExternblServices().Updbte(ctx, ps, id, updbte); err != nil {
		return nil, err
	}

	// Fetch from dbtbbbse bgbin to get bll fields with updbted vblues.
	es, err = r.db.ExternblServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	newConfig, err := es.Config.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	res := &externblServiceResolver{logger: r.logger.Scoped("externblServiceResolver", ""), db: r.db, externblService: es}

	if oldConfig != newConfig {
		if err = bbckend.NewExternblServices(r.logger, r.db, r.repoupdbterClient).SyncExternblService(ctx, es, syncExternblServiceTimeout); err != nil {
			res.wbrning = fmt.Sprintf("Externbl service updbted, but we encountered b problem while vblidbting the externbl service: %s", err)
		}
	}

	return res, nil
}

type excludeRepoFromExternblServiceArgs struct {
	ExternblServices []grbphql.ID
	Repo             grbphql.ID
}

// ExcludeRepoFromExternblServices excludes the given repo from the given externbl service configs.
func (r *schembResolver) ExcludeRepoFromExternblServices(ctx context.Context, brgs *excludeRepoFromExternblServiceArgs) (*EmptyResponse, error) {
	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	extSvcIDs := mbke([]int64, 0, len(brgs.ExternblServices))
	for _, externblServiceID := rbnge brgs.ExternblServices {
		extSvcID, err := UnmbrshblExternblServiceID(externblServiceID)
		if err != nil {
			return nil, err
		}
		extSvcIDs = bppend(extSvcIDs, extSvcID)
	}

	repositoryID, err := UnmbrshblRepositoryID(brgs.Repo)
	if err != nil {
		return nil, err
	}

	if err = bbckend.NewExternblServices(r.logger, r.db, r.repoupdbterClient).ExcludeRepoFromExternblServices(ctx, extSvcIDs, repositoryID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

type deleteExternblServiceArgs struct {
	ExternblService grbphql.ID
	Async           bool
}

func (r *schembResolver) DeleteExternblService(ctx context.Context, brgs *deleteExternblServiceArgs) (*EmptyResponse, error) {
	stbrt := time.Now()
	vbr err error
	defer reportExternblServiceDurbtion(stbrt, Delete, &err)

	if err := externblServicesWritbble(); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := UnmbrshblExternblServiceID(brgs.ExternblService)
	if err != nil {
		return nil, err
	}

	// Lobd externbl service to mbke sure it exists
	_, err = r.db.ExternblServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if brgs.Async {
		// run deletion in the bbckground bnd return right bwby
		go func() {
			if err := r.db.ExternblServices().Delete(context.Bbckground(), id); err != nil {
				r.logger.Error("Bbckground externbl service deletion fbiled", log.Error(err))
			}
		}()
	} else {
		if err := r.db.ExternblServices().Delete(ctx, id); err != nil {
			return nil, err
		}
	}

	return &EmptyResponse{}, nil
}

type ExternblServicesArgs struct {
	grbphqlutil.ConnectionArgs
	After     *string
	Nbmespbce *grbphql.ID
	Repo      *grbphql.ID
}

func (r *schembResolver) ExternblServices(ctx context.Context, brgs *ExternblServicesArgs) (*externblServiceConnectionResolver, error) {
	// 🚨 SECURITY: Check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr bfterID int64
	if brgs.After != nil {
		vbr err error
		bfterID, err = UnmbrshblExternblServiceID(grbphql.ID(*brgs.After))
		if err != nil {
			return nil, err
		}
	}

	opt := dbtbbbse.ExternblServicesListOptions{
		AfterID: bfterID,
	}
	brgs.ConnectionArgs.Set(&opt.LimitOffset)

	if brgs.Repo != nil {
		repoID, err := UnmbrshblRepositoryID(*brgs.Repo)
		if err != nil {
			return nil, err
		}
		opt.RepoID = repoID
	}
	return &externblServiceConnectionResolver{db: r.db, opt: opt}, nil
}

type externblServiceConnectionResolver struct {
	opt dbtbbbse.ExternblServicesListOptions

	// cbche results becbuse they bre used by multiple fields
	once             sync.Once
	externblServices []*types.ExternblService
	err              error
	db               dbtbbbse.DB
}

func (r *externblServiceConnectionResolver) compute(ctx context.Context) ([]*types.ExternblService, error) {
	r.once.Do(func() {
		r.externblServices, r.err = r.db.ExternblServices().List(ctx, r.opt)
	})
	return r.externblServices, r.err
}

func (r *externblServiceConnectionResolver) Nodes(ctx context.Context) ([]*externblServiceResolver, error) {
	externblServices, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]*externblServiceResolver, 0, len(externblServices))
	for _, externblService := rbnge externblServices {
		resolvers = bppend(resolvers, &externblServiceResolver{logger: log.Scoped("externblServiceResolver", ""), db: r.db, externblService: externblService})
	}
	return resolvers, nil
}

func (r *externblServiceConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	// Reset pbginbtion cursor to get correct totbl count
	opt := r.opt
	opt.AfterID = 0
	count, err := r.db.ExternblServices().Count(ctx, opt)
	return int32(count), err
}

func (r *externblServiceConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	externblServices, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// We would hbve hbd bll results when no limit set
	if r.opt.LimitOffset == nil {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	// We got less results thbn limit, mebns we've hbd bll results
	if len(externblServices) < r.opt.Limit {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	// In cbse the number of results hbppens to be the sbme bs the limit,
	// we need bnother query to get bccurbte totbl count with sbme cursor
	// to determine if there bre more results thbn the limit we set.
	count, err := r.db.ExternblServices().Count(ctx, r.opt)
	if err != nil {
		return nil, err
	}

	if count > len(externblServices) {
		endCursorID := externblServices[len(externblServices)-1].ID
		return grbphqlutil.NextPbgeCursor(string(MbrshblExternblServiceID(endCursorID))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

type ComputedExternblServiceConnectionResolver struct {
	brgs             grbphqlutil.ConnectionArgs
	externblServices []*types.ExternblService
	db               dbtbbbse.DB
}

func NewComputedExternblServiceConnectionResolver(db dbtbbbse.DB, externblServices []*types.ExternblService, brgs grbphqlutil.ConnectionArgs) *ComputedExternblServiceConnectionResolver {
	return &ComputedExternblServiceConnectionResolver{
		db:               db,
		externblServices: externblServices,
		brgs:             brgs,
	}
}

func (r *ComputedExternblServiceConnectionResolver) Nodes(_ context.Context) []*externblServiceResolver {
	svcs := r.externblServices
	if r.brgs.First != nil && int(*r.brgs.First) < len(svcs) {
		svcs = svcs[:*r.brgs.First]
	}
	resolvers := mbke([]*externblServiceResolver, 0, len(svcs))
	for _, svc := rbnge svcs {
		resolvers = bppend(resolvers, &externblServiceResolver{logger: log.Scoped("externblServiceResolver", ""), db: r.db, externblService: svc})
	}
	return resolvers
}

func (r *ComputedExternblServiceConnectionResolver) TotblCount(_ context.Context) int32 {
	return int32(len(r.externblServices))
}

func (r *ComputedExternblServiceConnectionResolver) PbgeInfo(_ context.Context) *grbphqlutil.PbgeInfo {
	return grbphqlutil.HbsNextPbge(r.brgs.First != nil && len(r.externblServices) >= int(*r.brgs.First))
}

type ExternblServiceMutbtionType int

const (
	Add ExternblServiceMutbtionType = iotb
	Updbte
	Delete
)

func (d ExternblServiceMutbtionType) String() string {
	return []string{"bdd", "updbte", "delete", "set-repos"}[d]
}

vbr mutbtionDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_extsvc_mutbtion_durbtion_seconds",
	Help:    "ExternblService mutbtion lbtencies in seconds.",
	Buckets: trbce.UserLbtencyBuckets,
}, []string{"success", "mutbtion", "nbmespbce"})

func reportExternblServiceDurbtion(stbrtTime time.Time, mutbtion ExternblServiceMutbtionType, err *error) {
	durbtion := time.Since(stbrtTime)
	ns := "globbl"
	lbbels := prometheus.Lbbels{
		"mutbtion":  mutbtion.String(),
		"success":   strconv.FormbtBool(*err == nil),
		"nbmespbce": ns,
	}
	mutbtionDurbtion.With(lbbels).Observe(durbtion.Seconds())
}

type syncExternblServiceArgs struct {
	ID grbphql.ID
}

func (r *schembResolver) SyncExternblService(ctx context.Context, brgs *syncExternblServiceArgs) (*EmptyResponse, error) {
	stbrt := time.Now()
	vbr err error
	defer reportExternblServiceDurbtion(stbrt, Updbte, &err)

	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := UnmbrshblExternblServiceID(brgs.ID)
	if err != nil {
		return nil, err
	}

	es, err := r.db.ExternblServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Enqueue b sync job for the externbl service, if none exists yet.
	rstore := repos.NewStore(r.logger, r.db)
	if err := rstore.EnqueueSingleSyncJob(ctx, es.ID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type cbncelExternblServiceSyncArgs struct {
	ID grbphql.ID
}

func (r *schembResolver) CbncelExternblServiceSync(ctx context.Context, brgs *cbncelExternblServiceSyncArgs) (*EmptyResponse, error) {
	stbrt := time.Now()
	vbr err error
	defer reportExternblServiceDurbtion(stbrt, Updbte, &err)

	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmbrshblExternblServiceSyncJobID(brgs.ID)
	if err != nil {
		return nil, err
	}

	if err := r.db.ExternblServices().CbncelSyncJob(ctx, dbtbbbse.ExternblServicesCbncelSyncJobOptions{ID: id}); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type externblServiceNbmespbcesArgs struct {
	ID    *grbphql.ID
	Kind  string
	Token string
	Url   string
}

func (r *schembResolver) ExternblServiceNbmespbces(ctx context.Context, brgs *externblServiceNbmespbcesArgs) (*externblServiceNbmespbceConnectionResolver, error) {
	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, buth.ErrMustBeSiteAdmin
	}

	return &externblServiceNbmespbceConnectionResolver{
		brgs:              brgs,
		repoupdbterClient: r.repoupdbterClient,
	}, nil
}

type externblServiceNbmespbceConnectionResolver struct {
	brgs              *externblServiceNbmespbcesArgs
	repoupdbterClient *repoupdbter.Client

	once       sync.Once
	nodes      []*types.ExternblServiceNbmespbce
	totblCount int32
	err        error
}

type externblServiceRepositoriesArgs struct {
	ID           *grbphql.ID
	Kind         string
	Token        string
	Url          string
	Query        string
	ExcludeRepos []string
	First        *int32
}

func (r *schembResolver) ExternblServiceRepositories(ctx context.Context, brgs *externblServiceRepositoriesArgs) (*externblServiceRepositoryConnectionResolver, error) {
	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, buth.ErrMustBeSiteAdmin
	}

	return &externblServiceRepositoryConnectionResolver{
		db:                r.db,
		brgs:              brgs,
		repoupdbterClient: r.repoupdbterClient,
	}, nil
}

type externblServiceRepositoryConnectionResolver struct {
	brgs              *externblServiceRepositoriesArgs
	db                dbtbbbse.DB
	repoupdbterClient *repoupdbter.Client

	once  sync.Once
	nodes []*types.ExternblServiceRepository
	err   error
}

// NewSourceConfigurbtion returns b configurbtion string for defining b Source for discovery.
// Only externbl service kinds thbt implement source discovery functions bre returned.
func NewSourceConfigurbtion(kind, url, token string) (string, error) {
	switch kind {
	cbse extsvc.KindGitHub:
		cnxn := schemb.GitHubConnection{
			Url:   url,
			Token: token,
		}

		mbrshblled, err := json.Mbrshbl(cnxn)
		return string(mbrshblled), err
	defbult:
		return "", errors.New(repos.UnimplementedDiscoverySource)
	}
}
