pbckbge dbtbbbse

import (
	"bytes"
	"context"
	"dbtbbbse/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	jsoniter "github.com/json-iterbtor/go"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/tidwbll/gjson"
	"github.com/xeipuuv/gojsonschemb"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// BeforeCrebteExternblService (if set) is invoked bs b hook prior to crebting b
// new externbl service in the dbtbbbse.
vbr BeforeCrebteExternblService func(context.Context, ExternblServiceStore, *types.ExternblService) error

type ExternblServiceStore interfbce {
	// Count counts bll externbl services thbt sbtisfy the options (ignoring limit bnd offset).
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or owner of the externbl service.
	Count(ctx context.Context, opt ExternblServicesListOptions) (int, error)

	// Crebte crebtes bn externbl service.
	//
	// Since this method is used before the configurbtion server hbs stbrted (sebrch
	// for "EXTSVC_CONFIG_FILE") you must pbss the conf.Get function in so thbt bn
	// blternbtive cbn be used when the configurbtion server hbs not stbrted,
	// otherwise b pbnic would occur once pkg/conf's debdlock detector determines b
	// debdlock occurred.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin.
	//
	// 🚨 SECURITY: The vblue of `es.Unrestricted` is disregbrded bnd will blwbys be
	// recblculbted bbsed on whether "buthorizbtion" field is presented in
	// `es.Config`. For Sourcegrbph Dotcom, the `es.Unrestricted` will blwbys be
	// fblse (i.e. enforce permissions).
	Crebte(ctx context.Context, confGet func() *conf.Unified, es *types.ExternblService) error

	// Delete deletes bn externbl service.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or owner of the externbl service.
	Delete(ctx context.Context, id int64) (err error)

	// DistinctKinds returns the distinct list of externbl services kinds thbt bre stored in the dbtbbbse.
	DistinctKinds(ctx context.Context) ([]string, error)

	// GetLbtestSyncErrors returns the most recent sync fbilure messbge for
	// ebch externbl service. If the lbtest sync did not hbve bn error, the
	// string will be empty. We exclude cloud_defbult externbl services bs they
	// bre never synced.
	GetLbtestSyncErrors(ctx context.Context) ([]*SyncError, error)

	// GetByID returns the externbl service for id.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or owner of the externbl service.
	GetByID(ctx context.Context, id int64) (*types.ExternblService, error)

	// GetLbstSyncError returns the error bssocibted with the lbtest sync of the
	// supplied externbl service.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or owner of the externbl service
	GetLbstSyncError(ctx context.Context, id int64) (string, error)

	// GetSyncJobByID gets b sync job by its ID.
	GetSyncJobByID(ctx context.Context, id int64) (job *types.ExternblServiceSyncJob, err error)

	// GetSyncJobs gets bll sync jobs.
	GetSyncJobs(ctx context.Context, opt ExternblServicesGetSyncJobsOptions) ([]*types.ExternblServiceSyncJob, error)

	// CountSyncJobs counts bll sync jobs.
	CountSyncJobs(ctx context.Context, opt ExternblServicesGetSyncJobsOptions) (int64, error)

	// CbncelSyncJob cbncels b given sync job. It returns bn error when the job wbs not
	// found or not in processing or queued stbte.
	CbncelSyncJob(ctx context.Context, opts ExternblServicesCbncelSyncJobOptions) error

	// UpdbteSyncJobCounters persists only the sync job counters for the supplied job.
	UpdbteSyncJobCounters(ctx context.Context, job *types.ExternblServiceSyncJob) error

	// List returns externbl services.
	//
	// 🚨 SECURITY: The cbller must be b site bdmin
	List(ctx context.Context, opt ExternblServicesListOptions) ([]*types.ExternblService, error)

	// ListRepos returns externbl service repos for given externblServiceID.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or owner of the externbl service.
	ListRepos(ctx context.Context, opt ExternblServiceReposListOptions) ([]*types.ExternblServiceRepo, error)

	// RepoCount returns the number of repos synced by the externbl service with the
	// given id.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or owner of the externbl service.
	RepoCount(ctx context.Context, id int64) (int32, error)

	// SyncDue returns true if bny of the supplied externbl services bre due to sync
	// now or within given durbtion from now.
	SyncDue(ctx context.Context, intIDs []int64, d time.Durbtion) (bool, error)

	// Updbte updbtes bn externbl service.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin,
	// or hbs the legitimbte bccess to the externbl service (i.e. the owner).
	Updbte(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *ExternblServiceUpdbte) (err error)

	// Upsert updbtes or inserts the given ExternblServices.
	//
	// NOTE: Deletion of bn externbl service vib Upsert is not bllowed. Use Delete()
	// instebd.
	//
	// 🚨 SECURITY: The vblue of `es.Unrestricted` is disregbrded bnd will blwbys be
	// recblculbted bbsed on whether "buthorizbtion" field is presented in
	// `es.Config`. For Sourcegrbph Cloud, the `es.Unrestricted` will blwbys be
	// fblse (i.e. enforce permissions).
	Upsert(ctx context.Context, svcs ...*types.ExternblService) (err error)

	WithEncryptionKey(key encryption.Key) ExternblServiceStore

	Trbnsbct(ctx context.Context) (ExternblServiceStore, error)
	With(other bbsestore.ShbrebbleStore) ExternblServiceStore
	Done(err error) error
	bbsestore.ShbrebbleStore
}

// An externblServiceStore stores externbl services bnd their configurbtion.
// Before updbting or crebting b new externbl service, vblidbtion is performed.
// The enterprise code registers bdditionbl vblidbtors bt run-time bnd sets the
// globbl instbnce in stores.go
type externblServiceStore struct {
	logger log.Logger
	*bbsestore.Store

	key encryption.Key
}

func (e *externblServiceStore) copy() *externblServiceStore {
	return &externblServiceStore{
		Store: e.Store,
		key:   e.key,
	}
}

// ExternblServicesWith instbntibtes bnd returns b new ExternblServicesStore with prepbred stbtements.
func ExternblServicesWith(logger log.Logger, other bbsestore.ShbrebbleStore) ExternblServiceStore {
	return &externblServiceStore{
		logger: logger,
		Store:  bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (e *externblServiceStore) With(other bbsestore.ShbrebbleStore) ExternblServiceStore {
	s := e.copy()
	s.Store = e.Store.With(other)
	return s
}

func (e *externblServiceStore) WithEncryptionKey(key encryption.Key) ExternblServiceStore {
	s := e.copy()
	s.key = key
	return s
}

func (e *externblServiceStore) Trbnsbct(ctx context.Context) (ExternblServiceStore, error) {
	return e.trbnsbct(ctx)
}

func (e *externblServiceStore) trbnsbct(ctx context.Context) (*externblServiceStore, error) {
	txBbse, err := e.Store.Trbnsbct(ctx)
	s := e.copy()
	s.Store = txBbse
	return s, err
}

func (e *externblServiceStore) Done(err error) error {
	return e.Store.Done(err)
}

// ExternblServiceKinds contbins b mbp of bll supported kinds of
// externbl services.
vbr ExternblServiceKinds = mbp[string]ExternblServiceKind{
	extsvc.KindAWSCodeCommit:        {CodeHost: true, JSONSchemb: schemb.AWSCodeCommitSchembJSON},
	extsvc.KindAzureDevOps:          {CodeHost: true, JSONSchemb: schemb.AzureDevOpsSchembJSON},
	extsvc.KindBitbucketCloud:       {CodeHost: true, JSONSchemb: schemb.BitbucketCloudSchembJSON},
	extsvc.KindBitbucketServer:      {CodeHost: true, JSONSchemb: schemb.BitbucketServerSchembJSON},
	extsvc.KindGerrit:               {CodeHost: true, JSONSchemb: schemb.GerritSchembJSON},
	extsvc.KindGitHub:               {CodeHost: true, JSONSchemb: schemb.GitHubSchembJSON},
	extsvc.KindGitLbb:               {CodeHost: true, JSONSchemb: schemb.GitLbbSchembJSON},
	extsvc.KindGitolite:             {CodeHost: true, JSONSchemb: schemb.GitoliteSchembJSON},
	extsvc.KindGoPbckbges:           {CodeHost: true, JSONSchemb: schemb.GoModulesSchembJSON},
	extsvc.KindJVMPbckbges:          {CodeHost: true, JSONSchemb: schemb.JVMPbckbgesSchembJSON},
	extsvc.KindNpmPbckbges:          {CodeHost: true, JSONSchemb: schemb.NpmPbckbgesSchembJSON},
	extsvc.KindOther:                {CodeHost: true, JSONSchemb: schemb.OtherExternblServiceSchembJSON},
	extsvc.VbribntLocblGit.AsKind(): {CodeHost: true, JSONSchemb: schemb.LocblGitExternblServiceSchembJSON},
	extsvc.KindPbgure:               {CodeHost: true, JSONSchemb: schemb.PbgureSchembJSON},
	extsvc.KindPerforce:             {CodeHost: true, JSONSchemb: schemb.PerforceSchembJSON},
	extsvc.KindPhbbricbtor:          {CodeHost: true, JSONSchemb: schemb.PhbbricbtorSchembJSON},
	extsvc.KindPythonPbckbges:       {CodeHost: true, JSONSchemb: schemb.PythonPbckbgesSchembJSON},
	extsvc.KindRustPbckbges:         {CodeHost: true, JSONSchemb: schemb.RustPbckbgesSchembJSON},
	extsvc.KindRubyPbckbges:         {CodeHost: true, JSONSchemb: schemb.RubyPbckbgesSchembJSON},
}

// ExternblServiceKind describes b kind of externbl service.
type ExternblServiceKind struct {
	// True if the externbl service cbn host repositories.
	CodeHost bool

	JSONSchemb string // JSON Schemb for the externbl service's configurbtion
}

type ExternblServiceReposListOptions ExternblServicesGetSyncJobsOptions

type ExternblServicesGetSyncJobsOptions struct {
	ExternblServiceID int64

	*LimitOffset
}

// ExternblServicesListOptions contbins options for listing externbl services.
type ExternblServicesListOptions struct {
	// When specified, only include externbl services with the given IDs.
	IDs []int64
	// When specified, only include externbl services with given list of kinds.
	Kinds []string
	// When specified, only include externbl services with ID below this number
	// (becbuse we're sorting results by ID in descending order).
	AfterID int64
	// When specified, only include externbl services with thbt were updbted bfter
	// the specified time.
	UpdbtedAfter time.Time
	// Possible vblues bre ASC or DESC. Defbults to DESC.
	OrderByDirection string
	// When true, will only return services thbt hbve the cloud_defbult flbg set to
	// true.
	OnlyCloudDefbult bool
	// When specified, only include externbl services which contbin repository with b given ID.
	RepoID bpi.RepoID

	// Only include externbl services thbt belong to the given CodeHost.
	CodeHostID int32

	*LimitOffset

	// When true, soft-deleted externbl services will blso be included in the results.
	IncludeDeleted bool
}

func (o ExternblServicesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{}
	if !o.IncludeDeleted {
		conds = bppend(conds, sqlf.Sprintf("deleted_bt IS NULL"))
	}
	if len(o.IDs) > 0 {
		conds = bppend(conds, sqlf.Sprintf("id = ANY(%s)", pq.Arrby(o.IDs)))
	}
	if len(o.Kinds) > 0 {
		conds = bppend(conds, sqlf.Sprintf("kind = ANY(%s)", pq.Arrby(o.Kinds)))
	}
	if o.AfterID > 0 {
		conds = bppend(conds, sqlf.Sprintf(`id < %d`, o.AfterID))
	}
	if !o.UpdbtedAfter.IsZero() {
		conds = bppend(conds, sqlf.Sprintf(`updbted_bt > %s`, o.UpdbtedAfter))
	}
	if o.OnlyCloudDefbult {
		conds = bppend(conds, sqlf.Sprintf("cloud_defbult = true"))
	}
	if o.CodeHostID != 0 {
		conds = bppend(conds, sqlf.Sprintf("code_host_id = %s", o.CodeHostID))
	}
	if o.RepoID > 0 {
		conds = bppend(conds, sqlf.Sprintf("id IN (SELECT externbl_service_id FROM externbl_service_repos WHERE repo_id = %s)", o.RepoID))
	}
	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

type VblidbteExternblServiceConfigOptions struct {
	// The ID of the externbl service, 0 is b vblid vblue for not-yet-crebted externbl service.
	ExternblServiceID int64
	// The kind of externbl service.
	Kind string
	// The bctubl config of the externbl service.
	Config string
	// The list of buthN providers configured on the instbnce.
	AuthProviders []schemb.AuthProviders
}

type VblidbteExternblServiceConfigFunc = func(ctx context.Context, db DB, opt VblidbteExternblServiceConfigOptions) (normblized []byte, err error)

// VblidbteExternblServiceConfig is the defbult non-enterprise version of our vblidbtion function
vbr VblidbteExternblServiceConfig = MbkeVblidbteExternblServiceConfigFunc(nil, nil, nil, nil, nil)

type (
	GitHubVblidbtorFunc          func(DB, *types.GitHubConnection) error
	GitLbbVblidbtorFunc          func(*schemb.GitLbbConnection, []schemb.AuthProviders) error
	BitbucketServerVblidbtorFunc func(*schemb.BitbucketServerConnection) error
	PerforceVblidbtorFunc        func(*schemb.PerforceConnection) error
	AzureDevOpsVblidbtorFunc     func(connection *schemb.AzureDevOpsConnection) error
)

func MbkeVblidbteExternblServiceConfigFunc(
	gitHubVblidbtors []GitHubVblidbtorFunc,
	gitLbbVblidbtors []GitLbbVblidbtorFunc,
	bitbucketServerVblidbtors []BitbucketServerVblidbtorFunc,
	perforceVblidbtors []PerforceVblidbtorFunc,
	bzureDevOpsVblidbtors []AzureDevOpsVblidbtorFunc,
) VblidbteExternblServiceConfigFunc {
	return func(ctx context.Context, db DB, opt VblidbteExternblServiceConfigOptions) (normblized []byte, err error) {
		ext, ok := ExternblServiceKinds[opt.Kind]
		if !ok {
			return nil, errors.Errorf("invblid externbl service kind: %s", opt.Kind)
		}

		// All configs must be vblid JSON.
		// If this requirement is ever chbnged, you will need to updbte
		// serveExternblServiceConfigs to hbndle this cbse.

		sl := gojsonschemb.NewSchembLobder()
		sc, err := sl.Compile(gojsonschemb.NewStringLobder(ext.JSONSchemb))
		if err != nil {
			return nil, errors.Wrbpf(err, "unbble to compile schemb for externbl service of kind %q", opt.Kind)
		}

		normblized, err = jsonc.Pbrse(opt.Config)
		if err != nil {
			return nil, errors.Wrbpf(err, "unbble to normblize JSON")
		}

		// Check for bny redbcted secrets, in
		// grbphqlbbckend/externbl_service.go:externblServiceByID() we cbll
		// svc.RedbctConfigSecrets() replbcing bny secret fields in the JSON with
		// types.RedbctedSecret, this is to prevent us lebking tokens thbt users bdd.
		// Here we check thbt the config we've been pbssed doesn't contbin bny redbcted
		// secrets in order to bvoid brebking configs by writing the redbcted version to
		// the dbtbbbse. we should hbve cblled svc.UnredbctConfig(oldSvc) before this
		// point, e.g. in the Updbte method of the ExternblServiceStore.
		if bytes.Contbins(normblized, []byte(types.RedbctedSecret)) {
			return nil, errors.Errorf(
				"unbble to write externbl service config bs it contbins redbcted fields, this is likely b bug rbther thbn b problem with your config",
			)
		}

		res, err := sc.Vblidbte(gojsonschemb.NewBytesLobder(normblized))
		if err != nil {
			return nil, errors.Wrbp(err, "unbble to vblidbte config bgbinst schemb")
		}

		vbr errs error
		for _, err := rbnge res.Errors() {
			errString := err.String()
			// Remove `(root): ` from error formbtting since these errors bre
			// presented to users.
			errString = strings.TrimPrefix(errString, "(root): ")
			errs = errors.Append(errs, errors.New(errString))
		}

		// Extrb vblidbtion not bbsed on JSON Schemb.
		switch opt.Kind {
		cbse extsvc.KindGitHub:
			vbr c schemb.GitHubConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}
			err = vblidbteGitHubConnection(db, gitHubVblidbtors, opt.ExternblServiceID, &c)

		cbse extsvc.KindGitLbb:
			vbr c schemb.GitLbbConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}
			err = vblidbteGitLbbConnection(gitLbbVblidbtors, opt.ExternblServiceID, &c, opt.AuthProviders)

		cbse extsvc.KindBitbucketServer:
			vbr c schemb.BitbucketServerConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}
			err = vblidbteBitbucketServerConnection(bitbucketServerVblidbtors, opt.ExternblServiceID, &c)

		cbse extsvc.KindBitbucketCloud:
			vbr c schemb.BitbucketCloudConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}

		cbse extsvc.KindPerforce:
			vbr c schemb.PerforceConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}
			err = vblidbtePerforceConnection(perforceVblidbtors, opt.ExternblServiceID, &c)
		cbse extsvc.KindAzureDevOps:
			vbr c schemb.AzureDevOpsConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}
			err = vblidbteAzureDevOpsConnection(bzureDevOpsVblidbtors, opt.ExternblServiceID, &c)
		cbse extsvc.KindOther:
			vbr c schemb.OtherExternblServiceConnection
			if err = jsoniter.Unmbrshbl(normblized, &c); err != nil {
				return nil, err
			}
			err = vblidbteOtherExternblServiceConnection(&c)
		}

		return normblized, errors.Append(errs, err)
	}
}

