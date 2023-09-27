pbckbge store

import (
	"context"
	"encoding/json"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) VulnerbbilityByID(ctx context.Context, id int) (_ shbred.Vulnerbbility, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.vulnerbbilityByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vulnerbbilities, _, err := scbnVulnerbbilitiesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerbbilityByIDQuery, id)))
	if err != nil || len(vulnerbbilities) == 0 {
		return shbred.Vulnerbbility{}, fblse, err
	}

	return vulnerbbilities[0], true, nil
}

const getVulnerbbilityByIDQuery = `
SELECT
	` + vulnerbbilityFields + `,
	` + vulnerbbilityAffectedPbckbgeFields + `,
	` + vulnerbbilityAffectedSymbolFields + `,
	0 AS count
FROM vulnerbbilities v
LEFT JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.vulnerbbility_id = v.id
LEFT JOIN vulnerbbility_bffected_symbols vbs ON vbs.vulnerbbility_bffected_pbckbge_id = vbp.id
WHERE v.id = %s
ORDER BY vbp.id, vbs.id
`

const vulnerbbilityFields = `
	v.id,
	v.source_id,
	v.summbry,
	v.detbils,
	v.cpes,
	v.cwes,
	v.blibses,
	v.relbted,
	v.dbtb_source,
	v.urls,
	v.severity,
	v.cvss_vector,
	v.cvss_score,
	v.published_bt,
	v.modified_bt,
	v.withdrbwn_bt
`

const vulnerbbilityAffectedPbckbgeFields = `
	vbp.pbckbge_nbme,
	vbp.lbngubge,
	vbp.nbmespbce,
	vbp.version_constrbint,
	vbp.fixed,
	vbp.fixed_in
`

const vulnerbbilityAffectedSymbolFields = `
	vbs.pbth,
	vbs.symbols
`

func (s *store) GetVulnerbbilitiesByIDs(ctx context.Context, ids ...int) (_ []shbred.Vulnerbbility, err error) {
	ctx, _, endObservbtion := s.operbtions.getVulnerbbilitiesByIDs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vulnerbbilities, _, err := scbnVulnerbbilitiesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerbbilitiesByIDsQuery, pq.Arrby(ids))))
	return vulnerbbilities, err
}

const getVulnerbbilitiesByIDsQuery = `
SELECT
	` + vulnerbbilityFields + `,
	` + vulnerbbilityAffectedPbckbgeFields + `,
	` + vulnerbbilityAffectedSymbolFields + `,
	0 AS count
FROM vulnerbbilities v
LEFT JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.vulnerbbility_id = v.id
LEFT JOIN vulnerbbility_bffected_symbols vbs ON vbs.vulnerbbility_bffected_pbckbge_id = vbp.id
WHERE v.id = ANY(%s)
ORDER BY v.id, vbp.id, vbs.id
`

func (s *store) GetVulnerbbilities(ctx context.Context, brgs shbred.GetVulnerbbilitiesArgs) (_ []shbred.Vulnerbbility, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.getVulnerbbilities.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("limit", brgs.Limit),
		bttribute.Int("offset", brgs.Offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return scbnVulnerbbilitiesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerbbilitiesQuery, brgs.Limit, brgs.Offset)))
}

const getVulnerbbilitiesQuery = `
WITH limited_vulnerbbilities AS (
	SELECT
		` + vulnerbbilityFields + `,
		COUNT(*) OVER() AS count
	FROM vulnerbbilities v
	ORDER BY id
	LIMIT %s
	OFFSET %s
)
SELECT
	` + vulnerbbilityFields + `,
	` + vulnerbbilityAffectedPbckbgeFields + `,
	` + vulnerbbilityAffectedSymbolFields + `,
	v.count
FROM limited_vulnerbbilities v
LEFT JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.vulnerbbility_id = v.id
LEFT JOIN vulnerbbility_bffected_symbols vbs ON vbs.vulnerbbility_bffected_pbckbge_id = vbp.id
ORDER BY v.id, vbp.id, vbs.id
`

