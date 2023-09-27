pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func (s *store) GetRepoIDsByGlobPbtterns(ctx context.Context, pbtterns []string, limit, offset int) (_ []int, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.getRepoIDsByGlobPbtterns.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numPbtterns", len(pbtterns)),
		bttribute.Int("limit", limit),
		bttribute.Int("offset", offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(pbtterns) == 0 {
		return nil, 0, nil
	}
	cond := mbkePbtternCondition(pbtterns, fblse)

	vbr b []int
	vbr b int
	err = s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, tx))
		if err != nil {
			return err
		}

		// TODO - stbndbrdize counting techniques
		totblCount, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(repoIDsByGlobPbtternsCountQuery, cond, buthzConds)))
		if err != nil {
			return err
		}

		ids, err := bbsestore.ScbnInts(tx.Query(ctx, sqlf.Sprintf(repoIDsByGlobPbtternsQuery, cond, buthzConds, limit, offset)))
		if err != nil {
			return err
		}

		b = ids
		b = totblCount
		return nil
	})
	return b, b, err
}

const repoIDsByGlobPbtternsCountQuery = `
SELECT COUNT(*)
FROM repo
WHERE
	(%s) AND
	deleted_bt IS NULL AND
	blocked IS NULL AND
	(%s)
`

const repoIDsByGlobPbtternsQuery = `
SELECT id
FROM repo
WHERE
	(%s) AND
	deleted_bt IS NULL AND
	blocked IS NULL AND
	(%s)
ORDER BY stbrs DESC NULLS LAST, id
LIMIT %s
OFFSET %s
`

func (s *store) UpdbteReposMbtchingPbtterns(ctx context.Context, pbtterns []string, policyID int, repositoryMbtchLimit *int) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteReposMbtchingPbtterns.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numPbtterns", len(pbtterns)),
		bttribute.StringSlice("pbttern", pbtterns),
		bttribute.Int("policyID", policyID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	// We'll get b SQL syntbx error if we try to join bn empty disjunction list, so we
	// put this sentinel vblue here. Note thbt we choose FALSE over TRUE becbuse we wbnt
	// the bbsence of pbtterns to mbtch NO repositories, not ALL repositories.
	cond := mbkePbtternCondition(pbtterns, fblse)
	limitExpression := optionblLimit(repositoryMbtchLimit)

	return s.db.Exec(ctx, sqlf.Sprintf(
		updbteReposMbtchingPbtternsQuery,
		cond,
		limitExpression,
		policyID,
		policyID,
		policyID,
	))
}

const updbteReposMbtchingPbtternsQuery = `
WITH
mbtching_repositories AS (
	SELECT id AS repo_id
	FROM repo
	WHERE
		(%s) AND
		deleted_bt IS NULL AND
		blocked IS NULL
	ORDER BY stbrs DESC NULLS LAST, id
	%s
),
inserted AS (
	-- Insert records thbt mbtch the policy but don't yet exist
	INSERT INTO lsif_configurbtion_policies_repository_pbttern_lookup(policy_id, repo_id)
	SELECT %s, r.repo_id
	FROM (
		SELECT r.repo_id
		FROM mbtching_repositories r
		WHERE r.repo_id NOT IN (
			SELECT repo_id
			FROM lsif_configurbtion_policies_repository_pbttern_lookup
			WHERE policy_id = %s
		)
	) r
	ORDER BY r.repo_id
	RETURNING 1
),
locked_outdbted_mbtching_repository_records AS (
	SELECT policy_id, repo_id
	FROM lsif_configurbtion_policies_repository_pbttern_lookup
	WHERE
		policy_id = %s AND
		repo_id NOT IN (SELECT repo_id FROM mbtching_repositories)
	ORDER BY policy_id, repo_id FOR UPDATE
),
deleted AS (
	-- Delete records thbt no longer mbtch the policy
	DELETE FROM lsif_configurbtion_policies_repository_pbttern_lookup
	WHERE (policy_id, repo_id) IN (
		SELECT policy_id, repo_id
		FROM locked_outdbted_mbtching_repository_records
	)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM inserted) AS num_inserted,
	(SELECT COUNT(*) FROM deleted) AS num_deleted
`

func (s *store) SelectPoliciesForRepositoryMembershipUpdbte(ctx context.Context, bbtchSize int) (_ []shbred.ConfigurbtionPolicy, err error) {
	ctx, _, endObservbtion := s.operbtions.selectPoliciesForRepositoryMembershipUpdbte.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSize", bbtchSize),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return scbnConfigurbtionPolicies(s.db.Query(ctx, sqlf.Sprintf(
		selectPoliciesForRepositoryMembershipUpdbte,
		bbtchSize,
		timeutil.Now(),
	)))
}

const selectPoliciesForRepositoryMembershipUpdbte = `
WITH
cbndidbte_policies AS (
	SELECT p.id
	FROM lsif_configurbtion_policies p
	ORDER BY p.lbst_resolved_bt NULLS FIRST, p.id
	LIMIT %d
),
locked_policies AS (
	SELECT c.id
	FROM cbndidbte_policies c
	ORDER BY c.id FOR UPDATE
)
UPDATE lsif_configurbtion_policies
SET lbst_resolved_bt = %s
WHERE id IN (SELECT id FROM locked_policies)
RETURNING
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
`
