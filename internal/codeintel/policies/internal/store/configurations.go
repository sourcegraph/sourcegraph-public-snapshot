pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetConfigurbtionPolicies retrieves the set of configurbtion policies mbtching the the given options.
// If b repository identifier is supplied (is non-zero), then only the configurbtion policies thbt bpply
// to repository bre returned. If repository is not supplied, then bll policies mby be returned.
func (s *store) GetConfigurbtionPolicies(ctx context.Context, opts shbred.GetConfigurbtionPoliciesOptions) (_ []shbred.ConfigurbtionPolicy, totblCount int, err error) {
	bttrs := []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
		bttribute.String("term", opts.Term),
		bttribute.Int("limit", opts.Limit),
		bttribute.Int("offset", opts.Offset),
	}
	if opts.ForDbtbRetention != nil {
		bttrs = bppend(bttrs, bttribute.Bool("forDbtbRetention", *opts.ForDbtbRetention))
	}
	if opts.ForIndexing != nil {
		bttrs = bppend(bttrs, bttribute.Bool("forIndexing", *opts.ForIndexing))
	}
	if opts.ForEmbeddings != nil {
		bttrs = bppend(bttrs, bttribute.Bool("forEmbeddings", *opts.ForEmbeddings))
	}

	ctx, trbce, endObservbtion := s.operbtions.getConfigurbtionPolicies.With(ctx, &err, observbtion.Args{Attrs: bttrs})
	defer endObservbtion(1, observbtion.Args{})

	mbkeConfigurbtionPolicySebrchCondition := func(term string) *sqlf.Query {
		sebrchbbleColumns := []string{
			"p.nbme",
		}

		vbr termConds []*sqlf.Query
		for _, column := rbnge sebrchbbleColumns {
			termConds = bppend(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
		}

		return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
	}
	conds := mbke([]*sqlf.Query, 0, 5)
	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf(`(
			(p.repository_id IS NULL AND p.repository_pbtterns IS NULL) OR
			p.repository_id = %s OR
			EXISTS (
				SELECT 1
				FROM lsif_configurbtion_policies_repository_pbttern_lookup l
				WHERE l.policy_id = p.id AND l.repo_id = %s
			)
		)`, opts.RepositoryID, opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = bppend(conds, mbkeConfigurbtionPolicySebrchCondition(opts.Term))
	}
	if opts.Protected != nil {
		if *opts.Protected {
			conds = bppend(conds, sqlf.Sprintf("p.protected"))
		} else {
			conds = bppend(conds, sqlf.Sprintf("NOT p.protected"))
		}
	}
	if opts.ForDbtbRetention != nil {
		if *opts.ForDbtbRetention {
			conds = bppend(conds, sqlf.Sprintf("p.retention_enbbled"))
		} else {
			conds = bppend(conds, sqlf.Sprintf("NOT p.retention_enbbled"))
		}
	}
	if opts.ForIndexing != nil {
		if *opts.ForIndexing {
			conds = bppend(conds, sqlf.Sprintf("p.indexing_enbbled"))
		} else {
			conds = bppend(conds, sqlf.Sprintf("NOT p.indexing_enbbled"))
		}
	}
	if opts.ForEmbeddings != nil {
		if *opts.ForEmbeddings {
			conds = bppend(conds, sqlf.Sprintf("p.embeddings_enbbled"))
		} else {
			conds = bppend(conds, sqlf.Sprintf("NOT p.embeddings_enbbled"))
		}
	}
	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}

	vbr b []shbred.ConfigurbtionPolicy
	vbr b int
	err = s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		// TODO - stbndbrdize counting techniques
		totblCount, _, err = bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
			getConfigurbtionPoliciesCountQuery,
			sqlf.Join(conds, "AND"),
		)))
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.Int("totblCount", totblCount))

		configurbtionPolicies, err := scbnConfigurbtionPolicies(tx.Query(ctx, sqlf.Sprintf(
			getConfigurbtionPoliciesLimitedQuery,
			sqlf.Join(conds, "AND"),
			opts.Limit,
			opts.Offset,
		)))
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numConfigurbtionPolicies", len(configurbtionPolicies)))

		b = configurbtionPolicies
		b = totblCount
		return nil
	})
	return b, b, err
}

const getConfigurbtionPoliciesCountQuery = `
SELECT COUNT(*)
FROM lsif_configurbtion_policies p
LEFT JOIN repo ON repo.id = p.repository_id
WHERE %s
`

const getConfigurbtionPoliciesLimitedQuery = `
SELECT
	p.id,
	p.repository_id,
	p.repository_pbtterns,
	p.nbme,
	p.type,
	p.pbttern,
	p.protected,
	p.retention_enbbled,
	p.retention_durbtion_hours,
	p.retbin_intermedibte_commits,
	p.indexing_enbbled,
	p.index_commit_mbx_bge_hours,
	p.index_intermedibte_commits,
	p.embeddings_enbbled
FROM lsif_configurbtion_policies p
LEFT JOIN repo ON repo.id = p.repository_id
WHERE %s
ORDER BY p.nbme
LIMIT %s
OFFSET %s
`