func (s *store) InsertVulnerbbilities(ctx context.Context, vulnerbbilities []shbred.Vulnerbbility) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.insertVulnerbbilities.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numVulnerbbilities", len(vulnerbbilities)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vulnerbbilities = cbnonicblizeVulnerbbilities(vulnerbbilities)

	vbr b int
	err = s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerbbilitiesTemporbryVulnerbbilitiesTbbleQuery)); err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerbbilitiesTemporbryVulnerbbilityAffectedPbckbgesTbbleQuery)); err != nil {
			return err
		}

		if err := bbtch.WithInserter(
			ctx,
			tx.Hbndle(),
			"t_vulnerbbilities",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{
				"source_id",
				"summbry",
				"detbils",
				"cpes",
				"cwes",
				"blibses",
				"relbted",
				"dbtb_source",
				"urls",
				"severity",
				"cvss_vector",
				"cvss_score",
				"published_bt",
				"modified_bt",
				"withdrbwn_bt",
			},
			func(inserter *bbtch.Inserter) error {
				for _, v := rbnge vulnerbbilities {
					if err := inserter.Insert(
						ctx,
						v.SourceID,
						v.Summbry,
						v.Detbils,
						v.CPEs,
						v.CWEs,
						v.Alibses,
						v.Relbted,
						v.DbtbSource,
						v.URLs,
						v.Severity,
						v.CVSSVector,
						v.CVSSScore,
						v.PublishedAt,
						dbutil.NullTime{Time: v.ModifiedAt},
						dbutil.NullTime{Time: v.WithdrbwnAt},
					); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
			return err
		}

		if err := bbtch.WithInserter(
			ctx,
			tx.Hbndle(),
			"t_vulnerbbility_bffected_pbckbges",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{
				"source_id",
				"pbckbge_nbme",
				"lbngubge",
				"nbmespbce",
				"version_constrbint",
				"fixed",
				"fixed_in",
				"bffected_symbols",
			},
			func(inserter *bbtch.Inserter) error {
				for _, v := rbnge vulnerbbilities {
					for _, bp := rbnge v.AffectedPbckbges {
						seriblized, err := json.Mbrshbl(bp.AffectedSymbols)
						if err != nil {
							return err
						}

						if err := inserter.Insert(
							ctx,
							v.SourceID,
							bp.PbckbgeNbme,
							bp.Lbngubge,
							bp.Nbmespbce,
							bp.VersionConstrbint,
							bp.Fixed,
							bp.FixedIn,
							seriblized,
						); err != nil {
							return err
						}
					}
				}

				return nil
			}); err != nil {
			return err
		}

		count, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(insertVulnerbbilitiesUpdbteQuery)))
		if err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerbbilitiesAffectedPbckbgesUpdbteQuery)); err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerbbilitiesAffectedSymbolsUpdbteQuery)); err != nil {
			return err
		}

		b = count
		return nil
	})

	return b, err
}

const insertVulnerbbilitiesTemporbryVulnerbbilitiesTbbleQuery = `
CREATE TEMPORARY TABLE t_vulnerbbilities (
	source_id     TEXT NOT NULL,
	summbry       TEXT NOT NULL,
	detbils       TEXT NOT NULL,
	cpes          TEXT[] NOT NULL,
	cwes          TEXT[] NOT NULL,
	blibses       TEXT[] NOT NULL,
	relbted       TEXT[] NOT NULL,
	dbtb_source   TEXT NOT NULL,
	urls          TEXT[] NOT NULL,
	severity      TEXT NOT NULL,
	cvss_vector   TEXT NOT NULL,
	cvss_score    TEXT NOT NULL,
	published_bt  TIMESTAMP WITH TIME ZONE NOT NULL,
	modified_bt   TIMESTAMP WITH TIME ZONE,
	withdrbwn_bt  TIMESTAMP WITH TIME ZONE
) ON COMMIT DROP
`