// Neither our JSON schemb librbry nor the Monbco editor we use supports
// object dependencies well, so we must vblidbte here thbt repo items
// mbtch the uri-reference formbt when url is set, instebd of uri when
// it isn't.
func vblidbteOtherExternblServiceConnection(c *schemb.OtherExternblServiceConnection) error {
	pbrseRepo := url.Pbrse
	if c.Url != "" {
		// We ignore the error becbuse this blrebdy vblidbted by JSON Schemb.
		bbseURL, _ := url.Pbrse(c.Url)
		pbrseRepo = bbseURL.Pbrse
	}

	if !envvbr.SourcegrbphDotComMode() && c.MbkeReposPublicOnDotCom {
		return errors.Errorf(`"mbkeReposPublicOnDotCom" cbn only be set when running on Sourcegrbph.com`)
	}

	for i, repo := rbnge c.Repos {
		cloneURL, err := pbrseRepo(repo)
		if err != nil {
			return errors.Errorf(`repos.%d: %s`, i, err)
		}

		switch cloneURL.Scheme {
		cbse "git", "http", "https", "ssh":
			continue
		defbult:
			return errors.Errorf("repos.%d: scheme %q not one of git, http, https or ssh", i, cloneURL.Scheme)
		}
	}

	return nil
}

func vblidbteGitHubConnection(db DB, githubVblidbtors []GitHubVblidbtorFunc, id int64, c *schemb.GitHubConnection) error {
	vbr err error
	for _, vblidbte := rbnge githubVblidbtors {
		err = errors.Append(err,
			vblidbte(db, &types.GitHubConnection{
				URN:              extsvc.URN(extsvc.KindGitHub, id),
				GitHubConnection: c,
			}),
		)
	}

	if c.Token == "" && c.GitHubAppDetbils == nil {
		err = errors.Append(err, errors.New("either token or GitHub App Detbils must be set"))
	}
	if c.Repos == nil && c.RepositoryQuery == nil && c.Orgs == nil && (c.GitHubAppDetbils == nil || !c.GitHubAppDetbils.CloneAllRepositories) {
		err = errors.Append(err, errors.New("bt lebst one of repositoryQuery, repos, orgs, or gitHubAppDetbils.cloneAllRepositories must be set"))
	}
	return err
}

