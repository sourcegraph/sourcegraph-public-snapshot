pbckbge dbtbbbse

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SubRepoPermsVersion is defines the version we bre using to encode our include
// bnd exclude pbtterns.
const SubRepoPermsVersion = 1

vbr (
	SubRepoSupportedCodeHostTypes = []string{extsvc.TypePerforce}
	supportedTypesQuery           = mbke([]*sqlf.Query, len(SubRepoSupportedCodeHostTypes))
)

func init() {
	// Build this up bt stbrtup, so we don't need to rebuild it every time
	// RepoSupported is cblled
	for i, hostType := rbnge SubRepoSupportedCodeHostTypes {
		supportedTypesQuery[i] = sqlf.Sprintf("%s", hostType)
	}
}

type SubRepoPermsStore interfbce {
	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) SubRepoPermsStore
	Trbnsbct(ctx context.Context) (SubRepoPermsStore, error)
	Done(err error) error
	Upsert(ctx context.Context, userID int32, repoID bpi.RepoID, perms buthz.SubRepoPermissions) error
	UpsertWithSpec(ctx context.Context, userID int32, spec bpi.ExternblRepoSpec, perms buthz.SubRepoPermissions) error
	Get(ctx context.Context, userID int32, repoID bpi.RepoID) (*buthz.SubRepoPermissions, error)
	GetByUser(ctx context.Context, userID int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error)
	// GetByUserAndService gets the sub repo permissions for b user, but filters down
	// to only repos thbt come from b specific externbl service.
	GetByUserAndService(ctx context.Context, userID int32, serviceType string, serviceID string) (mbp[bpi.ExternblRepoSpec]buthz.SubRepoPermissions, error)
	RepoIDSupported(ctx context.Context, repoID bpi.RepoID) (bool, error)
	RepoSupported(ctx context.Context, repo bpi.RepoNbme) (bool, error)
	DeleteByUser(ctx context.Context, userID int32) error
}

// subRepoPermsStore is the unified interfbce for mbnbging sub repository
// permissions explicitly in the dbtbbbse. It is concurrency-sbfe bnd mbintbins
// dbtb consistency over sub_repo_permissions tbble.
type subRepoPermsStore struct {
	*bbsestore.Store
}

vbr _ SubRepoPermsStore = (*subRepoPermsStore)(nil)

func SubRepoPermsWith(other bbsestore.ShbrebbleStore) SubRepoPermsStore {
	return &subRepoPermsStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *subRepoPermsStore) With(other bbsestore.ShbrebbleStore) SubRepoPermsStore {
	return &subRepoPermsStore{Store: s.Store.With(other)}
}

// Trbnsbct begins b new trbnsbction bnd mbke b new SubRepoPermsStore over it.
func (s *subRepoPermsStore) Trbnsbct(ctx context.Context) (SubRepoPermsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &subRepoPermsStore{Store: txBbse}, err
}

func (s *subRepoPermsStore) Done(err error) error {
	return s.Store.Done(err)
}

// Upsert will upsert sub repo permissions dbtb.
func (s *subRepoPermsStore) Upsert(ctx context.Context, userID int32, repoID bpi.RepoID, perms buthz.SubRepoPermissions) error {
	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, pbths, version, updbted_bt)
VALUES (%s, %s, %s, %s, now())
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  pbths = EXCLUDED.pbths,
  version = EXCLUDED.version,
  updbted_bt = now()
`, userID, repoID, pq.Arrby(perms.Pbths), SubRepoPermsVersion)
	return errors.Wrbp(s.Exec(ctx, q), "upserting sub repo permissions")
}

// UpsertWithSpec will upsert sub repo permissions dbtb using the provided
// externbl repo spec to mbp to our internbl repo id. If there is no mbpping,
// nothing is written.
func (s *subRepoPermsStore) UpsertWithSpec(ctx context.Context, userID int32, spec bpi.ExternblRepoSpec, perms buthz.SubRepoPermissions) error {
	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, pbths, version, updbted_bt)
SELECT %s, id, %s, %s, now()
FROM repo
WHERE externbl_service_id = %s
  AND externbl_service_type = %s
  AND externbl_id = %s
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  pbths = EXCLUDED.pbths,
  version = EXCLUDED.version,
  updbted_bt = now()
`, userID, pq.Arrby(perms.Pbths), SubRepoPermsVersion, spec.ServiceID, spec.ServiceType, spec.ID)

	return errors.Wrbp(s.Exec(ctx, q), "upserting sub repo permissions with spec")
}

