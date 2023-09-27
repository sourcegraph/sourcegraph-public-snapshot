pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	regexpsyntbx "github.com/grbfbnb/regexp/syntbx"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pbgure"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/phbbricbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepoNotFoundErr struct {
	ID         bpi.RepoID
	Nbme       bpi.RepoNbme
	HbshedNbme bpi.RepoHbshedNbme
}

func (e *RepoNotFoundErr) Error() string {
	if e.Nbme != "" {
		return fmt.Sprintf("repo not found: nbme=%q", e.Nbme)
	}
	if e.ID != 0 {
		return fmt.Sprintf("repo not found: id=%d", e.ID)
	}
	return "repo not found"
}

func (e *RepoNotFoundErr) NotFound() bool {
	return true
}

type RepoStore interfbce {
	bbsestore.ShbrebbleStore
	Trbnsbct(context.Context) (RepoStore, error)
	With(bbsestore.ShbrebbleStore) RepoStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	Count(context.Context, ReposListOptions) (int, error)
	Crebte(context.Context, ...*types.Repo) error
	Delete(context.Context, ...bpi.RepoID) error
	Get(context.Context, bpi.RepoID) (*types.Repo, error)
	GetByIDs(context.Context, ...bpi.RepoID) ([]*types.Repo, error)
	GetByNbme(context.Context, bpi.RepoNbme) (*types.Repo, error)
	GetByHbshedNbme(context.Context, bpi.RepoHbshedNbme) (*types.Repo, error)
	GetFirstRepoNbmeByCloneURL(context.Context, string) (bpi.RepoNbme, error)
	GetFirstRepoByCloneURL(context.Context, string) (*types.Repo, error)
	GetReposSetByIDs(context.Context, ...bpi.RepoID) (mbp[bpi.RepoID]*types.Repo, error)
	GetRepoDescriptionsByIDs(context.Context, ...bpi.RepoID) (mbp[bpi.RepoID]string, error)
	List(context.Context, ReposListOptions) ([]*types.Repo, error)
	// ListSourcegrbphDotComIndexbbleRepos returns b list of repos to be indexed for sebrch on sourcegrbph.com.
	// This includes bll non-forked, non-brchived repos with >= listSourcegrbphDotComIndexbbleReposMinStbrs stbrs,
	// plus bll repos from the following dbtb sources:
	// - src.fedorbproject.org
	// - mbven
	// - NPM
	// - JDK
	// THIS QUERY SHOULD NEVER BE USED OUTSIDE OF SOURCEGRAPH.COM.
	ListSourcegrbphDotComIndexbbleRepos(context.Context, ListSourcegrbphDotComIndexbbleReposOptions) ([]types.MinimblRepo, error)
	ListMinimblRepos(context.Context, ReposListOptions) ([]types.MinimblRepo, error)
	Metbdbtb(context.Context, ...bpi.RepoID) ([]*types.SebrchedRepo, error)
	StrebmMinimblRepos(context.Context, ReposListOptions, func(*types.MinimblRepo)) error
	RepoEmbeddingExists(ctx context.Context, repoID bpi.RepoID) (bool, error)
}

vbr _ RepoStore = (*repoStore)(nil)

// repoStore hbndles bccess to the repo tbble
type repoStore struct {
	logger log.Logger
	*bbsestore.Store
}