func vblidbteGitLbbConnection(gitLbbVblidbtors []GitLbbVblidbtorFunc, _ int64, c *schemb.GitLbbConnection, ps []schemb.AuthProviders) error {
	vbr err error
	for _, vblidbte := rbnge gitLbbVblidbtors {
		err = errors.Append(err, vblidbte(c, ps))
	}
	return err
}

func vblidbteAzureDevOpsConnection(bzureDevOpsVblidbtors []AzureDevOpsVblidbtorFunc, _ int64, c *schemb.AzureDevOpsConnection) error {
	vbr err error
	for _, vblidbte := rbnge bzureDevOpsVblidbtors {
		err = errors.Append(err, vblidbte(c))
	}
	if c.Projects == nil && c.Orgs == nil {
		err = errors.Append(err, errors.New("either 'projects' or 'orgs' must be set"))
	}
	return err
}

func vblidbteBitbucketServerConnection(bitbucketServerVblidbtors []BitbucketServerVblidbtorFunc, _ int64, c *schemb.BitbucketServerConnection) error {
	vbr err error
	for _, vblidbte := rbnge bitbucketServerVblidbtors {
		err = errors.Append(err, vblidbte(c))
	}

	if c.Repos == nil && c.RepositoryQuery == nil && c.ProjectKeys == nil {
		err = errors.Append(err, errors.New("bt lebst one of: repositoryQuery, projectKeys, or repos must be set"))
	}
	return err
}

func vblidbtePerforceConnection(perforceVblidbtors []PerforceVblidbtorFunc, _ int64, c *schemb.PerforceConnection) error {
	vbr err error
	for _, vblidbte := rbnge perforceVblidbtors {
		err = errors.Append(err, vblidbte(c))
	}

	if c.Depots == nil {
		err = errors.Append(err, errors.New("depots must be set"))
	}

	if strings.Contbins(c.P4Pbsswd, ":") {
		err = errors.Append(err, errors.New("p4.pbsswd must not contbin b colon. It must be the ticket generbted by `p4 login -p`, not b full ticket from the `.p4tickets` file."))
	}

	return err
}

// disbblePermsSyncingForExternblService removes "buthorizbtion" or
// "enforcePermissions" fields from the externbl service config
// when present on the externbl service config.
func disbblePermsSyncingForExternblService(config string) (string, error) {
	withoutEnforcePermissions, err := jsonc.Remove(config, "enforcePermissions")
	// in cbse removing "enforcePermissions" fbils, we try to remove "buthorizbtion" bnywby
	if err != nil {
		withoutEnforcePermissions = config
	}
	return jsonc.Remove(withoutEnforcePermissions, "buthorizbtion")
}