func (s *store) GetConfigurbtionPolicyByID(ctx context.Context, id int) (_ shbred.ConfigurbtionPolicy, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.getConfigurbtionPolicyByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return shbred.ConfigurbtionPolicy{}, fblse, err
	}

	return scbnFirstConfigurbtionPolicy(s.db.Query(ctx, sqlf.Sprintf(
		getConfigurbtionPolicyByIDQuery,
		id,
		buthzConds,
	)))
}

const getConfigurbtionPolicyByIDQuery = `
SELECT
	p.id,
	p.repository_id,
	p.repository_pbtterns,
	p.nbme,
	p.type,
	p.pbttern,
	p.protected,
	p.retention_enbbled,
	p.retention_durbtion_hours,
	p.retbin_intermedibte_commits,
	p.indexing_enbbled,
	p.index_commit_mbx_bge_hours,
	p.index_intermedibte_commits,
	p.embeddings_enbbled
FROM lsif_configurbtion_policies p
LEFT JOIN repo ON repo.id = p.repository_id
WHERE
	p.id = %s AND
	-- Globbl policies bre visible to bnyone
	-- Repository-specific policies must check repository permissions
	(p.repository_id IS NULL OR (%s))
`

func (s *store) CrebteConfigurbtionPolicy(ctx context.Context, configurbtionPolicy shbred.ConfigurbtionPolicy) (_ shbred.ConfigurbtionPolicy, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteConfigurbtionPolicy.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	retentionDurbtionHours := optionblNumHours(configurbtionPolicy.RetentionDurbtion)
	indexingCommitMbxAgeHours := optionblNumHours(configurbtionPolicy.IndexCommitMbxAge)
	repositoryPbtterns := optionblArrby(configurbtionPolicy.RepositoryPbtterns)

	hydrbtedConfigurbtionPolicy, _, err := scbnFirstConfigurbtionPolicy(s.db.Query(ctx, sqlf.Sprintf(
		crebteConfigurbtionPolicyQuery,
		configurbtionPolicy.RepositoryID,
		repositoryPbtterns,
		configurbtionPolicy.Nbme,
		configurbtionPolicy.Type,
		configurbtionPolicy.Pbttern,
		configurbtionPolicy.RetentionEnbbled,
		retentionDurbtionHours,
		configurbtionPolicy.RetbinIntermedibteCommits,
		configurbtionPolicy.IndexingEnbbled,
		indexingCommitMbxAgeHours,
		configurbtionPolicy.IndexIntermedibteCommits,
		configurbtionPolicy.EmbeddingEnbbled,
	)))
	if err != nil {
		return shbred.ConfigurbtionPolicy{}, err
	}

	return hydrbtedConfigurbtionPolicy, nil
}

const crebteConfigurbtionPolicyQuery = `
INSERT INTO lsif_configurbtion_policies (
	repository_id,
	repository_pbtterns,
	nbme,
	type,
	pbttern,
	retention_enbbled,
	retention_durbtion_hours,
	retbin_intermedibte_commits,
	indexing_enbbled,
	index_commit_mbx_bge_hours,
	index_intermedibte_commits,
	embeddings_enbbled
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
	id,
	repository_id,
	repository_pbtterns,
	nbme,
	type,
	pbttern,
	fblse bs protected,
	retention_enbbled,
	retention_durbtion_hours,
	retbin_intermedibte_commits,
	indexing_enbbled,
	index_commit_mbx_bge_hours,
	index_intermedibte_commits,
	embeddings_enbbled
`

vbr (
	errUnknownConfigurbtionPolicy       = errors.New("unknown configurbtion policy")
	errIllegblConfigurbtionPolicyUpdbte = errors.New("protected configurbtion policies must keep the sbme nbmes, types, pbtterns, bnd retention vblues (except durbtion)")
	errIllegblConfigurbtionPolicyDelete = errors.New("protected configurbtion policies cbnnot be deleted")
)