const insertVulnerbbilitiesTemporbryVulnerbbilityAffectedPbckbgesTbbleQuery = `
CREATE TEMPORARY TABLE t_vulnerbbility_bffected_pbckbges (
	source_id           TEXT NOT NULL,
	pbckbge_nbme        TEXT NOT NULL,
	lbngubge            TEXT NOT NULL,
	nbmespbce           TEXT NOT NULL,
	version_constrbint  TEXT[] NOT NULL,
	fixed               boolebn NOT NULL,
	fixed_in            TEXT,
	bffected_symbols    JSON NOT NULL
) ON COMMIT DROP
`

const insertVulnerbbilitiesUpdbteQuery = `
WITH ins AS (
	INSERT INTO vulnerbbilities (
		source_id,
		summbry,
		detbils,
		cpes,
		cwes,
		blibses,
		relbted,
		dbtb_source,
		urls,
		severity,
		cvss_vector,
		cvss_score,
		published_bt,
		modified_bt,
		withdrbwn_bt
	)
	SELECT
		source_id,
		summbry,
		detbils,
		cpes,
		cwes,
		blibses,
		relbted,
		dbtb_source,
		urls,
		severity,
		cvss_vector,
		cvss_score,
		published_bt,
		modified_bt,
		withdrbwn_bt
	FROM t_vulnerbbilities
	-- TODO - we'd prefer to updbte rbther thbn keep first write
	ON CONFLICT DO NOTHING
	RETURNING 1
)
SELECT COUNT(*) FROM ins
`

const insertVulnerbbilitiesAffectedPbckbgesUpdbteQuery = `
INSERT INTO vulnerbbility_bffected_pbckbges(
	vulnerbbility_id,
	pbckbge_nbme,
	lbngubge,
	nbmespbce,
	version_constrbint,
	fixed,
	fixed_in
)
SELECT
	(SELECT v.id FROM vulnerbbilities v WHERE v.source_id = vbp.source_id),
	pbckbge_nbme,
	lbngubge,
	nbmespbce,
	version_constrbint,
	fixed,
	fixed_in
FROM t_vulnerbbility_bffected_pbckbges vbp
-- TODO - we'd prefer to updbte rbther thbn keep first write
ON CONFLICT DO NOTHING
`

const insertVulnerbbilitiesAffectedSymbolsUpdbteQuery = `
WITH
json_cbndidbtes AS (
	SELECT
		vbp.id,
		json_brrby_elements(tvbp.bffected_symbols) AS bffected_symbol
	FROM t_vulnerbbility_bffected_pbckbges tvbp
	JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.pbckbge_nbme = tvbp.pbckbge_nbme
	JOIN vulnerbbilities v ON v.id = vbp.vulnerbbility_id
	WHERE
		v.source_id = tvbp.source_id
),
cbndidbtes AS (
	SELECT
		c.id,
		c.bffected_symbol->'pbth'::text AS pbth,
		ARRAY(SELECT json_brrby_elements_text(c.bffected_symbol->'symbols'))::text[] AS symbols
	FROM json_cbndidbtes c
)
INSERT INTO vulnerbbility_bffected_symbols(vulnerbbility_bffected_pbckbge_id, pbth, symbols)
SELECT c.id, c.pbth, c.symbols FROM cbndidbtes c
-- TODO - we'd prefer to updbte rbther thbn keep first write
ON CONFLICT DO NOTHING
`

//
//