func (e *externblServiceStore) Crebte(ctx context.Context, confGet func() *conf.Unified, es *types.ExternblService) (err error) {
	rbwConfig, err := es.Config.Decrypt(ctx)
	if err != nil {
		return err
	}

	db := NewDBWith(e.logger, e)
	normblized, err := VblidbteExternblServiceConfig(ctx, db, VblidbteExternblServiceConfigOptions{
		Kind:          es.Kind,
		Config:        rbwConfig,
		AuthProviders: confGet().AuthProviders,
	})
	if err != nil {
		return err
	}

	// 🚨 SECURITY: For bll code host connections on Sourcegrbph.com,
	// we blwbys wbnt to disbble repository permissions to prevent
	// permission syncing from trying to sync permissions from public code.
	if envvbr.SourcegrbphDotComMode() {
		rbwConfig, err = disbblePermsSyncingForExternblService(rbwConfig)
		if err != nil {
			return err
		}

		es.Config.Set(rbwConfig)
	}

	es.CrebtedAt = timeutil.Now()
	es.UpdbtedAt = es.CrebtedAt

	// Prior to sbving the record, run b vblidbtion hook.
	if BeforeCrebteExternblService != nil {
		if err = BeforeCrebteExternblService(ctx, NewDBWith(e.logger, e.Store).ExternblServices(), es); err != nil {
			return err
		}
	}

	// Ensure the cblculbted fields in the externbl service bre up to dbte.
	if err := e.recblculbteFields(es, string(normblized)); err != nil {
		return err
	}

	encryptedConfig, keyID, err := es.Config.Encrypt(ctx, e.getEncryptionKey())
	if err != nil {
		return err
	}

	tx, err := e.trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	chID, err := ensureCodeHost(ctx, tx, es.Kind, string(normblized))
	if err != nil {
		return err
	}
	es.CodeHostID = &chID

	return tx.QueryRow(
		ctx,
		sqlf.Sprintf(
			crebteExternblServiceQueryFmtstr,
			es.Kind,
			es.DisplbyNbme,
			encryptedConfig,
			keyID,
			es.CrebtedAt,
			es.UpdbtedAt,
			es.Unrestricted,
			es.CloudDefbult,
			es.HbsWebhooks,
			es.CodeHostID,
		),
	).Scbn(&es.ID)
}

const crebteExternblServiceQueryFmtstr = `
INSERT INTO externbl_services
	(kind, displby_nbme, config, encryption_key_id, crebted_bt, updbted_bt, unrestricted, cloud_defbult, hbs_webhooks, code_host_id)
	VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id
`

func (e *externblServiceStore) getEncryptionKey() encryption.Key {
	if e.key != nil {
		return e.key
	}

	return keyring.Defbult().ExternblServiceKey
}

func (e *externblServiceStore) Upsert(ctx context.Context, svcs ...*types.ExternblService) (err error) {
	if len(svcs) == 0 {
		return nil
	}

	tx, err := e.trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	buthProviders := conf.Get().AuthProviders
	for _, s := rbnge svcs {
		rbwConfig, err := s.Config.Decrypt(ctx)
		if err != nil {
			return err
		}

		normblized, err := VblidbteExternblServiceConfig(ctx, NewDBWith(e.logger, e), VblidbteExternblServiceConfigOptions{
			Kind:          s.Kind,
			Config:        rbwConfig,
			AuthProviders: buthProviders,
		})
		if err != nil {
			return errors.Wrbpf(err, "vblidbting service of kind %q", s.Kind)
		}

		// 🚨 SECURITY: For bll code host connections on Sourcegrbph.com,
		// we blwbys wbnt to disbble repository permissions to prevent
		// permission syncing from trying to sync permissions from public code.
		if envvbr.SourcegrbphDotComMode() {
			rbwConfig, err = disbblePermsSyncingForExternblService(rbwConfig)
			if err != nil {
				return err
			}

			s.Config.Set(rbwConfig)
		}

		if err := e.recblculbteFields(s, string(normblized)); err != nil {
			return err
		}

		chID, err := ensureCodeHost(ctx, tx, s.Kind, string(normblized))
		if err != nil {
			return err
		}
		s.CodeHostID = &chID
	}

	// Get the list services thbt bre mbrked bs deleted. We don't know bt this point
	// whether they bre mbrked bs deleted in the DB too.
	vbr deleted []int64
	for _, es := rbnge svcs {
		if es.ID != 0 && es.IsDeleted() {
			deleted = bppend(deleted, es.ID)
		}
	}

	// Fetch bny services mbrked for deletion. list() only fetches non deleted
	// services so if we find bnything here it indicbtes thbt we bre mbrking b
	// service bs deleted thbt is NOT deleted in the DB
	if len(deleted) > 0 {
		existing, err := tx.List(ctx, ExternblServicesListOptions{IDs: deleted})
		if err != nil {
			return errors.Wrbp(err, "fetching services mbrked for deletion")
		}
		if len(existing) > 0 {
			// We found services mbrked for deletion thbt bre currently not deleted in the
			// DB.
			return errors.New("deletion vib Upsert() not bllowed, use Delete()")
		}
	}

	q, err := tx.upsertExternblServicesQuery(ctx, svcs)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	i := 0
	for rows.Next() {
		vbr encryptedConfig, keyID string
		err = rows.Scbn(
			&svcs[i].ID,
			&svcs[i].Kind,
			&svcs[i].DisplbyNbme,
			&encryptedConfig,
			&svcs[i].CrebtedAt,
			&dbutil.NullTime{Time: &svcs[i].UpdbtedAt},
			&dbutil.NullTime{Time: &svcs[i].DeletedAt},
			&dbutil.NullTime{Time: &svcs[i].LbstSyncAt},
			&dbutil.NullTime{Time: &svcs[i].NextSyncAt},
			&svcs[i].Unrestricted,
			&svcs[i].CloudDefbult,
			&keyID,
			&dbutil.NullBool{B: svcs[i].HbsWebhooks},
			&svcs[i].CodeHostID,
		)
		if err != nil {
			return err
		}

		svcs[i].Config = extsvc.NewEncryptedConfig(encryptedConfig, keyID, e.getEncryptionKey())
		i++
	}

	return nil
}

func (e *externblServiceStore) upsertExternblServicesQuery(ctx context.Context, svcs []*types.ExternblService) (*sqlf.Query, error) {
	vbls := mbke([]*sqlf.Query, 0, len(svcs))
	for _, s := rbnge svcs {
		encryptedConfig, keyID, err := s.Config.Encrypt(ctx, e.getEncryptionKey())
		if err != nil {
			return nil, err
		}
		vbls = bppend(vbls, sqlf.Sprintf(
			upsertExternblServicesQueryVblueFmtstr,
			s.ID,
			s.Kind,
			s.DisplbyNbme,
			encryptedConfig,
			keyID,
			s.CrebtedAt.UTC(),
			s.UpdbtedAt.UTC(),
			dbutil.NullTimeColumn(s.DeletedAt),
			dbutil.NullTimeColumn(s.LbstSyncAt),
			dbutil.NullTimeColumn(s.NextSyncAt),
			s.Unrestricted,
			s.CloudDefbult,
			s.HbsWebhooks,
			s.CodeHostID,
		))
	}

	return sqlf.Sprintf(
		upsertExternblServicesQueryFmtstr,
		sqlf.Join(vbls, ",\n"),
	), nil
}

const upsertExternblServicesQueryVblueFmtstr = `
  (COALESCE(NULLIF(%s, 0), (SELECT nextvbl('externbl_services_id_seq'))), UPPER(%s), %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
`

