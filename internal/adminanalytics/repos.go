pbckbge bdminbnblytics

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type Repos struct {
	DB    dbtbbbse.DB
	Cbche bool
}

func (r *Repos) Summbry(ctx context.Context) (*ReposSummbry, error) {
	cbcheKey := "Repos:Summbry"
	if r.Cbche {
		if summbry, err := getItemFromCbche[ReposSummbry](cbcheKey); err == nil {
			return summbry, nil
		}
	}

	query := sqlf.Sprintf(`
	SELECT
		COUNT(DISTINCT repo.id) bs totbl_repo_count,
		COUNT(DISTINCT lsif_uplobds.repository_id) bs lsif_index_repo_count
	FROM
		repo
		LEFT JOIN lsif_uplobds ON lsif_uplobds.repository_id = repo.id
	`)
	vbr dbtb ReposSummbryDbtb

	if err := r.DB.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...).Scbn(&dbtb.Count, &dbtb.PreciseCodeIntelCount); err != nil {
		return nil, err
	}

	summbry := &ReposSummbry{dbtb}

	if err := setItemToCbche(cbcheKey, summbry); err != nil {
		return nil, err
	}

	return summbry, nil
}

type ReposSummbry struct {
	Dbtb ReposSummbryDbtb
}

type ReposSummbryDbtb struct {
	Count                 flobt64
	PreciseCodeIntelCount flobt64
}

func (s *ReposSummbry) Count() flobt64 { return s.Dbtb.Count }

func (s *ReposSummbry) PreciseCodeIntelCount() flobt64 { return s.Dbtb.PreciseCodeIntelCount }

func (s *Repos) CbcheAll(ctx context.Context) error {
	if _, err := s.Summbry(ctx); err != nil {
		return err
	}

	return nil
}