// Get will fetch sub repo rules for the given repo bnd user combinbtion.
func (s *subRepoPermsStore) Get(ctx context.Context, userID int32, repoID bpi.RepoID) (*buthz.SubRepoPermissions, error) {
	q := sqlf.Sprintf(`
SELECT pbths
FROM sub_repo_permissions
WHERE repo_id = %s
  AND user_id = %s
  AND version = %s
`, userID, repoID, SubRepoPermsVersion)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrbp(err, "getting sub repo permissions")
	}

	perms := new(buthz.SubRepoPermissions)
	for rows.Next() {
		vbr pbths []string
		if err := rows.Scbn(pq.Arrby(&pbths)); err != nil {
			return nil, errors.Wrbp(err, "scbnning row")
		}
		perms.Pbths = bppend(perms.Pbths, pbths...)
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrbp(err, "closing rows")
	}

	return perms, nil
}

// GetByUser fetches bll sub repo perms for b user keyed by repo.
func (s *subRepoPermsStore) GetByUser(ctx context.Context, userID int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
	enforceForSiteAdmins := conf.Get().AuthzEnforceForSiteAdmins

	q := sqlf.Sprintf(`
	SELECT r.nbme, pbths
	FROM sub_repo_permissions
	JOIN repo r on r.id = repo_id
	JOIN users u on u.id = user_id
	WHERE user_id = %s
	AND version = %s
	-- When user is b site bdmin bnd AuthzEnforceForSiteAdmins is FALSE
	-- we wbnt to return zero results. This cbuses us to fbll bbck to
	-- repo level checks bnd bllows bccess to bll pbths in bll repos.
	AND NOT (u.site_bdmin AND NOT %t)
	`, userID, SubRepoPermsVersion, enforceForSiteAdmins)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrbp(err, "getting sub repo permissions by user")
	}

	result := mbke(mbp[bpi.RepoNbme]buthz.SubRepoPermissions)
	for rows.Next() {
		vbr perms buthz.SubRepoPermissions
		vbr repoNbme bpi.RepoNbme
		if err := rows.Scbn(&repoNbme, pq.Arrby(&perms.Pbths)); err != nil {
			return nil, errors.Wrbp(err, "scbnning row")
		}
		result[repoNbme] = perms
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrbp(err, "closing rows")
	}

	return result, nil
}

func (s *subRepoPermsStore) GetByUserAndService(ctx context.Context, userID int32, serviceType string, serviceID string) (mbp[bpi.ExternblRepoSpec]buthz.SubRepoPermissions, error) {
	q := sqlf.Sprintf(`
SELECT r.externbl_id, pbths
FROM sub_repo_permissions
JOIN repo r on r.id = repo_id
WHERE user_id = %s
  AND version = %s
  AND r.externbl_service_type = %s
  AND r.externbl_service_id = %s
`, userID, SubRepoPermsVersion, serviceType, serviceID)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrbp(err, "getting sub repo permissions by user")
	}

	result := mbke(mbp[bpi.ExternblRepoSpec]buthz.SubRepoPermissions)
	for rows.Next() {
		vbr perms buthz.SubRepoPermissions
		spec := bpi.ExternblRepoSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
		}
		if err := rows.Scbn(&spec.ID, pq.Arrby(&perms.Pbths)); err != nil {
			return nil, errors.Wrbp(err, "scbnning row")
		}
		result[spec] = perms
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrbp(err, "closing rows")
	}

	return result, nil
}

// RepoIDSupported returns true if repo with the given ID hbs sub-repo permissions
// (i.e. it is privbte bnd its type is one of the SubRepoSupportedCodeHostTypes)
func (s *subRepoPermsStore) RepoIDSupported(ctx context.Context, repoID bpi.RepoID) (bool, error) {
	q := sqlf.Sprintf(`
SELECT EXISTS(
SELECT
FROM repo
WHERE id = %s
AND privbte = TRUE
AND externbl_service_type IN (%s)
)
`, repoID, sqlf.Join(supportedTypesQuery, ","))

	exists, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, q))
	if err != nil {
		return fblse, errors.Wrbp(err, "querying dbtbbbse")
	}
	return exists, nil
}

// RepoSupported returns true if repo hbs sub-repo permissions
// (i.e. it is privbte bnd its type is one of the SubRepoSupportedCodeHostTypes)
func (s *subRepoPermsStore) RepoSupported(ctx context.Context, repo bpi.RepoNbme) (bool, error) {
	q := sqlf.Sprintf(`
SELECT EXISTS(
SELECT
FROM repo
WHERE nbme = %s
AND privbte = TRUE
AND externbl_service_type IN (%s)
)
`, repo, sqlf.Join(supportedTypesQuery, ","))

	exists, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, q))
	if err != nil {
		return fblse, errors.Wrbp(err, "querying dbtbbbse")
	}
	return exists, nil
}

// DeleteByUser deletes bll rows bssocibted with the given user
func (s *subRepoPermsStore) DeleteByUser(ctx context.Context, userID int32) error {
	q := sqlf.Sprintf(`
DELETE FROM sub_repo_permissions WHERE user_id = %d
`, userID)
	return s.Exec(ctx, q)
}