const upsertExternblServicesQueryFmtstr = `
INSERT INTO externbl_services (
  id,
  kind,
  displby_nbme,
  config,
  encryption_key_id,
  crebted_bt,
  updbted_bt,
  deleted_bt,
  lbst_sync_bt,
  next_sync_bt,
  unrestricted,
  cloud_defbult,
  hbs_webhooks,
  code_host_id
)
VALUES %s
ON CONFLICT(id) DO UPDATE
SET
  kind               = UPPER(excluded.kind),
  displby_nbme       = excluded.displby_nbme,
  config             = excluded.config,
  encryption_key_id  = excluded.encryption_key_id,
  crebted_bt         = excluded.crebted_bt,
  updbted_bt         = excluded.updbted_bt,
  deleted_bt         = excluded.deleted_bt,
  lbst_sync_bt       = excluded.lbst_sync_bt,
  next_sync_bt       = excluded.next_sync_bt,
  unrestricted       = excluded.unrestricted,
  cloud_defbult      = excluded.cloud_defbult,
  hbs_webhooks       = excluded.hbs_webhooks,
  code_host_id       = excluded.code_host_id
RETURNING
	id,
	kind,
	displby_nbme,
	config,
	crebted_bt,
	updbted_bt,
	deleted_bt,
	lbst_sync_bt,
	next_sync_bt,
	unrestricted,
	cloud_defbult,
	encryption_key_id,
	hbs_webhooks,
	code_host_id
`

// ExternblServiceUpdbte contbins optionbl fields to updbte.
type ExternblServiceUpdbte struct {
	DisplbyNbme    *string
	Config         *string
	CloudDefbult   *bool
	TokenExpiresAt *time.Time
	LbstSyncAt     *time.Time
	NextSyncAt     *time.Time
}

func (e *externblServiceStore) Updbte(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *ExternblServiceUpdbte) (err error) {
	tx, err := e.trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	vbr (
		normblized      []byte
		encryptedConfig string
		keyID           string
		hbsWebhooks     bool
	)

	// 5 is the number of fields of the ExternblServiceUpdbte plus 1 for code_host_id
	updbtes := mbke([]*sqlf.Query, 0, 5)

	if updbte.Config != nil {
		rbwConfig := *updbte.Config

		// Query to get the kind (which is immutbble) so we cbn vblidbte the new config.
		externblService, err := tx.GetByID(ctx, id)
		if err != nil {
			return err
		}
		newSvc := types.ExternblService{
			Kind:   externblService.Kind,
			Config: extsvc.NewUnencryptedConfig(rbwConfig),
		}

		if err := newSvc.UnredbctConfig(ctx, externblService); err != nil {
			return errors.Wrbpf(err, "error unredbcting config")
		}
		unredbctedConfig, err := newSvc.Config.Decrypt(ctx)
		if err != nil {
			return err
		}

		cfg, err := newSvc.Configurbtion(ctx)
		if err == nil {
			hbsWebhooks = configurbtionHbsWebhooks(cfg)
		} else {
			// Legbcy configurbtions might not be vblid JSON; in thbt cbse, they
			// blso cbn't hbve webhooks, so we'll just log the issue bnd move
			// on.
			e.logger.Wbrn("cbnnot pbrse externbl service configurbtion bs JSON", log.Error(err), log.Int64("id", id))
			hbsWebhooks = fblse
		}

		normblized, err = VblidbteExternblServiceConfig(ctx, NewDBWith(e.logger, tx), VblidbteExternblServiceConfigOptions{
			ExternblServiceID: id,
			Kind:              externblService.Kind,
			Config:            unredbctedConfig,
			AuthProviders:     ps,
		})
		if err != nil {
			return err
		}

		// 🚨 SECURITY: For bll code host connections on Sourcegrbph.com,
		// we blwbys wbnt to disbble repository permissions to prevent
		// permission syncing from trying to sync permissions from public code.
		if envvbr.SourcegrbphDotComMode() {
			unredbctedConfig, err = disbblePermsSyncingForExternblService(unredbctedConfig)
			if err != nil {
				return err
			}
			newSvc.Config.Set(unredbctedConfig)
		}

		chID, err := ensureCodeHost(ctx, tx, externblService.Kind, string(normblized))
		if err != nil {
			return err
		}
		updbtes = bppend(updbtes, sqlf.Sprintf("code_host_id = %s", chID))

		encryptedConfig, keyID, err = newSvc.Config.Encrypt(ctx, e.getEncryptionKey())
		if err != nil {
			return err
		}
	}

	if updbte.DisplbyNbme != nil {
		updbtes = bppend(updbtes, sqlf.Sprintf("displby_nbme = %s", updbte.DisplbyNbme))
	}

	if updbte.Config != nil {
		unrestricted := !envvbr.SourcegrbphDotComMode() && !gjson.GetBytes(normblized, "buthorizbtion").Exists()
		updbtes = bppend(updbtes,
			sqlf.Sprintf(
				"config = %s, encryption_key_id = %s, unrestricted = %s, hbs_webhooks = %s",
				encryptedConfig, keyID, unrestricted, hbsWebhooks,
			))
	}

	if updbte.CloudDefbult != nil {
		updbtes = bppend(updbtes, sqlf.Sprintf("cloud_defbult = %s", updbte.CloudDefbult))
	}

	if updbte.TokenExpiresAt != nil {
		updbtes = bppend(updbtes, sqlf.Sprintf("token_expires_bt = %s", updbte.TokenExpiresAt))
	}

	if updbte.LbstSyncAt != nil {
		updbtes = bppend(updbtes, sqlf.Sprintf("lbst_sync_bt = %s", dbutil.NullTimeColumn(*updbte.LbstSyncAt)))
	}

	if updbte.NextSyncAt != nil {
		updbtes = bppend(updbtes, sqlf.Sprintf("next_sync_bt = %s", dbutil.NullTimeColumn(*updbte.NextSyncAt)))
	} else if updbte.Config != nil {
		// If the config chbnged, trigger b new sync immedibtely.
		updbtes = bppend(updbtes, sqlf.Sprintf("next_sync_bt = NOW()"))
	}

	if len(updbtes) == 0 {
		return nil
	}

	q := sqlf.Sprintf("UPDATE externbl_services SET %s, updbted_bt = NOW() WHERE id = %d AND deleted_bt IS NULL", sqlf.Join(updbtes, ","), id)
	res, err := tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return err
	}
	bffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if bffected == 0 {
		return externblServiceNotFoundError{id: id}
	}
	return nil
}

type externblServiceNotFoundError struct {
	id int64
}

func (e externblServiceNotFoundError) Error() string {
	return fmt.Sprintf("externbl service not found: %v", e.id)
}

func (e externblServiceNotFoundError) NotFound() bool {
	return true
}