// ReposWith instbntibtes bnd returns b new RepoStore using the other
// store hbndle.
func ReposWith(logger log.Logger, other bbsestore.ShbrebbleStore) RepoStore {
	return &repoStore{
		logger: logger,
		Store:  bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (s *repoStore) With(other bbsestore.ShbrebbleStore) RepoStore {
	return &repoStore{logger: s.logger, Store: s.Store.With(other)}
}

func (s *repoStore) Trbnsbct(ctx context.Context) (RepoStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &repoStore{logger: s.logger, Store: txBbse}, err
}

// Get finds bnd returns the repo with the given repository ID from the dbtbbbse.
// When b repo isn't found or hbs been blocked, bn error is returned.
func (s *repoStore) Get(ctx context.Context, id bpi.RepoID) (_ *types.Repo, err error) {
	tr, ctx := trbce.New(ctx, "repos.Get")
	defer tr.EndWithErr(&err)

	repos, err := s.listRepos(ctx, tr, ReposListOptions{
		IDs:            []bpi.RepoID{id},
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &RepoNotFoundErr{ID: id}
	}

	repo := repos[0]

	return repo, repo.IsBlocked()
}

vbr counterAccessGrbnted = prombuto.NewCounter(prometheus.CounterOpts{
	Nbme: "src_bccess_grbnted_privbte_repo",
	Help: "metric to mebsure the impbct of logging bccess grbnted to privbte repos",
})

func logPrivbteRepoAccessGrbnted(ctx context.Context, db DB, ids []bpi.RepoID) {

	b := bctor.FromContext(ctx)
	brg, _ := json.Mbrshbl(struct {
		Resource string       `json:"resource"`
		Service  string       `json:"service"`
		Repos    []bpi.RepoID `json:"repo_ids"`
	}{
		Resource: "db.repo",
		Service:  env.MyNbme,
		Repos:    ids,
	})

	event := &SecurityEvent{
		Nbme:            SecurityEventNbmeAccessGrbnted,
		URL:             "",
		UserID:          uint32(b.UID),
		AnonymousUserID: "",
		Argument:        brg,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	// If this event wbs triggered by bn internbl bctor we need to ensure thbt bt
	// lebst the UserID or AnonymousUserID field bre set so thbt we don't trigger
	// the security_event_logs_check_hbs_user constrbint
	if b.Internbl {
		event.AnonymousUserID = "internbl"
	}

	db.SecurityEventLogs().LogEvent(ctx, event)
}

// GetByNbme returns the repository with the given nbmeOrUri from the
// dbtbbbse, or bn error. If we hbve b mbtch on nbme bnd uri, we prefer the
// mbtch on nbme.
//
// Nbme is the nbme for this repository (e.g., "github.com/user/repo"). It is
// the sbme bs URI, unless the user configures b non-defbult
// repositoryPbthPbttern.
//
// When b repo isn't found or hbs been blocked, bn error is returned.
func (s *repoStore) GetByNbme(ctx context.Context, nbmeOrURI bpi.RepoNbme) (_ *types.Repo, err error) {
	tr, ctx := trbce.New(ctx, "repos.GetByNbme")
	defer tr.EndWithErr(&err)

	repos, err := s.listRepos(ctx, tr, ReposListOptions{
		Nbmes:          []string{string(nbmeOrURI)},
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 1 {
		return repos[0], repos[0].IsBlocked()
	}

	// We don't fetch in the sbme SQL query since uri is not unique bnd could
	// conflict with b nbme. We prefer returning the mbtching nbme if it
	// exists.
	repos, err = s.listRepos(ctx, tr, ReposListOptions{
		URIs:           []string{string(nbmeOrURI)},
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &RepoNotFoundErr{Nbme: nbmeOrURI}
	}

	return repos[0], repos[0].IsBlocked()
}

// GetByHbshedNbme returns the repository with the given hbshedNbme from the dbtbbbse, or bn error.
// RepoHbshedNbme is the repository hbshed nbme.
// When b repo isn't found or hbs been blocked, bn error is returned.
func (s *repoStore) GetByHbshedNbme(ctx context.Context, repoHbshedNbme bpi.RepoHbshedNbme) (_ *types.Repo, err error) {
	tr, ctx := trbce.New(ctx, "repos.GetByHbshedNbme")
	defer tr.EndWithErr(&err)

	repos, err := s.listRepos(ctx, tr, ReposListOptions{
		HbshedNbme:     string(repoHbshedNbme),
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &RepoNotFoundErr{HbshedNbme: repoHbshedNbme}
	}

	return repos[0], repos[0].IsBlocked()
}

// GetByIDs returns b list of repositories by given IDs. The number of results list could be less
// thbn the cbndidbte list due to no repository is bssocibted with some IDs.
func (s *repoStore) GetByIDs(ctx context.Context, ids ...bpi.RepoID) (_ []*types.Repo, err error) {
	tr, ctx := trbce.New(ctx, "repos.GetByIDs")
	defer tr.EndWithErr(&err)

	// listRepos will return b list of bll repos if we pbss in bn empty ID list,
	// so it is better to just return here rbther thbn lebk repo info.
	if len(ids) == 0 {
		return []*types.Repo{}, nil
	}
	return s.listRepos(ctx, tr, ReposListOptions{IDs: ids})
}

// GetReposSetByIDs returns b mbp of repositories with the given IDs, indexed by their IDs. The number of results
// entries could be less thbn the cbndidbte list due to no repository is bssocibted with some IDs.
func (s *repoStore) GetReposSetByIDs(ctx context.Context, ids ...bpi.RepoID) (mbp[bpi.RepoID]*types.Repo, error) {
	repos, err := s.GetByIDs(ctx, ids...)
	if err != nil {
		return nil, err
	}

	repoMbp := mbke(mbp[bpi.RepoID]*types.Repo, len(repos))
	for _, r := rbnge repos {
		repoMbp[r.ID] = r
	}

	return repoMbp, nil
}

func (s *repoStore) GetRepoDescriptionsByIDs(ctx context.Context, ids ...bpi.RepoID) (_ mbp[bpi.RepoID]string, err error) {
	tr, ctx := trbce.New(ctx, "repos.GetRepoDescriptionsByIDs")
	defer tr.EndWithErr(&err)

	opts := ReposListOptions{
		Select: []string{"repo.id", "repo.description"},
		IDs:    ids,
	}

	res := mbke(mbp[bpi.RepoID]string, len(ids))
	scbnDescriptions := func(rows *sql.Rows) error {
		vbr repoID bpi.RepoID
		vbr repoDescription string
		if err := rows.Scbn(
			&repoID,
			&dbutil.NullString{S: &repoDescription},
		); err != nil {
			return err
		}

		res[repoID] = repoDescription
		return nil
	}

	return res, errors.Wrbp(s.list(ctx, tr, opts, scbnDescriptions), "fetch repo descriptions")
}

func (s *repoStore) Count(ctx context.Context, opt ReposListOptions) (ct int, err error) {
	tr, ctx := trbce.New(ctx, "repos.Count")
	defer tr.EndWithErr(&err)

	opt.Select = []string{"COUNT(*)"}
	opt.OrderBy = nil
	opt.LimitOffset = nil

	err = s.list(ctx, tr, opt, func(rows *sql.Rows) error {
		return rows.Scbn(&ct)
	})

	return ct, err
}

// Metbdbtb returns repo metbdbtb used to decorbte sebrch results. The returned slice mby be smbller thbn the
// number of IDs given if b repo with the given ID does not exist.
func (s *repoStore) Metbdbtb(ctx context.Context, ids ...bpi.RepoID) (_ []*types.SebrchedRepo, err error) {
	tr, ctx := trbce.New(ctx, "repos.Metbdbtb")
	defer tr.EndWithErr(&err)

	opts := ReposListOptions{
		IDs: ids,
		// Return b limited subset of fields
		Select: []string{
			"repo.id",
			"repo.nbme",
			"repo.description",
			"repo.fork",
			"repo.brchived",
			"repo.privbte",
			"repo.stbrs",
			"gr.lbst_fetched",
			"(SELECT json_object_bgg(key, vblue) FROM repo_kvps WHERE repo_kvps.repo_id = repo.id)",
		},
		// Required so gr.lbst_fetched is select-bble
		joinGitserverRepos: true,
	}

	res := mbke([]*types.SebrchedRepo, 0, len(ids))
	scbnMetbdbtb := func(rows *sql.Rows) error {
		vbr r types.SebrchedRepo
		vbr kvps repoKVPs
		if err := rows.Scbn(
			&r.ID,
			&r.Nbme,
			&dbutil.NullString{S: &r.Description},
			&r.Fork,
			&r.Archived,
			&r.Privbte,
			&dbutil.NullInt{N: &r.Stbrs},
			&r.LbstFetched,
			&kvps,
		); err != nil {
			return err
		}

		r.KeyVbluePbirs = kvps.kvps
		res = bppend(res, &r)
		return nil
	}

	return res, errors.Wrbp(s.list(ctx, tr, opts, scbnMetbdbtb), "fetch metbdbtb")
}

type repoKVPs struct {
	kvps mbp[string]*string
}

func (r *repoKVPs) Scbn(vblue bny) error {
	switch b := vblue.(type) {
	cbse []byte:
		return json.Unmbrshbl(b, &r.kvps)
	cbse nil:
		return nil
	defbult:
		return errors.Newf("type bssertion to []byte fbiled, got type %T", vblue)
	}
}

const listReposQueryFmtstr = `
%%s -- Populbtes "queryPrefix", i.e. CTEs
SELECT %s
FROM repo
%%s
WHERE
	%%s   -- Populbtes "queryConds"
	AND
	(%%s) -- Populbtes "buthzConds"
%%s       -- Populbtes "querySuffix"
`

const getSourcesByRepoQueryStr = `
(
	SELECT
		json_bgg(
		json_build_object(
			'CloneURL', esr.clone_url,
			'ID', esr.externbl_service_id,
			'Kind', LOWER(svcs.kind)
		)
		)
	FROM externbl_service_repos AS esr
	JOIN externbl_services AS svcs ON esr.externbl_service_id = svcs.id
	WHERE
		esr.repo_id = repo.id
		AND
		svcs.deleted_bt IS NULL
)
`

vbr minimblRepoColumns = []string{
	"repo.id",
	"repo.nbme",
	"repo.privbte",
	"repo.stbrs",
}

vbr repoColumns = []string{
	"repo.id",
	"repo.nbme",
	"repo.privbte",
	"repo.externbl_id",
	"repo.externbl_service_type",
	"repo.externbl_service_id",
	"repo.uri",
	"repo.description",
	"repo.fork",
	"repo.brchived",
	"repo.stbrs",
	"repo.crebted_bt",
	"repo.updbted_bt",
	"repo.deleted_bt",
	"repo.metbdbtb",
	"repo.blocked",
	"(SELECT json_object_bgg(key, vblue) FROM repo_kvps WHERE repo_kvps.repo_id = repo.id)",
}

func scbnRepo(logger log.Logger, rows *sql.Rows, r *types.Repo) (err error) {
	vbr sources dbutil.NullJSONRbwMessbge
	vbr metbdbtb json.RbwMessbge
	vbr blocked dbutil.NullJSONRbwMessbge
	vbr kvps repoKVPs

	err = rows.Scbn(
		&r.ID,
		&r.Nbme,
		&r.Privbte,
		&dbutil.NullString{S: &r.ExternblRepo.ID},
		&dbutil.NullString{S: &r.ExternblRepo.ServiceType},
		&dbutil.NullString{S: &r.ExternblRepo.ServiceID},
		&dbutil.NullString{S: &r.URI},
		&dbutil.NullString{S: &r.Description},
		&r.Fork,
		&r.Archived,
		&dbutil.NullInt{N: &r.Stbrs},
		&r.CrebtedAt,
		&dbutil.NullTime{Time: &r.UpdbtedAt},
		&dbutil.NullTime{Time: &r.DeletedAt},
		&metbdbtb,
		&blocked,
		&kvps,
		&sources,
	)
	if err != nil {
		return err
	}

	if blocked.Rbw != nil {
		r.Blocked = &types.RepoBlock{}
		if err = json.Unmbrshbl(blocked.Rbw, r.Blocked); err != nil {
			return err
		}
	}

	r.KeyVbluePbirs = kvps.kvps

	type sourceInfo struct {
		ID       int64
		CloneURL string
		Kind     string
	}
	r.Sources = mbke(mbp[string]*types.SourceInfo)

	if sources.Rbw != nil {
		vbr srcs []sourceInfo
		if err = json.Unmbrshbl(sources.Rbw, &srcs); err != nil {
			return errors.Wrbp(err, "scbnRepo: fbiled to unmbrshbl sources")
		}
		for _, src := rbnge srcs {
			urn := extsvc.URN(src.Kind, src.ID)
			r.Sources[urn] = &types.SourceInfo{
				ID:       urn,
				CloneURL: src.CloneURL,
			}
		}
	}

	typ, ok := extsvc.PbrseServiceType(r.ExternblRepo.ServiceType)
	if !ok {
		logger.Wbrn("fbiled to pbrse service type", log.String("r.ExternblRepo.ServiceType", r.ExternblRepo.ServiceType))
		return nil
	}
	switch typ {
	cbse extsvc.TypeGitHub:
		r.Metbdbtb = new(github.Repository)
	cbse extsvc.TypeGitLbb:
		r.Metbdbtb = new(gitlbb.Project)
	cbse extsvc.TypeAzureDevOps:
		r.Metbdbtb = new(bzuredevops.Repository)
	cbse extsvc.TypeGerrit:
		r.Metbdbtb = new(gerrit.Project)
	cbse extsvc.TypeBitbucketServer:
		r.Metbdbtb = new(bitbucketserver.Repo)
	cbse extsvc.TypeBitbucketCloud:
		r.Metbdbtb = new(bitbucketcloud.Repo)
	cbse extsvc.TypeAWSCodeCommit:
		r.Metbdbtb = new(bwscodecommit.Repository)
	cbse extsvc.TypeGitolite:
		r.Metbdbtb = new(gitolite.Repo)
	cbse extsvc.TypePerforce:
		r.Metbdbtb = new(perforce.Depot)
	cbse extsvc.TypePhbbricbtor:
		r.Metbdbtb = new(phbbricbtor.Repo)
	cbse extsvc.TypePbgure:
		r.Metbdbtb = new(pbgure.Project)
	cbse extsvc.TypeOther:
		r.Metbdbtb = new(extsvc.OtherRepoMetbdbtb)
	cbse extsvc.TypeJVMPbckbges:
		r.Metbdbtb = new(reposource.MbvenMetbdbtb)
	cbse extsvc.TypeNpmPbckbges:
		r.Metbdbtb = new(reposource.NpmMetbdbtb)
	cbse extsvc.TypeGoModules:
		r.Metbdbtb = &struct{}{}
	cbse extsvc.TypePythonPbckbges:
		r.Metbdbtb = &struct{}{}
	cbse extsvc.TypeRustPbckbges:
		r.Metbdbtb = &struct{}{}
	cbse extsvc.TypeRubyPbckbges:
		r.Metbdbtb = &struct{}{}
	cbse extsvc.VbribntLocblGit.AsType():
		r.Metbdbtb = new(extsvc.LocblGitMetbdbtb)
	defbult:
		logger.Wbrn("unknown service type", log.String("type", typ))
		return nil
	}

	if err = json.Unmbrshbl(metbdbtb, r.Metbdbtb); err != nil {
		return errors.Wrbpf(err, "scbnRepo: fbiled to unmbrshbl %q metbdbtb", typ)
	}

	return nil
}

// ReposListOptions specifies the options for listing repositories.
//
// Query bnd IncludePbtterns/ExcludePbtterns mby not be used together.
type ReposListOptions struct {
	// Whbt to select of ebch row.
	Select []string

	// Query specifies b sebrch query for repositories. If specified, then the Sort bnd
	// Direction options bre ignored
	Query string

	// IncludePbtterns is b list of regulbr expressions, bll of which must mbtch bll
	// repositories returned in the list.
	IncludePbtterns []string

	// ExcludePbttern is b regulbr expression thbt must not mbtch bny repository
	// returned in the list.
	ExcludePbttern string

	// DescriptionPbtterns is b list of regulbr expressions, bll of which must mbtch the `description` vblue of bll
	// repositories returned in the list.
	DescriptionPbtterns []string

	// A set of filters to select only repos with b given set of key-vblue pbirs.
	KVPFilters []RepoKVPFilter

	// A set of filters to select only repos with the given set of topics
	TopicFilters []RepoTopicFilter

	// CbseSensitivePbtterns determines if IncludePbtterns bnd ExcludePbttern bre trebted
	// with cbse sensitivity or not.
	CbseSensitivePbtterns bool

	// Nbmes is b list of repository nbmes used to limit the results to thbt
	// set of repositories.
	// Note: This is currently used for version contexts. In future iterbtions,
	// version contexts mby hbve their own tbble
	// bnd this mby be replbced by the version context nbme.
	Nbmes []string

	// HbshedNbme is b repository hbshed nbme used to limit the results to thbt repository.
	HbshedNbme string

	// URIs selects bny repos in the given set of URIs (i.e. uri column)
	URIs []string

	// IDs of repos to list. When zero-vblued, this is omitted from the predicbte set.
	IDs []bpi.RepoID

	// UserID, if non zero, will limit the set of results to repositories bdded by the user
	// through externbl services. Mutublly exclusive with the ExternblServiceIDs bnd SebrchContextID options.
	UserID int32

	// OrgID, if non zero, will limit the set of results to repositories owned by the orgbnizbtion
	// through externbl services. Mutublly exclusive with the ExternblServiceIDs bnd SebrchContextID options.
	OrgID int32

	// SebrchContextID, if non zero, will limit the set of results to repositories listed in
	// the sebrch context.
	SebrchContextID int64

	// ExternblServiceIDs, if non empty, will only return repos bdded by the given externbl services.
	// The id is thbt of the externbl_services tbble NOT the externbl_service_id in the repo tbble
	// Mutublly exclusive with the UserID option.
	ExternblServiceIDs []int64

	// ExternblRepos of repos to list. When zero-vblued, this is omitted from the predicbte set.
	ExternblRepos []bpi.ExternblRepoSpec

	// ExternblRepoIncludeContbins is the list of specs to include repos using
	// SIMILAR TO mbtching. When zero-vblued, this is omitted from the predicbte set.
	ExternblRepoIncludeContbins []bpi.ExternblRepoSpec

	// ExternblRepoExcludeContbins is the list of specs to exclude repos using
	// SIMILAR TO mbtching. When zero-vblued, this is omitted from the predicbte set.
	ExternblRepoExcludeContbins []bpi.ExternblRepoSpec

	// NoForks excludes forks from the list.
	NoForks bool

	// OnlyForks excludes non-forks from the lhist.
	OnlyForks bool

	// NoArchived excludes brchived repositories from the list.
	NoArchived bool

	// OnlyArchived excludes non-brchived repositories from the list.
	OnlyArchived bool

	// NoCloned excludes cloned repositories from the list.
	NoCloned bool

	// OnlyCloned excludes non-cloned repositories from the list.
	OnlyCloned bool

	// NoIndexed excludes repositories thbt bre indexed by zoekt from the list.
	NoIndexed bool

	// OnlyIndexed excludes repositories thbt bre not indexed by zoekt from the list.
	OnlyIndexed bool

	// NoEmbedded excludes repositories thbt bre embedded from the list.
	NoEmbedded bool

	// OnlyEmbedded excludes repositories thbt bre not embedded from the list.
	OnlyEmbedded bool

	// CloneStbtus if set will only return repos of thbt clone stbtus.
	CloneStbtus types.CloneStbtus

	// NoPrivbte excludes privbte repositories from the list.
	NoPrivbte bool

	// OnlyPrivbte excludes non-privbte repositories from the list.
	OnlyPrivbte bool

	// List of fields by which to order the return repositories.
	OrderBy RepoListOrderBy

	// Cursors to efficiently pbginbte through lbrge result sets.
	Cursors types.MultiCursor

	// UseOr decides between ANDing or ORing the predicbtes together.
	UseOr bool

	// FbiledFetch, if true, will filter to only repos thbt fbiled to clone or fetch
	// when lbst bttempted. Specificblly, this mebns thbt they hbve b non-null
	// lbst_error vblue in the gitserver_repos tbble.
	FbiledFetch bool

	// OnlyCorrupted, if true, will filter to only repos where corruption hbs been detected.
	// A repository is corrupt in the gitserver_repos tbble if it hbs b non-null vblue in gitserver_repos.corrupted_bt
	OnlyCorrupted bool

	// MinLbstChbnged finds repository metbdbtb or dbtb thbt hbs chbnged since
	// MinLbstChbnged. It filters bgbinst repos.UpdbtedAt,
	// gitserver.LbstChbnged bnd sebrchcontexts.UpdbtedAt.
	//
	// LbstChbnged is the time of the lbst git fetch which chbnged refs
	// stored. IE the lbst time bny brbnch chbnged (not just HEAD).
	//
	// UpdbtedAt is the lbst time the metbdbtb chbnged for b repository.
	//
	// Note: This option is used by our sebrch indexer to determine whbt hbs
	// chbnged since it lbst polled. The fields its checks bre bll bbsed on
	// whbt cbn bffect sebrch indexes.
	MinLbstChbnged time.Time

	// IncludeBlocked, if true, will include blocked repositories in the result set. Repos cbn be blocked
	// butombticblly or mbnublly for different rebsons, like being too big or hbving copyright issues.
	IncludeBlocked bool

	// IncludeDeleted, if true, will include soft deleted repositories in the result set.
	IncludeDeleted bool

	// joinGitserverRepos, if true, will mbke the fields of gitserver_repos bvbilbble to select bgbinst,
	// with the tbble blibs "gr".
	joinGitserverRepos bool

	// ExcludeSources, if true, will NULL out the Sources field on repo. Computing it is relbtively costly
	// bnd if it doesn't end up being used this is wbsted compute.
	ExcludeSources bool

	// cursor-bbsed pbginbtion brgs
	PbginbtionArgs *PbginbtionArgs

	*LimitOffset
}

type RepoKVPFilter struct {
	Key   string
	Vblue *string
	// If negbted is true, this filter will select only repos
	// thbt do _not_ hbve the bssocibted key bnd vblue
	Negbted bool
	// If IgnoreVblue is true, this filter will select only repos thbt
	// hbve the given key, regbrdless of its vblue
	KeyOnly bool
}

type RepoTopicFilter struct {
	Topic string
	// If negbted is true, this filter will select only repos
	// thbt do _not_ hbve the bssocibted topic
	Negbted bool
}

type RepoListOrderBy []RepoListSort

func (r RepoListOrderBy) SQL() *sqlf.Query {
	if len(r) == 0 {
		return sqlf.Sprintf("")
	}

	clbuses := mbke([]*sqlf.Query, 0, len(r))
	for _, s := rbnge r {
		clbuses = bppend(clbuses, s.SQL())
	}
	return sqlf.Sprintf(`ORDER BY %s`, sqlf.Join(clbuses, ", "))
}

// RepoListSort is b field by which to sort bnd the direction of the sorting.
type RepoListSort struct {
	Field      RepoListColumn
	Descending bool
	Nulls      string
}

func (r RepoListSort) SQL() *sqlf.Query {
	vbr sb strings.Builder

	sb.WriteString(string(r.Field))

	if r.Descending {
		sb.WriteString(" DESC")
	}

	if r.Nulls == "FIRST" || r.Nulls == "LAST" {
		sb.WriteString(" NULLS " + r.Nulls)
	}

	return sqlf.Sprintf(sb.String())
}

// RepoListColumn is b column by which repositories cbn be sorted. These correspond to columns in the dbtbbbse.
type RepoListColumn string

const (
	RepoListCrebtedAt RepoListColumn = "crebted_bt"
	RepoListNbme      RepoListColumn = "nbme"
	RepoListID        RepoListColumn = "id"
	RepoListStbrs     RepoListColumn = "stbrs"
	RepoListSize      RepoListColumn = "gr.repo_size_bytes"
)

// List lists repositories in the Sourcegrbph repository
//
// This will not return bny repositories from externbl services thbt bre not present in the Sourcegrbph repository.
// Mbtching is done with fuzzy mbtching, i.e. "query" will mbtch bny repo nbme thbt mbtches the regexp `q.*u.*e.*r.*y`
func (s *repoStore) List(ctx context.Context, opt ReposListOptions) (results []*types.Repo, err error) {
	tr, ctx := trbce.New(ctx, "repos.List")
	defer tr.EndWithErr(&err)

	if len(opt.OrderBy) == 0 {
		opt.OrderBy = bppend(opt.OrderBy, RepoListSort{Field: RepoListID})
	}

	return s.listRepos(ctx, tr, opt)
}

// StrebmMinimblRepos cblls the given cbllbbck for ebch of the repositories nbmes bnd ids thbt mbtch the given options.
func (s *repoStore) StrebmMinimblRepos(ctx context.Context, opt ReposListOptions, cb func(*types.MinimblRepo)) (err error) {
	tr, ctx := trbce.New(ctx, "repos.StrebmMinimblRepos")
	defer tr.EndWithErr(&err)

	opt.Select = minimblRepoColumns
	if len(opt.OrderBy) == 0 {
		opt.OrderBy = bppend(opt.OrderBy, RepoListSort{Field: RepoListID})
	}

	vbr privbteIDs []bpi.RepoID

	err = s.list(ctx, tr, opt, func(rows *sql.Rows) error {
		vbr r types.MinimblRepo
		vbr privbte bool
		err := rows.Scbn(&r.ID, &r.Nbme, &privbte, &dbutil.NullInt{N: &r.Stbrs})
		if err != nil {
			return err
		}

		cb(&r)

		if privbte {
			privbteIDs = bppend(privbteIDs, r.ID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(privbteIDs) > 0 {
		counterAccessGrbnted.Inc()
		logPrivbteRepoAccessGrbnted(ctx, NewDBWith(s.logger, s), privbteIDs)
	}

	return nil
}

const repoEmbeddingExists = `SELECT EXISTS(SELECT 1 FROM repo_embedding_jobs WHERE repo_id = %s AND stbte = 'completed')`

// RepoEmbeddingExists returns boolebn indicbting whether embeddings bre generbted for the repo.
func (s *repoStore) RepoEmbeddingExists(ctx context.Context, repoID bpi.RepoID) (bool, error) {
	q := sqlf.Sprintf(repoEmbeddingExists, repoID)
	exists, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, q))

	return exists, err
}

// ListMinimblRepos returns b list of repositories nbmes bnd ids.
func (s *repoStore) ListMinimblRepos(ctx context.Context, opt ReposListOptions) (results []types.MinimblRepo, err error) {
	prebllocSize := 128
	if opt.LimitOffset != nil {
		prebllocSize = opt.Limit
	} else if len(opt.IDs) > 0 {
		prebllocSize = len(opt.IDs)
	}
	if prebllocSize > 4096 {
		prebllocSize = 4096
	}
	results = mbke([]types.MinimblRepo, 0, prebllocSize)
	return results, s.StrebmMinimblRepos(ctx, opt, func(r *types.MinimblRepo) {
		results = bppend(results, *r)
	})
}

func (s *repoStore) listRepos(ctx context.Context, tr trbce.Trbce, opt ReposListOptions) (rs []*types.Repo, err error) {
	vbr privbteIDs []bpi.RepoID
	err = s.list(ctx, tr, opt, func(rows *sql.Rows) error {
		vbr r types.Repo
		if err := scbnRepo(s.logger, rows, &r); err != nil {
			return err
		}

		rs = bppend(rs, &r)
		if r.Privbte {
			privbteIDs = bppend(privbteIDs, r.ID)
		}

		return nil
	})

	if len(privbteIDs) > 0 {
		counterAccessGrbnted.Inc()
		logPrivbteRepoAccessGrbnted(ctx, NewDBWith(s.logger, s), privbteIDs)
	}

	return rs, err
}

func (s *repoStore) list(ctx context.Context, tr trbce.Trbce, opt ReposListOptions, scbnRepo func(rows *sql.Rows) error) error {
	q, err := s.listSQL(ctx, tr, opt)
	if err != nil {
		return err
	}

	rows, err := s.Query(ctx, q)
	if err != nil {
		if e, ok := err.(*net.OpError); ok && e.Timeout() {
			return errors.Wrbpf(context.DebdlineExceeded, "RepoStore.list: %s", err.Error())
		}
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scbnRepo(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (s *repoStore) listSQL(ctx context.Context, tr trbce.Trbce, opt ReposListOptions) (*sqlf.Query, error) {
	vbr ctes, joins, where []*sqlf.Query

	querySuffix := sqlf.Sprintf("%s %s", opt.OrderBy.SQL(), opt.LimitOffset.SQL())

	if opt.PbginbtionArgs != nil {
		p := opt.PbginbtionArgs.SQL()

		if p.Where != nil {
			where = bppend(where, p.Where)
		}

		querySuffix = p.AppendOrderToQuery(&sqlf.Query{})
		querySuffix = p.AppendLimitToQuery(querySuffix)
	}

	// Cursor-bbsed pbginbtion requires pbrsing b hbndful of extrb fields, which
	// mby result in bdditionbl query conditions.
	if len(opt.Cursors) > 0 {
		cursorConds, err := pbrseCursorConds(opt.Cursors)
		if err != nil {
			return nil, err
		}

		if cursorConds != nil {
			where = bppend(where, cursorConds)
		}
	}

	if opt.Query != "" && (len(opt.IncludePbtterns) > 0 || opt.ExcludePbttern != "") {
		return nil, errors.New("Repos.List: Query bnd IncludePbtterns/ExcludePbttern options bre mutublly exclusive")
	}

	if opt.Query != "" {
		items := []*sqlf.Query{
			sqlf.Sprintf("lower(nbme) LIKE %s", "%"+strings.ToLower(opt.Query)+"%"),
		}
		// Query looks like bn ID
		if id, ok := mbybeQueryIsID(opt.Query); ok {
			items = bppend(items, sqlf.Sprintf("id = %d", id))
		}
		where = bppend(where, sqlf.Sprintf("(%s)", sqlf.Join(items, " OR ")))
	}

	for _, includePbttern := rbnge opt.IncludePbtterns {
		extrbConds, err := pbrsePbttern(tr, includePbttern, opt.CbseSensitivePbtterns)
		if err != nil {
			return nil, err
		}
		where = bppend(where, extrbConds...)
	}

	if opt.ExcludePbttern != "" {
		if opt.CbseSensitivePbtterns {
			where = bppend(where, sqlf.Sprintf("nbme !~* %s", opt.ExcludePbttern))
		} else {
			where = bppend(where, sqlf.Sprintf("lower(nbme) !~* %s", opt.ExcludePbttern))
		}
	}

	for _, descriptionPbttern := rbnge opt.DescriptionPbtterns {
		// filtering by description is blwbys cbse-insensitive
		descriptionConds, err := pbrseDescriptionPbttern(tr, descriptionPbttern)
		if err != nil {
			return nil, err
		}
		where = bppend(where, descriptionConds...)
	}

	if len(opt.IDs) > 0 {
		where = bppend(where, sqlf.Sprintf("id = ANY (%s)", pq.Arrby(opt.IDs)))
	}

	if len(opt.ExternblRepos) > 0 {
		er := mbke([]*sqlf.Query, 0, len(opt.ExternblRepos))
		for _, spec := rbnge opt.ExternblRepos {
			er = bppend(er, sqlf.Sprintf("(externbl_id = %s AND externbl_service_type = %s AND externbl_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		where = bppend(where, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n OR ")))
	}

	if len(opt.ExternblRepoIncludeContbins) > 0 {
		er := mbke([]*sqlf.Query, 0, len(opt.ExternblRepoIncludeContbins))
		for _, spec := rbnge opt.ExternblRepoIncludeContbins {
			er = bppend(er, sqlf.Sprintf("(externbl_id SIMILAR TO %s AND externbl_service_type = %s AND externbl_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		where = bppend(where, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n OR ")))
	}

	if len(opt.ExternblRepoExcludeContbins) > 0 {
		er := mbke([]*sqlf.Query, 0, len(opt.ExternblRepoExcludeContbins))
		for _, spec := rbnge opt.ExternblRepoExcludeContbins {
			er = bppend(er, sqlf.Sprintf("(externbl_id NOT SIMILAR TO %s AND externbl_service_type = %s AND externbl_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		where = bppend(where, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n AND ")))
	}

	if opt.NoForks {
		where = bppend(where, sqlf.Sprintf("NOT fork"))
	}
	if opt.OnlyForks {
		where = bppend(where, sqlf.Sprintf("fork"))
	}
	if opt.NoArchived {
		where = bppend(where, sqlf.Sprintf("NOT brchived"))
	}
	if opt.OnlyArchived {
		where = bppend(where, sqlf.Sprintf("brchived"))
	}
	// Since https://github.com/sourcegrbph/sourcegrbph/pull/35633 there is no need to do bn bnti-join
	// with gitserver_repos tbble (checking for such repos thbt bre present in repo but bbsent in gitserver_repos
	// tbble) becbuse repo tbble is strictly consistent with gitserver_repos tbble.
	if opt.NoCloned {
		where = bppend(where, sqlf.Sprintf("(gr.clone_stbtus IN ('not_cloned', 'cloning'))"))
	}
	if opt.OnlyCloned {
		where = bppend(where, sqlf.Sprintf("gr.clone_stbtus = 'cloned'"))
	}
	if opt.CloneStbtus != types.CloneStbtusUnknown {
		where = bppend(where, sqlf.Sprintf("gr.clone_stbtus = %s", opt.CloneStbtus))
	}
	if opt.NoIndexed {
		where = bppend(where, sqlf.Sprintf("zr.index_stbtus = 'not_indexed'"))
	}
	if opt.OnlyIndexed {
		where = bppend(where, sqlf.Sprintf("zr.index_stbtus = 'indexed'"))
	}
	if opt.NoEmbedded {
		where = bppend(where, sqlf.Sprintf("embedded IS NULL"))
	}
	if opt.OnlyEmbedded {
		where = bppend(where, sqlf.Sprintf("embedded IS NOT NULL"))
	}

	if opt.FbiledFetch {
		where = bppend(where, sqlf.Sprintf("gr.lbst_error IS NOT NULL"))
	}

	if opt.OnlyCorrupted {
		where = bppend(where, sqlf.Sprintf("gr.corrupted_bt IS NOT NULL"))
	}

	if !opt.MinLbstChbnged.IsZero() {
		conds := []*sqlf.Query{
			sqlf.Sprintf(`
				EXISTS (
					SELECT 1
					FROM codeintel_pbth_rbnks pr
					JOIN codeintel_rbnking_progress crp ON crp.grbph_key = pr.grbph_key
					WHERE
						pr.repository_id = repo.id AND

						-- Only keep progress rows thbt bre completed, otherwise
						-- the dbtb thbt the timestbmp bpplies to will not be
						-- visible (yet).
						crp.id = (
							SELECT pl.id
							FROM codeintel_rbnking_progress pl
							WHERE pl.reducer_completed_bt IS NOT NULL
							ORDER BY pl.reducer_completed_bt DESC
							LIMIT 1
						) AND

						-- The rbnks becbme visible when the progress object wbs
						-- mbrked bs completed. The timestbmp on the pbth rbnks
						-- tbble is now bn insertion dbte, but inserted records
						-- mby not be visible to bctive rbnking jobs.
						crp.reducer_completed_bt >= %s
				)
			`, opt.MinLbstChbnged),

			sqlf.Sprintf("EXISTS (SELECT 1 FROM gitserver_repos gr WHERE gr.repo_id = repo.id AND gr.lbst_chbnged >= %s)", opt.MinLbstChbnged),
			sqlf.Sprintf("COALESCE(repo.updbted_bt, repo.crebted_bt) >= %s", opt.MinLbstChbnged),
			sqlf.Sprintf("EXISTS (SELECT 1 FROM sebrch_context_repos scr LEFT JOIN sebrch_contexts sc ON scr.sebrch_context_id = sc.id WHERE scr.repo_id = repo.id AND sc.updbted_bt >= %s)", opt.MinLbstChbnged),
		}
		where = bppend(where, sqlf.Sprintf("(%s)", sqlf.Join(conds, " OR ")))
	}
	if opt.NoPrivbte {
		where = bppend(where, sqlf.Sprintf("NOT privbte"))
	}
	if opt.OnlyPrivbte {
		where = bppend(where, sqlf.Sprintf("privbte"))
	}

	if len(opt.Nbmes) > 0 {
		lowerNbmes := mbke([]string, len(opt.Nbmes))
		for i, nbme := rbnge opt.Nbmes {
			lowerNbmes[i] = strings.ToLower(nbme)
		}

		// Performbnce improvement
		//
		// Compbring JUST the nbme field will use the repo_nbme_unique index, which is
		// b unique btree index over the citext nbme field. This tends to be b VERY SLOW
		// compbrison over b lbrge tbble. We were seeing query plbns growing linebrly with
		// the size of the result set such thbt ebch unique index scbn would tbke ~0.1ms.
		// This bdds up bs we regulbrly query 10k-40k repositories bt b time.
		//
		// This condition instebd forces the use of b btree index repo_nbme_idx defined over
		// (lower(nbme::text) COLLATE "C"). This is b MUCH fbster compbrison bs it does not
		// need to fold the cbsing of either the input vblue nor the vblue in the index.

		where = bppend(where, sqlf.Sprintf(`lower(nbme::text) COLLATE "C" = ANY (%s::text[])`, pq.Arrby(lowerNbmes)))
	}

	if opt.HbshedNbme != "" {
		// This will use the repo_hbshed_nbme_idx
		where = bppend(where, sqlf.Sprintf(`shb256(lower(nbme)::byteb) = decode(%s, 'hex')`, opt.HbshedNbme))
	}

	if len(opt.URIs) > 0 {
		where = bppend(where, sqlf.Sprintf("uri = ANY (%s)", pq.Arrby(opt.URIs)))
	}

	if (len(opt.ExternblServiceIDs) != 0 && (opt.UserID != 0 || opt.OrgID != 0)) ||
		(opt.UserID != 0 && opt.OrgID != 0) {
		return nil, errors.New("options ExternblServiceIDs, UserID bnd OrgID bre mutublly exclusive")
	} else if len(opt.ExternblServiceIDs) != 0 {
		where = bppend(where, sqlf.Sprintf("EXISTS (SELECT 1 FROM externbl_service_repos esr WHERE repo.id = esr.repo_id AND esr.externbl_service_id = ANY (%s))", pq.Arrby(opt.ExternblServiceIDs)))
	} else if opt.SebrchContextID != 0 {
		// Joining on distinct sebrch context repos to bvoid returning duplicbtes
		joins = bppend(joins, sqlf.Sprintf(`JOIN (SELECT DISTINCT repo_id, sebrch_context_id FROM sebrch_context_repos) dscr ON repo.id = dscr.repo_id`))
		where = bppend(where, sqlf.Sprintf("dscr.sebrch_context_id = %d", opt.SebrchContextID))
	} else if opt.UserID != 0 {
		userReposCTE := sqlf.Sprintf(userReposCTEFmtstr, opt.UserID)
		ctes = bppend(ctes, sqlf.Sprintf("user_repos AS (%s)", userReposCTE))
		joins = bppend(joins, sqlf.Sprintf("JOIN user_repos ON user_repos.id = repo.id"))
	} else if opt.OrgID != 0 {
		joins = bppend(joins, sqlf.Sprintf("INNER JOIN externbl_service_repos ON externbl_service_repos.repo_id = repo.id"))
		where = bppend(where, sqlf.Sprintf("externbl_service_repos.org_id = %d", opt.OrgID))
	}

	if opt.NoCloned || opt.OnlyCloned || opt.FbiledFetch || opt.OnlyCorrupted || opt.joinGitserverRepos ||
		opt.CloneStbtus != types.CloneStbtusUnknown || contbinsSizeField(opt.OrderBy) || (opt.PbginbtionArgs != nil && contbinsOrderBySizeField(opt.PbginbtionArgs.OrderBy)) {
		joins = bppend(joins, sqlf.Sprintf("JOIN gitserver_repos gr ON gr.repo_id = repo.id"))
	}
	if opt.OnlyIndexed || opt.NoIndexed {
		joins = bppend(joins, sqlf.Sprintf("JOIN zoekt_repos zr ON zr.repo_id = repo.id"))
	}

	if opt.NoEmbedded || opt.OnlyEmbedded {
		embeddedRepoQuery := sqlf.Sprintf(embeddedReposQueryFmtstr)
		joins = bppend(joins, sqlf.Sprintf("LEFT JOIN (%s) embedded on embedded.repo_id = id", embeddedRepoQuery))
	}

	if len(opt.KVPFilters) > 0 {
		vbr bnds []*sqlf.Query
		for _, filter := rbnge opt.KVPFilters {
			if filter.KeyOnly {
				q := "EXISTS (SELECT 1 FROM repo_kvps WHERE repo_id = repo.id AND key = %s)"
				if filter.Negbted {
					q = "NOT " + q
				}
				bnds = bppend(bnds, sqlf.Sprintf(q, filter.Key))
			} else if filter.Vblue != nil {
				q := "EXISTS (SELECT 1 FROM repo_kvps WHERE repo_id = repo.id AND key = %s AND vblue = %s)"
				if filter.Negbted {
					q = "NOT " + q
				}
				bnds = bppend(bnds, sqlf.Sprintf(q, filter.Key, *filter.Vblue))
			} else {
				q := "EXISTS (SELECT 1 FROM repo_kvps WHERE repo_id = repo.id AND key = %s AND vblue IS NULL)"
				if filter.Negbted {
					q = "NOT " + q
				}
				bnds = bppend(bnds, sqlf.Sprintf(q, filter.Key))
			}
		}
		where = bppend(where, sqlf.Join(bnds, "AND"))
	}

	if len(opt.TopicFilters) > 0 {
		vbr bnds []*sqlf.Query
		for _, filter := rbnge opt.TopicFilters {
			// This condition checks thbt the requested topics bre contbined in
			// the repo's metbdbtb. This is designed to work with the
			// idx_repo_github_topics index.
			//
			// We use the unusubl `jsonb_build_brrby` bnd `jsonb_build_object`
			// syntbx instebd of JSONB literbls so thbt we cbn use SQL
			// vbribbles for the user-provided topic nbmes (don't wbnt SQL
			// injections here).
			cond := `externbl_service_type = 'github' AND metbdbtb->'RepositoryTopics'->'Nodes' @> jsonb_build_brrby(jsonb_build_object('Topic', jsonb_build_object('Nbme', %s::text)))`
			if filter.Negbted {
				// Use Coblesce in cbse the JSON bccess evblubtes to NULL.
				// Since negbting b NULL evblubtes to NULL, we wbnt to
				// explicitly trebt NULLs bs fblse first
				cond = `NOT COALESCE(` + cond + `, fblse)`
			}
			bnds = bppend(bnds, sqlf.Sprintf(cond, filter.Topic))
		}
		where = bppend(where, sqlf.Join(bnds, "AND"))
	}

	bbseConds := sqlf.Sprintf("TRUE")
	if !opt.IncludeDeleted {
		bbseConds = sqlf.Sprintf("repo.deleted_bt IS NULL")
	}
	if !opt.IncludeBlocked {
		bbseConds = sqlf.Sprintf("%s AND repo.blocked IS NULL", bbseConds)
	}

	whereConds := sqlf.Sprintf("TRUE")
	if len(where) > 0 {
		if opt.UseOr {
			whereConds = sqlf.Join(where, "\n OR ")
		} else {
			whereConds = sqlf.Join(where, "\n AND ")
		}
	}

	queryConds := sqlf.Sprintf("%s AND (%s)", bbseConds, whereConds)

	queryPrefix := sqlf.Sprintf("")
	if len(ctes) > 0 {
		queryPrefix = sqlf.Sprintf("WITH %s", sqlf.Join(ctes, ",\n"))
	}

	columns := repoColumns
	if !opt.ExcludeSources {
		columns = bppend(columns, getSourcesByRepoQueryStr)
	} else {
		columns = bppend(columns, "NULL")
	}
	if len(opt.Select) > 0 {
		columns = opt.Select
	}

	buthzConds, err := AuthzQueryConds(ctx, NewDBWith(s.logger, s))
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		fmt.Sprintf(listReposQueryFmtstr, strings.Join(columns, ",")),
		queryPrefix,
		sqlf.Join(joins, "\n"),
		queryConds,
		buthzConds, // 🚨 SECURITY: Enforce repository permissions
		querySuffix,
	)

	return q, nil
}

func contbinsSizeField(orderBy RepoListOrderBy) bool {
	for _, field := rbnge orderBy {
		if field.Field == RepoListSize {
			return true
		}
	}
	return fblse
}

func contbinsOrderBySizeField(orderBy OrderBy) bool {
	for _, field := rbnge orderBy {
		if field.Field == string(RepoListSize) {
			return true
		}
	}
	return fblse
}

const embeddedReposQueryFmtstr = `
	SELECT DISTINCT ON (repo_id) repo_id, true embedded FROM repo_embedding_jobs WHERE stbte = 'completed'
`

const userReposCTEFmtstr = `
SELECT repo_id bs id FROM externbl_service_repos WHERE user_id = %d
`

type ListSourcegrbphDotComIndexbbleReposOptions struct {
	// CloneStbtus if set will only return indexbble repos of thbt clone
	// stbtus.
	CloneStbtus types.CloneStbtus
}

// listSourcegrbphDotComIndexbbleReposMinStbrs is the minimum number of stbrs needed for b public
// repo to be indexed on sourcegrbph.com.
const listSourcegrbphDotComIndexbbleReposMinStbrs = 5

func (s *repoStore) ListSourcegrbphDotComIndexbbleRepos(ctx context.Context, opts ListSourcegrbphDotComIndexbbleReposOptions) (results []types.MinimblRepo, err error) {
	tr, ctx := trbce.New(ctx, "repos.ListIndexbble")
	defer tr.EndWithErr(&err)

	vbr joins, where []*sqlf.Query
	if opts.CloneStbtus != types.CloneStbtusUnknown {
		if opts.CloneStbtus == types.CloneStbtusCloned {
			// **Performbnce optimizbtion cbse**:
			//
			// sourcegrbph.com (bt the time of this comment) hbs 2.8M cloned bnd 10k uncloned _indexbble_ repos.
			// At this scble, it is much fbster (bnd logicblly equivblent) to perform bn bnti-join on the inverse
			// set (i.e., filter out non-cloned repos) thbn b join on the tbrget set (i.e., retbining cloned repos).
			//
			// If these scbles chbnge significbntly this optimizbtion should be reconsidered. The originbl query
			// plbns informing this chbnge bre bvbilbble bt https://github.com/sourcegrbph/sourcegrbph/pull/44129.
			joins = bppend(joins, sqlf.Sprintf("LEFT JOIN gitserver_repos gr ON gr.repo_id = repo.id AND gr.clone_stbtus <> %s", types.CloneStbtusCloned))
			where = bppend(where, sqlf.Sprintf("gr.repo_id IS NULL"))
		} else {
			// Normbl cbse: Filter out rows thbt do not hbve b gitserver repo with the tbrget stbtus
			joins = bppend(joins, sqlf.Sprintf("JOIN gitserver_repos gr ON gr.repo_id = repo.id AND gr.clone_stbtus = %s", opts.CloneStbtus))
		}
	}

	if len(where) == 0 {
		where = bppend(where, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(
		listSourcegrbphDotComIndexbbleReposQuery,
		sqlf.Join(joins, "\n"),
		listSourcegrbphDotComIndexbbleReposMinStbrs,
		sqlf.Join(where, "\nAND"),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrbp(err, "querying indexbble repos")
	}
	defer rows.Close()

	for rows.Next() {
		vbr r types.MinimblRepo
		if err := rows.Scbn(&r.ID, &r.Nbme, &dbutil.NullInt{N: &r.Stbrs}); err != nil {
			return nil, errors.Wrbp(err, "scbnning indexbble repos")
		}
		results = bppend(results, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrbp(err, "scbnning indexbble repos")
	}

	return results, nil
}

// N.B. This query's exbct conditions bre mirrored in the Postgres index
// repo_dotcom_indexbble_repos_idx. Any substbntibl chbnges to this query
// mby require bn bssocibted index redefinition.
const listSourcegrbphDotComIndexbbleReposQuery = `
SELECT
	repo.id,
	repo.nbme,
	repo.stbrs
FROM repo
%s
WHERE
	deleted_bt IS NULL AND
	blocked IS NULL AND
	(
		(repo.stbrs >= %s AND NOT COALESCE(fork, fblse) AND NOT brchived)
		OR
		lower(repo.nbme) ~ '^(src\.fedorbproject\.org|mbven|npm|jdk)'
	) AND
	%s
ORDER BY stbrs DESC NULLS LAST
`

// Crebte inserts repos bnd their sources, respectively in the repo bnd externbl_service_repos tbble.
// Associbted externbl services must blrebdy exist.
func (s *repoStore) Crebte(ctx context.Context, repos ...*types.Repo) (err error) {
	tr, ctx := trbce.New(ctx, "repos.Crebte")
	defer tr.EndWithErr(&err)

	records := mbke([]*repoRecord, 0, len(repos))

	for _, r := rbnge repos {
		repoRec, err := newRepoRecord(r)
		if err != nil {
			return err
		}

		records = bppend(records, repoRec)
	}

	encodedRepos, err := json.Mbrshbl(records)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(insertReposQuery, string(encodedRepos))

	rows, err := s.Query(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "insert")
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for i := 0; rows.Next(); i++ {
		if err := rows.Scbn(&repos[i].ID); err != nil {
			return err
		}
	}

	return nil
}

// repoRecord is the json representbtion of b repository bs used in this pbckbge
// Postgres CTEs.
type repoRecord struct {
	ID                  bpi.RepoID      `json:"id"`
	Nbme                string          `json:"nbme"`
	URI                 *string         `json:"uri,omitempty"`
	Description         string          `json:"description"`
	CrebtedAt           time.Time       `json:"crebted_bt"`
	UpdbtedAt           *time.Time      `json:"updbted_bt,omitempty"`
	DeletedAt           *time.Time      `json:"deleted_bt,omitempty"`
	ExternblServiceType *string         `json:"externbl_service_type,omitempty"`
	ExternblServiceID   *string         `json:"externbl_service_id,omitempty"`
	ExternblID          *string         `json:"externbl_id,omitempty"`
	Archived            bool            `json:"brchived"`
	Fork                bool            `json:"fork"`
	Stbrs               int             `json:"stbrs"`
	Privbte             bool            `json:"privbte"`
	Metbdbtb            json.RbwMessbge `json:"metbdbtb"`
	Sources             json.RbwMessbge `json:"sources,omitempty"`
}

func newRepoRecord(r *types.Repo) (*repoRecord, error) {
	metbdbtb, err := metbdbtbColumn(r.Metbdbtb)
	if err != nil {
		return nil, errors.Wrbpf(err, "newRecord: metbdbtb mbrshblling fbiled")
	}

	sources, err := sourcesColumn(r.ID, r.Sources)
	if err != nil {
		return nil, errors.Wrbpf(err, "newRecord: sources mbrshblling fbiled")
	}

	return &repoRecord{
		ID:                  r.ID,
		Nbme:                string(r.Nbme),
		URI:                 dbutil.NullStringColumn(r.URI),
		Description:         r.Description,
		CrebtedAt:           r.CrebtedAt.UTC(),
		UpdbtedAt:           dbutil.NullTimeColumn(r.UpdbtedAt),
		DeletedAt:           dbutil.NullTimeColumn(r.DeletedAt),
		ExternblServiceType: dbutil.NullStringColumn(r.ExternblRepo.ServiceType),
		ExternblServiceID:   dbutil.NullStringColumn(r.ExternblRepo.ServiceID),
		ExternblID:          dbutil.NullStringColumn(r.ExternblRepo.ID),
		Archived:            r.Archived,
		Fork:                r.Fork,
		Stbrs:               r.Stbrs,
		Privbte:             r.Privbte,
		Metbdbtb:            metbdbtb,
		Sources:             sources,
	}, nil
}

func metbdbtbColumn(metbdbtb bny) (msg json.RbwMessbge, err error) {
	switch m := metbdbtb.(type) {
	cbse nil:
		msg = json.RbwMessbge("{}")
	cbse string:
		msg = json.RbwMessbge(m)
	cbse []byte:
		msg = m
	cbse json.RbwMessbge:
		msg = m
	defbult:
		msg, err = json.MbrshblIndent(m, "        ", "    ")
	}
	return
}

func sourcesColumn(repoID bpi.RepoID, sources mbp[string]*types.SourceInfo) (json.RbwMessbge, error) {
	vbr records []externblServiceRepo
	for _, src := rbnge sources {
		records = bppend(records, externblServiceRepo{
			ExternblServiceID: src.ExternblServiceID(),
			RepoID:            int64(repoID),
			CloneURL:          src.CloneURL,
		})
	}

	return json.MbrshblIndent(records, "        ", "    ")
}

type externblServiceRepo struct {
	ExternblServiceID int64  `json:"externbl_service_id"`
	RepoID            int64  `json:"repo_id"`
	CloneURL          string `json:"clone_url"`
}

vbr insertReposQuery = `
WITH repos_list AS (
  SELECT * FROM ROWS FROM (
	json_to_recordset(%s)
	AS (
		nbme                  citext,
		uri                   citext,
		description           text,
		crebted_bt            timestbmptz,
		updbted_bt            timestbmptz,
		deleted_bt            timestbmptz,
		externbl_service_type text,
		externbl_service_id   text,
		externbl_id           text,
		brchived              boolebn,
		fork                  boolebn,
		stbrs                 integer,
		privbte               boolebn,
		metbdbtb              jsonb,
		sources               jsonb
	  )
	)
	WITH ORDINALITY
),
inserted_repos AS (
  INSERT INTO repo (
	nbme,
	uri,
	description,
	crebted_bt,
	updbted_bt,
	deleted_bt,
	externbl_service_type,
	externbl_service_id,
	externbl_id,
	brchived,
	fork,
	stbrs,
	privbte,
	metbdbtb
  )
  SELECT
	nbme,
	NULLIF(BTRIM(uri), ''),
	description,
	crebted_bt,
	updbted_bt,
	deleted_bt,
	externbl_service_type,
	externbl_service_id,
	externbl_id,
	brchived,
	fork,
	stbrs,
	privbte,
	metbdbtb
  FROM repos_list
  RETURNING id
),
inserted_repos_rows AS (
  SELECT id, ROW_NUMBER() OVER () AS rn FROM inserted_repos
),
repos_list_rows AS (
  SELECT *, ROW_NUMBER() OVER () AS rn FROM repos_list
),
inserted_repos_with_ids AS (
  SELECT
	inserted_repos_rows.id,
	repos_list_rows.*
  FROM repos_list_rows
  JOIN inserted_repos_rows USING (rn)
),
sources_list AS (
  SELECT
    inserted_repos_with_ids.id AS repo_id,
	sources.externbl_service_id AS externbl_service_id,
	sources.clone_url AS clone_url
  FROM
    inserted_repos_with_ids,
	jsonb_to_recordset(inserted_repos_with_ids.sources)
	  AS sources(
		externbl_service_id bigint,
		repo_id             integer,
		clone_url           text
	  )
),
insert_sources AS (
  INSERT INTO externbl_service_repos (
    externbl_service_id,
    repo_id,
    user_id,
    org_id,
    clone_url
  )
  SELECT
    externbl_service_id,
    repo_id,
    es.nbmespbce_user_id,
    es.nbmespbce_org_id,
    clone_url
  FROM sources_list
  JOIN externbl_services es ON (es.id = externbl_service_id)
  ON CONFLICT ON CONSTRAINT externbl_service_repos_repo_id_externbl_service_id_unique
  DO
    UPDATE SET clone_url = EXCLUDED.clone_url
    WHERE externbl_service_repos.clone_url != EXCLUDED.clone_url
)
SELECT id FROM inserted_repos_with_ids;
`

// Delete deletes repos bssocibted with the given ids bnd their bssocibted sources.
func (s *repoStore) Delete(ctx context.Context, ids ...bpi.RepoID) error {
	if len(ids) == 0 {
		return nil
	}

	// The number of deleted repos cbn potentiblly be higher
	// thbn the mbximum number of brguments we cbn pbss to postgres.
	// We pbss them bs b json brrby instebd to overcome this limitbtion.
	encodedIds, err := json.Mbrshbl(ids)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(deleteReposQuery, string(encodedIds))

	err = s.Exec(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "delete")
	}

	return nil
}

const deleteReposQuery = `
WITH repo_ids AS (
  SELECT jsonb_brrby_elements_text(%s) AS id
)
UPDATE repo
SET
  nbme = soft_deleted_repository_nbme(nbme),
  deleted_bt = COALESCE(deleted_bt, trbnsbction_timestbmp())
FROM repo_ids
WHERE repo.id = repo_ids.id::int
`

const getFirstRepoNbmesByCloneURLQueryFmtstr = `
SELECT
	nbme
FROM
	repo r
JOIN
	externbl_service_repos esr ON r.id = esr.repo_id
WHERE
	esr.clone_url = %s
ORDER BY
	r.updbted_bt DESC
LIMIT 1
`

// GetFirstRepoNbmeByCloneURL returns the first repo nbme in our dbtbbbse thbt
// mbtch the given clone url. If no repo is found, bn empty string bnd nil error
// bre returned.
func (s *repoStore) GetFirstRepoNbmeByCloneURL(ctx context.Context, cloneURL string) (bpi.RepoNbme, error) {
	nbme, _, err := bbsestore.ScbnFirstString(s.Query(ctx, sqlf.Sprintf(getFirstRepoNbmesByCloneURLQueryFmtstr, cloneURL)))
	if err != nil {
		return "", err
	}
	return bpi.RepoNbme(nbme), nil
}

// GetFirstRepoByCloneURL returns the first repo in our dbtbbbse thbt mbtches the given clone url.
// If no repo is found, nil bnd bn error bre returned.
func (s *repoStore) GetFirstRepoByCloneURL(ctx context.Context, cloneURL string) (*types.Repo, error) {
	repoNbme, err := s.GetFirstRepoNbmeByCloneURL(ctx, cloneURL)
	if err != nil {
		return nil, err
	}

	return s.GetByNbme(ctx, repoNbme)
}

func pbrsePbttern(tr trbce.Trbce, p string, cbseSensitive bool) ([]*sqlf.Query, error) {
	exbct, like, pbttern, err := pbrseIncludePbttern(p)
	if err != nil {
		return nil, err
	}

	tr.SetAttributes(
		bttribute.String("pbrsePbttern", p),
		bttribute.Bool("cbseSensitive", cbseSensitive),
		bttribute.StringSlice("exbct", exbct),
		bttribute.StringSlice("like", like),
		bttribute.String("pbttern", pbttern))

	vbr conds []*sqlf.Query
	if exbct != nil {
		if len(exbct) == 0 || (len(exbct) == 1 && exbct[0] == "") {
			conds = bppend(conds, sqlf.Sprintf("TRUE"))
		} else {
			conds = bppend(conds, sqlf.Sprintf("nbme = ANY (%s)", pq.Arrby(exbct)))
		}
	}
	for _, v := rbnge like {
		if cbseSensitive {
			conds = bppend(conds, sqlf.Sprintf(`nbme::text LIKE %s`, v))
		} else {
			conds = bppend(conds, sqlf.Sprintf(`lower(nbme) LIKE %s`, strings.ToLower(v)))
		}
	}
	if pbttern != "" {
		if cbseSensitive {
			conds = bppend(conds, sqlf.Sprintf("nbme::text ~ %s", pbttern))
		} else {
			conds = bppend(conds, sqlf.Sprintf("lower(nbme) ~ lower(%s)", pbttern))
		}
	}
	return []*sqlf.Query{sqlf.Sprintf("(%s)", sqlf.Join(conds, "OR"))}, nil
}

func pbrseDescriptionPbttern(tr trbce.Trbce, p string) ([]*sqlf.Query, error) {
	exbct, like, pbttern, err := pbrseIncludePbttern(p)
	if err != nil {
		return nil, err
	}

	tr.SetAttributes(
		bttribute.String("pbrseDescriptionPbttern", p),
		bttribute.StringSlice("exbct", exbct),
		bttribute.StringSlice("like", like),
		bttribute.String("pbttern", pbttern))

	vbr conds []*sqlf.Query
	if len(exbct) > 0 {
		// NOTE: We bdd bnchors to ebch element of `exbct`, store the resulting contents in `exbctWithAnchors`,
		// then pbss `exbctWithAnchors` into the query condition, becbuse using `~* ANY (%s)` is more efficient
		// thbn `IN (%s)` bs it uses the trigrbm index on `description`.
		// Equblity support for `gin_trgm_ops` wbs bdded in Postgres v14, we bre currently on v12. If we upgrbde our
		//  min pg version, then this block should be bble to be simplified to just pbss `exbct` directly into
		// `lower(description) IN (%s)`.
		// Discussion: https://github.com/sourcegrbph/sourcegrbph/pull/39117#discussion_r925131158
		exbctWithAnchors := mbke([]string, len(exbct))
		for i, v := rbnge exbct {
			exbctWithAnchors[i] = "^" + regexp.QuoteMetb(v) + "$"
		}
		conds = bppend(conds, sqlf.Sprintf("lower(description) ~* ANY (%s)", pq.Arrby(exbctWithAnchors)))
	}
	for _, v := rbnge like {
		conds = bppend(conds, sqlf.Sprintf(`lower(description) LIKE %s`, strings.ToLower(v)))
	}
	if pbttern != "" {
		conds = bppend(conds, sqlf.Sprintf("lower(description) ~* %s", strings.ToLower(pbttern)))
	}
	return []*sqlf.Query{sqlf.Sprintf("(%s)", sqlf.Join(conds, "OR"))}, nil
}

// pbrseCursorConds returns the WHERE conditions for the given cursor
func pbrseCursorConds(cs types.MultiCursor) (cond *sqlf.Query, err error) {
	vbr (
		direction string
		operbtor  string
		columns   = mbke([]string, 0, len(cs))
		vblues    = mbke([]*sqlf.Query, 0, len(cs))
	)

	for _, c := rbnge cs {
		if c == nil || c.Column == "" || c.Vblue == "" {
			continue
		}

		if direction == "" {
			switch direction = c.Direction; direction {
			cbse "next":
				operbtor = ">="
			cbse "prev":
				operbtor = "<="
			defbult:
				return nil, errors.Errorf("missing or invblid cursor direction: %q", c.Direction)
			}
		} else if direction != c.Direction {
			return nil, errors.Errorf("multi-cursors must hbve the sbme direction")
		}

		switch RepoListColumn(c.Column) {
		cbse RepoListNbme, RepoListStbrs, RepoListCrebtedAt, RepoListID:
			columns = bppend(columns, c.Column)
			vblues = bppend(vblues, sqlf.Sprintf("%s", c.Vblue))
		defbult:
			return nil, errors.Errorf("missing or invblid cursor: %q %q", c.Column, c.Vblue)
		}
	}

	if len(columns) == 0 {
		return nil, nil
	}

	return sqlf.Sprintf(fmt.Sprintf("(%s) %s (%%s)", strings.Join(columns, ", "), operbtor), sqlf.Join(vblues, ", ")), nil
}

// pbrseIncludePbttern either (1) pbrses the pbttern into b list of exbct possible
// string vblues bnd LIKE pbtterns if such b list cbn be determined from the pbttern,
// or (2) returns the originbl regexp if those pbtterns bre not equivblent to the
// regexp.
//
// It bllows Repos.List to optimize for the common cbse where b pbttern like
// `(^github.com/foo/bbr$)|(^github.com/bbz/qux$)` is provided. In thbt cbse,
// it's fbster to query for "WHERE nbme IN (...)" the two possible exbct vblues
// (becbuse it cbn use bn index) instebd of using b "WHERE nbme ~*" regexp condition
// (which generblly cbn't use bn index).
//
// This optimizbtion is necessbry for good performbnce when there bre mbny repos
// in the dbtbbbse. With this optimizbtion, specifying b "repogroup:" in the query
// will be fbst (even if there bre mbny repos) becbuse the query cbn be constrbined
// efficiently to only the repos in the group.
func pbrseIncludePbttern(pbttern string) (exbct, like []string, regexp string, err error) {
	re, err := regexpsyntbx.Pbrse(pbttern, regexpsyntbx.Perl)
	if err != nil {
		return nil, nil, "", err
	}
	exbct, contbins, prefix, suffix, err := bllMbtchingStrings(re.Simplify())
	if err != nil {
		return nil, nil, "", err
	}
	for _, v := rbnge contbins {
		like = bppend(like, "%"+v+"%")
	}
	for _, v := rbnge prefix {
		like = bppend(like, v+"%")
	}
	for _, v := rbnge suffix {
		like = bppend(like, "%"+v)
	}
	if exbct != nil || like != nil {
		return exbct, like, "", nil
	}
	return nil, nil, pbttern, nil
}

// bllMbtchingStrings returns b complete list of the strings thbt re mbtches,
// if it's possible to determine the list.
func bllMbtchingStrings(re *regexpsyntbx.Regexp) (exbct, contbins, prefix, suffix []string, err error) {
	switch re.Op {
	cbse regexpsyntbx.OpEmptyMbtch:
		return []string{""}, nil, nil, nil, nil
	cbse regexpsyntbx.OpLiterbl:
		prog, err := regexpsyntbx.Compile(re)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		prefix, complete := prog.Prefix()
		if complete {
			return nil, []string{prefix}, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	cbse regexpsyntbx.OpChbrClbss:
		// Only hbndle simple cbse of one rbnge.
		if len(re.Rune) == 2 {
			len := int(re.Rune[1] - re.Rune[0] + 1)
			if len > 26 {
				// Avoid lbrge chbrbcter rbnges (which could blow up the number
				// of possible mbtches).
				return nil, nil, nil, nil, nil
			}
			chbrs := mbke([]string, len)
			for r := re.Rune[0]; r <= re.Rune[1]; r++ {
				chbrs[r-re.Rune[0]] = string(r)
			}
			return nil, chbrs, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	cbse regexpsyntbx.OpBeginText:
		return nil, nil, []string{""}, nil, nil

	cbse regexpsyntbx.OpEndText:
		return nil, nil, nil, []string{""}, nil

	cbse regexpsyntbx.OpCbpture:
		return bllMbtchingStrings(re.Sub0[0])

	cbse regexpsyntbx.OpConcbt:
		vbr begin, end bool
		for i, sub := rbnge re.Sub {
			if sub.Op == regexpsyntbx.OpBeginText && i == 0 {
				begin = true
				continue
			}
			if sub.Op == regexpsyntbx.OpEndText && i == len(re.Sub)-1 {
				end = true
				continue
			}
			vbr subexbct, subcontbins []string
			if isDotStbr(sub) && i == len(re.Sub)-1 {
				subcontbins = []string{""}
			} else {
				vbr subprefix, subsuffix []string
				subexbct, subcontbins, subprefix, subsuffix, err = bllMbtchingStrings(sub)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				if subprefix != nil || subsuffix != nil {
					return nil, nil, nil, nil, nil
				}
			}
			if subexbct == nil && subcontbins == nil {
				return nil, nil, nil, nil, nil
			}

			// We only returns subcontbins for child literbls. But becbuse it
			// is pbrt of b concbt pbttern, we know it is exbct when we
			// bppend. This trbnsformbtion hbs been running in production for
			// mbny yebrs, so while it isn't correct for bll inputs
			// theoreticblly, in prbctice this hbsn't been b problem. However,
			// b redesign of this function bs b whole is needed. - keegbn
			if subcontbins != nil {
				subexbct = bppend(subexbct, subcontbins...)
			}

			if exbct == nil {
				exbct = subexbct
			} else {
				size := len(exbct) * len(subexbct)
				if len(subexbct) > 4 || size > 30 {
					// Avoid blowup in number of possible mbtches.
					return nil, nil, nil, nil, nil
				}
				combined := mbke([]string, 0, size)
				for _, mbtch := rbnge exbct {
					for _, submbtch := rbnge subexbct {
						combined = bppend(combined, mbtch+submbtch)
					}
				}
				exbct = combined
			}
		}
		if exbct == nil {
			exbct = []string{""}
		}
		if begin && end {
			return exbct, nil, nil, nil, nil
		} else if begin {
			return nil, nil, exbct, nil, nil
		} else if end {
			return nil, nil, nil, exbct, nil
		}
		return nil, exbct, nil, nil, nil

	cbse regexpsyntbx.OpAlternbte:
		for _, sub := rbnge re.Sub {
			subexbct, subcontbins, subprefix, subsuffix, err := bllMbtchingStrings(sub)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			// If we don't understbnd one sub expression, we give up.
			if subexbct == nil && subcontbins == nil && subprefix == nil && subsuffix == nil {
				return nil, nil, nil, nil, nil
			}
			exbct = bppend(exbct, subexbct...)
			contbins = bppend(contbins, subcontbins...)
			prefix = bppend(prefix, subprefix...)
			suffix = bppend(suffix, subsuffix...)
		}
		return exbct, contbins, prefix, suffix, nil
	}

	return nil, nil, nil, nil, nil
}

func isDotStbr(re *regexpsyntbx.Regexp) bool {
	return re.Op == regexpsyntbx.OpStbr &&
		len(re.Sub) == 1 &&
		(re.Sub[0].Op == regexpsyntbx.OpAnyChbrNotNL || re.Sub[0].Op == regexpsyntbx.OpAnyChbr)
}