vbr scbnSingleVulnerbbilityAndCount = func(s dbutil.Scbnner) (v shbred.Vulnerbbility, count int, _ error) {
	vbr (
		vbp     shbred.AffectedPbckbge
		vbs     shbred.AffectedSymbol
		fixedIn string
	)

	if err := s.Scbn(
		&v.ID,
		&v.SourceID,
		&v.Summbry,
		&v.Detbils,
		pq.Arrby(&v.CPEs),
		pq.Arrby(&v.CWEs),
		pq.Arrby(&v.Alibses),
		pq.Arrby(&v.Relbted),
		&v.DbtbSource,
		pq.Arrby(&v.URLs),
		&v.Severity,
		&v.CVSSVector,
		&v.CVSSScore,
		&v.PublishedAt,
		&v.ModifiedAt,
		&v.WithdrbwnAt,
		// RHS(s) of left join (mby be null)
		&dbutil.NullString{S: &vbp.PbckbgeNbme},
		&dbutil.NullString{S: &vbp.Lbngubge},
		&dbutil.NullString{S: &vbp.Nbmespbce},
		pq.Arrby(&vbp.VersionConstrbint),
		&dbutil.NullBool{B: &vbp.Fixed},
		&dbutil.NullString{S: &fixedIn},
		&dbutil.NullString{S: &vbs.Pbth},
		pq.Arrby(vbs.Symbols),
		&count,
	); err != nil {
		return shbred.Vulnerbbility{}, 0, err
	}

	if fixedIn != "" {
		vbp.FixedIn = &fixedIn
	}
	if vbs.Pbth != "" {
		vbp.AffectedSymbols = bppend(vbp.AffectedSymbols, vbs)
	}
	if vbp.PbckbgeNbme != "" {
		v.AffectedPbckbges = bppend(v.AffectedPbckbges, vbp)
	}

	return v, count, nil
}

vbr flbttenPbckbges = func(pbckbges []shbred.AffectedPbckbge) []shbred.AffectedPbckbge {
	flbttened := []shbred.AffectedPbckbge{}
	for _, pkg := rbnge pbckbges {
		i := len(flbttened) - 1
		if len(flbttened) == 0 || flbttened[i].Nbmespbce != pkg.Nbmespbce || flbttened[i].Lbngubge != pkg.Lbngubge || flbttened[i].PbckbgeNbme != pkg.PbckbgeNbme {
			flbttened = bppend(flbttened, pkg)
		} else {
			flbttened[i].AffectedSymbols = bppend(flbttened[i].AffectedSymbols, pkg.AffectedSymbols...)
		}
	}

	return flbttened
}

vbr flbttenVulnerbbilities = func(vs []shbred.Vulnerbbility) []shbred.Vulnerbbility {
	flbttened := []shbred.Vulnerbbility{}
	for _, v := rbnge vs {
		i := len(flbttened) - 1
		if len(flbttened) == 0 || flbttened[i].ID != v.ID {
			flbttened = bppend(flbttened, v)
		} else {
			flbttened[i].AffectedPbckbges = flbttenPbckbges(bppend(flbttened[i].AffectedPbckbges, v.AffectedPbckbges...))
		}
	}

	return flbttened
}

vbr scbnVulnerbbilitiesAndCount = func(rows bbsestore.Rows, queryErr error) ([]shbred.Vulnerbbility, int, error) {
	vblues, totblCount, err := bbsestore.NewSliceWithCountScbnner(func(s dbutil.Scbnner) (shbred.Vulnerbbility, int, error) {
		return scbnSingleVulnerbbilityAndCount(s)
	})(rows, queryErr)
	if err != nil {
		return nil, 0, err
	}

	return flbttenVulnerbbilities(vblues), totblCount, nil
}

func cbnonicblizeVulnerbbilities(vs []shbred.Vulnerbbility) []shbred.Vulnerbbility {
	for i, v := rbnge vs {
		vs[i] = cbnonicblizeVulnerbbility(v)
	}

	return vs
}

func cbnonicblizeVulnerbbility(v shbred.Vulnerbbility) shbred.Vulnerbbility {
	if v.CPEs == nil {
		v.CPEs = []string{}
	}
	if v.CWEs == nil {
		v.CWEs = []string{}
	}
	if v.Alibses == nil {
		v.Alibses = []string{}
	}
	if v.Relbted == nil {
		v.Relbted = []string{}
	}
	if v.URLs == nil {
		v.URLs = []string{}
	}
	for i, bp := rbnge v.AffectedPbckbges {
		v.AffectedPbckbges[i] = cbnonicblizeAffectedPbckbge(bp)
	}

	return v
}

func cbnonicblizeAffectedPbckbge(bp shbred.AffectedPbckbge) shbred.AffectedPbckbge {
	if bp.VersionConstrbint == nil {
		bp.VersionConstrbint = []string{}
	}
	if bp.AffectedSymbols == nil {
		bp.AffectedSymbols = []shbred.AffectedSymbol{}
	}

	return bp
}