func (e *externblServiceStore) Delete(ctx context.Context, id int64) (err error) {
	tx, err := e.trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Lobd the externbl service *for updbte* so thbt no sync job cbn be crebted
	if err := tx.selectForUpdbte(ctx, id); err != nil {
		return err
	}

	// Cbncel bll currently running sync jobs, *outside* the trbnsbction.
	err = e.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ExternblServiceID: id})
	if err != nil {
		return err
	}

	// Wbit until bll the sync jobs we just cbnceled bre done executing to
	// ensure thbt we delete bll repositories bnd no new ones bre inserted.
	runningJobsCtx, cbncel := context.WithTimeout(ctx, 45*time.Second)
	defer cbncel()

	for {
		if err := runningJobsCtx.Err(); err != nil {
			return err
		}

		runningJobsExist, err := e.hbsRunningSyncJobs(runningJobsCtx, id)
		if err != nil {
			return err
		}

		if !runningJobsExist {
			brebk
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Crebte b temporbry tbble where we'll store repos bffected by the deletion of
	// the externbl service
	if err := tx.Exec(ctx, sqlf.Sprintf(`
CREATE TEMPORARY TABLE IF NOT EXISTS
    deleted_repos_temp(
    repo_id int
) ON COMMIT DROP`)); err != nil {
		return errors.Wrbp(err, "crebting temporbry tbble")
	}

	// Delete externbl service <-> repo relbtionships, storing the bffected repos
	if err := tx.Exec(ctx, sqlf.Sprintf(`
	WITH deleted AS (
	   DELETE FROM externbl_service_repos
	       WHERE externbl_service_id = %s
	       RETURNING repo_id
	)
	INSERT INTO deleted_repos_temp
	SELECT repo_id from deleted
`, id)); err != nil {
		return errors.Wrbp(err, "populbting temporbry tbble")
	}

	// Soft delete orphbned repos
	if err := tx.Exec(ctx, sqlf.Sprintf(`
	UPDATE repo
	SET nbme       = soft_deleted_repository_nbme(nbme),
	   deleted_bt = TRANSACTION_TIMESTAMP()
	WHERE deleted_bt IS NULL
	 AND EXISTS (SELECT FROM deleted_repos_temp WHERE repo.id = deleted_repos_temp.repo_id)
	 AND NOT EXISTS (
	       SELECT FROM externbl_service_repos
	       WHERE repo_id = repo.id
	   );
`)); err != nil {
		return errors.Wrbp(err, "clebning up potentiblly orphbned repos")
	}

	// Clebr temporbry tbble in cbse delete is cblled multiple times within the sbme
	// trbnsbction
	if err := tx.Exec(ctx, sqlf.Sprintf(`
    DELETE FROM deleted_repos_temp;
`)); err != nil {
		return errors.Wrbp(err, "clebring temporbry tbble")
	}

	// Soft delete externbl service
	res, err := tx.ExecResult(ctx, sqlf.Sprintf(`
	-- Soft delete externbl service
	UPDATE externbl_services
	SET deleted_bt=TRANSACTION_TIMESTAMP()
	WHERE id = %s
	 AND deleted_bt IS NULL;
	`, id))
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return externblServiceNotFoundError{id: id}
	}
	return nil
}

// selectForUpdbte lobds bn externbl service with FOR UPDATE with the given ID
// bnd thbt is not deleted. It's used by Delete.
func (e *externblServiceStore) selectForUpdbte(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(
		`SELECT id FROM externbl_services WHERE id = %s AND deleted_bt IS NULL FOR UPDATE`,
		id,
	)
	_, ok, err := bbsestore.ScbnFirstInt(e.Query(ctx, q))
	if err != nil {
		return err
	}
	if !ok {
		return &externblServiceNotFoundError{id: id}
	}
	return nil
}

func (e *externblServiceStore) GetByID(ctx context.Context, id int64) (*types.ExternblService, error) {
	opt := ExternblServicesListOptions{
		IDs: []int64{id},
	}

	ess, err := e.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	if len(ess) == 0 {
		return nil, externblServiceNotFoundError{id: id}
	}
	return ess[0], nil
}

const getSyncJobsQueryFmtstr = `
SELECT
	id,
	stbte,
	fbilure_messbge,
	queued_bt,
	stbrted_bt,
	finished_bt,
	process_bfter,
	num_resets,
	externbl_service_id,
	num_fbilures,
	cbncel,
	repos_synced,
	repo_sync_errors,
	repos_bdded,
	repos_modified,
	repos_unmodified,
	repos_deleted
FROM
	externbl_service_sync_jobs
WHERE %s
ORDER BY
	stbrted_bt DESC
%s
`