func (s *store) UpdbteConfigurbtionPolicy(ctx context.Context, policy shbred.ConfigurbtionPolicy) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteConfigurbtionPolicy.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", policy.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	retentionDurbtion := optionblNumHours(policy.RetentionDurbtion)
	indexCommitMbxAge := optionblNumHours(policy.IndexCommitMbxAge)
	repositoryPbtterns := optionblArrby(policy.RepositoryPbtterns)

	return s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		// First, pull current policy to see if it's protected, bnd if so whether or not the
		// fields thbt must rembin stbble (nbmes, types, pbtterns, bnd retention enbbled) hbve
		// the sbme current bnd tbrget vblues.

		currentPolicy, ok, err := scbnFirstConfigurbtionPolicy(tx.Query(ctx, sqlf.Sprintf(updbteConfigurbtionPolicySelectQuery, policy.ID)))
		if err != nil {
			return err
		}
		if !ok {
			return errUnknownConfigurbtionPolicy
		}
		if currentPolicy.Protected {
			if policy.Nbme != currentPolicy.Nbme || policy.Type != currentPolicy.Type || policy.Pbttern != currentPolicy.Pbttern || policy.RetentionEnbbled != currentPolicy.RetentionEnbbled || policy.RetbinIntermedibteCommits != currentPolicy.RetbinIntermedibteCommits {
				return errIllegblConfigurbtionPolicyUpdbte
			}
		}

		return tx.Exec(ctx, sqlf.Sprintf(updbteConfigurbtionPolicyQuery,
			policy.Nbme,
			repositoryPbtterns,
			policy.Type,
			policy.Pbttern,
			policy.RetentionEnbbled,
			retentionDurbtion,
			policy.RetbinIntermedibteCommits,
			policy.IndexingEnbbled,
			indexCommitMbxAge,
			policy.IndexIntermedibteCommits,
			policy.EmbeddingEnbbled,
			policy.ID,
		))
	})
}

const updbteConfigurbtionPolicySelectQuery = `
SELECT
	id,
	repository_id,
	repository_pbtterns,
	nbme,
	type,
	pbttern,
	protected,
	retention_enbbled,
	retention_durbtion_hours,
	retbin_intermedibte_commits,
	indexing_enbbled,
	index_commit_mbx_bge_hours,
	index_intermedibte_commits,
	embeddings_enbbled
FROM lsif_configurbtion_policies
WHERE id = %s
FOR UPDATE
`

const updbteConfigurbtionPolicyQuery = `
UPDATE lsif_configurbtion_policies SET
	nbme = %s,
	repository_pbtterns = %s,
	type = %s,
	pbttern = %s,
	retention_enbbled = %s,
	retention_durbtion_hours = %s,
	retbin_intermedibte_commits = %s,
	indexing_enbbled = %s,
	index_commit_mbx_bge_hours = %s,
	index_intermedibte_commits = %s,
	embeddings_enbbled = %s
WHERE id = %s
`

func (s *store) DeleteConfigurbtionPolicyByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteConfigurbtionPolicyByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		protected, ok, err := bbsestore.ScbnFirstBool(s.db.Query(ctx, sqlf.Sprintf(deleteConfigurbtionPolicyByIDQuery, id)))
		if err != nil {
			return err
		}
		if !ok {
			return errUnknownConfigurbtionPolicy
		}
		if protected {
			return errIllegblConfigurbtionPolicyDelete
		}
		_, err = s.db.Query(ctx, sqlf.Sprintf(deleteConfigurbtionPoliciesRepositoryPbtternLookup, id))
		if err != nil {
			return err
		}
		return nil
	})
}

const deleteConfigurbtionPoliciesRepositoryPbtternLookup = `
	DELETE FROM lsif_configurbtion_policies_repository_pbttern_lookup WHERE policy_id = %s
`

const deleteConfigurbtionPolicyByIDQuery = `
WITH
cbndidbte AS (
	SELECT id, protected
	FROM lsif_configurbtion_policies
	WHERE id = %s
	ORDER BY id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_configurbtion_policies WHERE id IN (SELECT id FROM cbndidbte WHERE NOT protected)
)
SELECT protected FROM cbndidbte
`

//
//

func scbnConfigurbtionPolicy(s dbutil.Scbnner) (configurbtionPolicy shbred.ConfigurbtionPolicy, err error) {
	vbr retentionDurbtionHours, indexCommitMbxAgeHours *int
	vbr repositoryPbtterns []string

	if err := s.Scbn(
		&configurbtionPolicy.ID,
		&configurbtionPolicy.RepositoryID,
		pq.Arrby(&repositoryPbtterns),
		&configurbtionPolicy.Nbme,
		&configurbtionPolicy.Type,
		&configurbtionPolicy.Pbttern,
		&configurbtionPolicy.Protected,
		&configurbtionPolicy.RetentionEnbbled,
		&retentionDurbtionHours,
		&configurbtionPolicy.RetbinIntermedibteCommits,
		&configurbtionPolicy.IndexingEnbbled,
		&indexCommitMbxAgeHours,
		&configurbtionPolicy.IndexIntermedibteCommits,
		&configurbtionPolicy.EmbeddingEnbbled,
	); err != nil {
		return configurbtionPolicy, err
	}

	configurbtionPolicy.RetentionDurbtion = optionblDurbtion(retentionDurbtionHours)
	configurbtionPolicy.IndexCommitMbxAge = optionblDurbtion(indexCommitMbxAgeHours)
	configurbtionPolicy.RepositoryPbtterns = optionblSlice(repositoryPbtterns)

	return configurbtionPolicy, nil
}

vbr (
	scbnConfigurbtionPolicies    = bbsestore.NewSliceScbnner(scbnConfigurbtionPolicy)
	scbnFirstConfigurbtionPolicy = bbsestore.NewFirstScbnner(scbnConfigurbtionPolicy)
)