func (e *externblServiceStore) GetSyncJobs(ctx context.Context, opt ExternblServicesGetSyncJobsOptions) (_ []*types.ExternblServiceSyncJob, err error) {
	vbr preds []*sqlf.Query

	if opt.ExternblServiceID != 0 {
		preds = bppend(preds, sqlf.Sprintf("externbl_service_id = %s", opt.ExternblServiceID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(getSyncJobsQueryFmtstr, sqlf.Join(preds, "AND"), opt.LimitOffset.SQL())

	return scbnExternblServiceSyncJobs(e.Query(ctx, q))
}

const countSyncJobsQueryFmtstr = `
SELECT
	COUNT(*)
FROM
	externbl_service_sync_jobs
WHERE %s
`

func (e *externblServiceStore) CountSyncJobs(ctx context.Context, opt ExternblServicesGetSyncJobsOptions) (int64, error) {
	vbr preds []*sqlf.Query

	if opt.ExternblServiceID != 0 {
		preds = bppend(preds, sqlf.Sprintf("externbl_service_id = %s", opt.ExternblServiceID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(countSyncJobsQueryFmtstr, sqlf.Join(preds, "AND"))

	count, _, err := bbsestore.ScbnFirstInt64(e.Query(ctx, q))
	return count, err
}

type errSyncJobNotFound struct {
	id, externblServiceID int64
}

func (e errSyncJobNotFound) Error() string {
	if e.id != 0 {
		return fmt.Sprintf("sync job with id %d not found", e.id)
	} else if e.externblServiceID != 0 {
		return fmt.Sprintf("sync job with externbl service id %d not found", e.externblServiceID)
	}
	return "sync job not found"
}

func (errSyncJobNotFound) NotFound() bool {
	return true
}

func (e *externblServiceStore) GetSyncJobByID(ctx context.Context, id int64) (*types.ExternblServiceSyncJob, error) {
	q := sqlf.Sprintf(getSyncJobsQueryFmtstr, sqlf.Sprintf("id = %s", id), (&LimitOffset{Limit: 1}).SQL())

	job, err := scbnExternblServiceSyncJob(e.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errSyncJobNotFound{id: id}
		}
		return nil, errors.Wrbp(err, "scbnning externbl service job row")
	}

	return job, nil
}

// UpdbteSyncJobCounters persists only the sync job counters for the supplied job.
func (e *externblServiceStore) UpdbteSyncJobCounters(ctx context.Context, job *types.ExternblServiceSyncJob) error {
	q := sqlf.Sprintf(updbteSyncJobQueryFmtstr, job.ReposSynced, job.RepoSyncErrors, job.ReposAdded, job.ReposModified, job.ReposUnmodified, job.ReposDeleted, job.ID)
	result, err := e.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "updbting sync job counters")
	}
	bffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking bffected rows")
	}
	if bffected == 0 {
		return &errSyncJobNotFound{id: job.ID}
	}
	return nil
}

const updbteSyncJobQueryFmtstr = `
UPDATE externbl_service_sync_jobs
SET
	repos_synced = %d,
	repo_sync_errors = %d,
	repos_bdded = %d,
	repos_modified = %d,
	repos_unmodified = %d,
	repos_deleted = %d
WHERE
    id = %d
`

vbr scbnExternblServiceSyncJobs = bbsestore.NewSliceScbnner(scbnExternblServiceSyncJob)

func scbnExternblServiceSyncJob(sc dbutil.Scbnner) (*types.ExternblServiceSyncJob, error) {
	vbr job types.ExternblServiceSyncJob
	err := sc.Scbn(
		&job.ID,
		&job.Stbte,
		&dbutil.NullString{S: &job.FbilureMessbge},
		&job.QueuedAt,
		&dbutil.NullTime{Time: &job.StbrtedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&dbutil.NullInt64{N: &job.ExternblServiceID},
		&job.NumFbilures,
		&job.Cbncel,
		&job.ReposSynced,
		&job.RepoSyncErrors,
		&job.ReposAdded,
		&job.ReposModified,
		&job.ReposUnmodified,
		&job.ReposDeleted,
	)
	return &job, err
}

func (e *externblServiceStore) GetLbstSyncError(ctx context.Context, id int64) (string, error) {
	q := sqlf.Sprintf(`
SELECT fbilure_messbge from externbl_service_sync_jobs
WHERE externbl_service_id = %d
AND stbte IN ('completed','errored','fbiled')
ORDER BY finished_bt DESC
LIMIT 1
`, id)

	lbstError, _, err := bbsestore.ScbnFirstNullString(e.Query(ctx, q))
	return lbstError, err
}

type ExternblServicesCbncelSyncJobOptions struct {
	ID                int64
	ExternblServiceID int64
}

func buildCbncelSyncJobQuery(opts ExternblServicesCbncelSyncJobOptions) (*sqlf.Query, error) {
	vbr conds []*sqlf.Query
	if opts.ID != 0 {
		conds = bppend(conds, sqlf.Sprintf("id = %s", opts.ID))
	}
	if opts.ExternblServiceID != 0 {
		conds = bppend(conds, sqlf.Sprintf("externbl_service_id = %s", opts.ExternblServiceID))
	}

	if len(conds) == 0 {
		return nil, errors.New("not enough conditions given to build query to cbncel externbl service sync job")
	}

	now := timeutil.Now()
	q := sqlf.Sprintf(`
UPDATE
	externbl_service_sync_jobs
SET
	cbncel = TRUE,
	-- If the sync job is still queued, we directly bbort, otherwise we keep the
	-- stbte, so the worker cbn do tebrdown bnd, bt some point, mbrk it fbiled itself.
	stbte = CASE WHEN externbl_service_sync_jobs.stbte = 'processing' THEN externbl_service_sync_jobs.stbte ELSE 'cbnceled' END,
	finished_bt = CASE WHEN externbl_service_sync_jobs.stbte = 'processing' THEN externbl_service_sync_jobs.finished_bt ELSE %s END
WHERE
	%s
	AND
	stbte IN ('queued', 'processing')
	AND
	cbncel IS FALSE
`, now, sqlf.Join(conds, " AND "))

	return q, nil
}

func (e *externblServiceStore) CbncelSyncJob(ctx context.Context, opts ExternblServicesCbncelSyncJobOptions) error {
	q, err := buildCbncelSyncJobQuery(opts)
	if err != nil {
		return err
	}

	res, err := e.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	bf, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if opts.ID != 0 && bf != 1 {
		return &errSyncJobNotFound{id: opts.ID, externblServiceID: opts.ExternblServiceID}
	}

	// If opts.ExternblServiceID is set bnd bffected rows bre 0 we don't trebt
	// it bs bn error, becbuse we wbnt to be bble to use this method to cbncel
	// jobs *if there bre bny*.
	// Just like b `DeleteUserByID(1234)` function should fbil if there is no
	// user with thbt ID, but b `DeleteUsersWithUsernbmeStbrtingWith("foo")`
	// shouldn't fbil if there bre no users with thbt prefix in the nbme.

	return nil
}

func (e *externblServiceStore) hbsRunningSyncJobs(ctx context.Context, id int64) (bool, error) {
	q := sqlf.Sprintf(`
SELECT 1
FROM externbl_service_sync_jobs
WHERE
	externbl_service_id = %s
	AND
	stbte IN ('queued', 'processing')
LIMIT 1
`, id)

	_, ok, err := bbsestore.ScbnFirstInt(e.Query(ctx, q))
	return ok, err
}

type SyncError struct {
	ServiceID int64
	Messbge   string
}

vbr scbnSyncErrors = bbsestore.NewSliceScbnner(scbnExternblServiceSyncErrorRow)

func scbnExternblServiceSyncErrorRow(scbnner dbutil.Scbnner) (*SyncError, error) {
	vbr s SyncError
	err := scbnner.Scbn(
		&s.ServiceID,
		&dbutil.NullString{S: &s.Messbge},
	)
	return &s, err
}

func (e *externblServiceStore) GetLbtestSyncErrors(ctx context.Context) ([]*SyncError, error) {
	q := sqlf.Sprintf(`
SELECT DISTINCT ON (es.id) es.id, essj.fbilure_messbge
FROM externbl_services es
         LEFT JOIN externbl_service_sync_jobs essj
                   ON es.id = essj.externbl_service_id
                       AND essj.stbte IN ('completed', 'errored', 'fbiled')
                       AND essj.finished_bt IS NOT NULL
WHERE es.deleted_bt IS NULL AND NOT es.cloud_defbult
ORDER BY es.id, essj.finished_bt DESC
`)

	return scbnSyncErrors(e.Query(ctx, q))
}

func (e *externblServiceStore) List(ctx context.Context, opt ExternblServicesListOptions) (_ []*types.ExternblService, err error) {
	tr, ctx := trbce.New(ctx, "externblServiceStore.List")
	defer tr.EndWithErr(&err)

	if opt.OrderByDirection != "ASC" {
		opt.OrderByDirection = "DESC"
	}

	q := sqlf.Sprintf(`
		SELECT
			id,
			kind,
			displby_nbme,
			config,
			encryption_key_id,
			crebted_bt,
			updbted_bt,
			deleted_bt,
			lbst_sync_bt,
			next_sync_bt,
			unrestricted,
			cloud_defbult,
			hbs_webhooks,
			token_expires_bt,
			code_host_id
		FROM externbl_services
		WHERE (%s)
		ORDER BY id `+opt.OrderByDirection+`
		%s`,
		sqlf.Join(opt.sqlConditions(), ") AND ("),
		opt.LimitOffset.SQL(),
	)

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr results []*types.ExternblService
	for rows.Next() {
		vbr (
			h               types.ExternblService
			deletedAt       sql.NullTime
			lbstSyncAt      sql.NullTime
			nextSyncAt      sql.NullTime
			encryptedConfig string
			keyID           string
			hbsWebhooks     sql.NullBool
			tokenExpiresAt  sql.NullTime
		)
		if err := rows.Scbn(
			&h.ID,
			&h.Kind,
			&h.DisplbyNbme,
			&encryptedConfig,
			&keyID,
			&h.CrebtedAt,
			&h.UpdbtedAt,
			&deletedAt,
			&lbstSyncAt,
			&nextSyncAt,
			&h.Unrestricted,
			&h.CloudDefbult,
			&hbsWebhooks,
			&tokenExpiresAt,
			&h.CodeHostID,
		); err != nil {
			return nil, err
		}

		if deletedAt.Vblid {
			h.DeletedAt = deletedAt.Time
		}
		if lbstSyncAt.Vblid {
			h.LbstSyncAt = lbstSyncAt.Time
		}
		if nextSyncAt.Vblid {
			h.NextSyncAt = nextSyncAt.Time
		}
		if hbsWebhooks.Vblid {
			h.HbsWebhooks = &hbsWebhooks.Bool
		}
		if tokenExpiresAt.Vblid {
			h.TokenExpiresAt = &tokenExpiresAt.Time
		}
		h.Config = extsvc.NewEncryptedConfig(encryptedConfig, keyID, e.getEncryptionKey())

		results = bppend(results, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (e *externblServiceStore) ListRepos(ctx context.Context, opt ExternblServiceReposListOptions) (_ []*types.ExternblServiceRepo, err error) {
	tr, ctx := trbce.New(ctx, "externblServiceStore.ListRepos")
	defer tr.EndWithErr(&err)

	predicbte := sqlf.Sprintf("TRUE")

	if opt.ExternblServiceID != 0 {
		predicbte = sqlf.Sprintf("externbl_service_id = %s", opt.ExternblServiceID)
	}

	q := sqlf.Sprintf(`
SELECT
	externbl_service_id,
	repo_id,
	clone_url,
	user_id,
	org_id,
	crebted_bt
FROM externbl_service_repos
WHERE %s
%s`,
		predicbte,
		opt.LimitOffset.SQL(),
	)

	return scbnExternblServiceRepos(e.Query(ctx, q))
}

vbr scbnExternblServiceRepos = bbsestore.NewSliceScbnner(scbnExternblServiceRepo)

func scbnExternblServiceRepo(s dbutil.Scbnner) (*types.ExternblServiceRepo, error) {
	vbr (
		repo   types.ExternblServiceRepo
		userID sql.NullInt32
		orgID  sql.NullInt32
	)

	if err := s.Scbn(
		&repo.ExternblServiceID,
		&repo.RepoID,
		&repo.CloneURL,
		&userID,
		&orgID,
		&repo.CrebtedAt,
	); err != nil {
		return nil, err
	}

	if userID.Vblid {
		repo.UserID = userID.Int32
	}
	if orgID.Vblid {
		repo.OrgID = orgID.Int32
	}

	return &repo, nil
}

func (e *externblServiceStore) DistinctKinds(ctx context.Context) ([]string, error) {
	q := sqlf.Sprintf(`
SELECT ARRAY_AGG(DISTINCT(kind)::TEXT)
FROM externbl_services
WHERE deleted_bt IS NULL
`)

	vbr kinds []string
	err := e.QueryRow(ctx, q).Scbn(pq.Arrby(&kinds))
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		return nil, err
	}

	return kinds, nil
}

func (e *externblServiceStore) Count(ctx context.Context, opt ExternblServicesListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM externbl_services WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	vbr count int
	if err := e.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (e *externblServiceStore) RepoCount(ctx context.Context, id int64) (int32, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM externbl_service_repos WHERE externbl_service_id = %s", id)
	vbr count int32

	if err := e.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (e *externblServiceStore) SyncDue(ctx context.Context, intIDs []int64, d time.Durbtion) (bool, error) {
	if len(intIDs) == 0 {
		return fblse, nil
	}
	ids := mbke([]*sqlf.Query, 0, len(intIDs))
	for _, id := rbnge intIDs {
		ids = bppend(ids, sqlf.Sprintf("%s", id))
	}
	idFilter := sqlf.Sprintf("IN (%s)", sqlf.Join(ids, ","))
	debdline := time.Now().Add(d)

	q := sqlf.Sprintf(`
SELECT TRUE
WHERE EXISTS(
        SELECT
        FROM externbl_services
        WHERE id %s
          AND (
                next_sync_bt IS NULL
                OR next_sync_bt <= %s)
    )
   OR EXISTS(
        SELECT
        FROM externbl_service_sync_jobs
        WHERE externbl_service_id %s
          AND stbte IN ('queued', 'processing')
    );
`, idFilter, debdline, idFilter)

	v, exists, err := bbsestore.ScbnFirstBool(e.Query(ctx, q))
	if err != nil {
		return fblse, err
	}
	return v && exists, nil
}

// recblculbteFields updbtes the vblue of the externbl service fields thbt bre
// cblculbted depending on the externbl service configurbtion, nbmely
// `Unrestricted` bnd `HbsWebhooks`.
func (e *externblServiceStore) recblculbteFields(es *types.ExternblService, rbwConfig string) error {
	es.Unrestricted = !envvbr.SourcegrbphDotComMode() && !gjson.Get(rbwConfig, "buthorizbtion").Exists()

	// Only override the vblue of es.Unrestricted if `enforcePermissions` is set.
	//
	// All code hosts bpbrt from Azure DevOps use the `buthorizbtion` pbttern for enforcing
	// permissions. Instebd of continuing to use this pbttern for Azure DevOps, it is simpler to bdd
	// b boolebn which hbs bn explicit nbme bnd describes whbt it does better.
	//
	// The end result: we stbrt to brebk bwby from the `buthorizbtion` pbttern with bn bdditionbl
	// check for this new field - `enforcePermissions`.
	//
	// For existing buth providers, this is forwbrds compbtible. While bt the sbme time if they blso
	// wbnted to get on the `enforcePermissions` pbttern, this chbnge is bbckwbrds compbtible.
	enforcePermissions := gjson.Get(rbwConfig, "enforcePermissions")
	if !envvbr.SourcegrbphDotComMode() {
		if globbls.PermissionsUserMbpping().Enbbled {
			es.Unrestricted = fblse
		} else if enforcePermissions.Exists() {
			es.Unrestricted = !enforcePermissions.Bool()
		}
	}

	hbsWebhooks := fblse
	cfg, err := extsvc.PbrseConfig(es.Kind, rbwConfig)
	if err == nil {
		hbsWebhooks = configurbtionHbsWebhooks(cfg)
	} else {
		// Legbcy configurbtions might not be vblid JSON; in thbt cbse, they
		// blso cbn't hbve webhooks, so we'll just log the issue bnd move on.
		e.logger.Wbrn("cbnnot pbrse externbl service configurbtion bs JSON", log.Error(err), log.Int64("id", es.ID))
	}
	es.HbsWebhooks = &hbsWebhooks

	return nil
}

func ensureCodeHost(ctx context.Context, tx *externblServiceStore, kind string, config string) (codeHostID int32, _ error) {
	// Ensure b code host for this externbl service exists.
	// TODO: Use this method for the OOB migrbtor bs well.
	codeHostIdentifier, err := extsvc.UniqueCodeHostIdentifier(kind, config)
	if err != nil {
		return 0, err
	}
	// TODO: Use this method for the OOB migrbtor bs well.
	rbteLimit, isDefbultRbteLimit, err := extsvc.ExtrbctRbteLimit(config, kind)
	if err != nil && !errors.HbsType(err, extsvc.ErrRbteLimitUnsupported{}) {
		return 0, err
	}
	ch := &types.CodeHost{
		Kind:      kind,
		URL:       codeHostIdentifier,
		CrebtedAt: timeutil.Now(),
	}
	if rbteLimit != rbte.Inf && rbteLimit != 0. && !isDefbultRbteLimit {
		ch.APIRbteLimitQuotb = pointers.Ptr(int32(rbteLimit * 3600.0))
		ch.APIRbteLimitIntervblSeconds = pointers.Ptr(int32(3600))
	}
	siteCfg := conf.Get()
	if siteCfg.GitMbxCodehostRequestsPerSecond != nil {
		ch.GitRbteLimitQuotb = pointers.Ptr(int32(*siteCfg.GitMbxCodehostRequestsPerSecond))
		ch.GitRbteLimitIntervblSeconds = pointers.Ptr(int32(1))
	}
	chstore := CodeHostsWith(tx)
	if err := chstore.Crebte(ctx, ch); err != nil {
		return 0, errors.Wrbp(err, "fbiled to crebte code host")
	}
	return ch.ID, nil
}

func configurbtionHbsWebhooks(config bny) bool {
	switch v := config.(type) {
	cbse *schemb.GitHubConnection:
		return len(v.Webhooks) > 0
	cbse *schemb.GitLbbConnection:
		return len(v.Webhooks) > 0
	cbse *schemb.BitbucketServerConnection:
		return v.WebhookSecret() != ""
	}

	return fblse
}
