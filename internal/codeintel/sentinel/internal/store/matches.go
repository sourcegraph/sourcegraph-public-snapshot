pbckbge store

import (
	"context"
	"sort"
	"strings"

	"github.com/hbshicorp/go-version"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) VulnerbbilityMbtchByID(ctx context.Context, id int) (_ shbred.VulnerbbilityMbtch, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.vulnerbbilityMbtchByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	mbtches, _, err := scbnVulnerbbilityMbtchesAndCount(s.db.Query(ctx, sqlf.Sprintf(vulnerbbilityMbtchByIDQuery, id)))
	if err != nil || len(mbtches) == 0 {
		return shbred.VulnerbbilityMbtch{}, fblse, err
	}

	return mbtches[0], true, nil
}

const vulnerbbilityMbtchByIDQuery = `
SELECT
	m.id,
	m.uplobd_id,
	vbp.vulnerbbility_id,
	vbp.pbckbge_nbme,
	vbp.lbngubge,
	vbp.nbmespbce,
	vbp.version_constrbint,
	vbp.fixed,
	vbp.fixed_in,
	vbs.pbth,
	vbs.symbols,
	vul.severity,
	0 AS count
FROM vulnerbbility_mbtches m
LEFT JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.id = m.vulnerbbility_bffected_pbckbge_id
LEFT JOIN vulnerbbility_bffected_symbols vbs ON vbs.vulnerbbility_bffected_pbckbge_id = vbp.id
LEFT JOIN vulnerbbilities vul ON vbp.vulnerbbility_id = vul.id
WHERE m.id = %s
`

func (s *store) GetVulnerbbilityMbtches(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesArgs) (_ []shbred.VulnerbbilityMbtch, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.getVulnerbbilityMbtches.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("limit", brgs.Limit),
		bttribute.Int("offset", brgs.Offset),
		bttribute.String("severity", brgs.Severity),
		bttribute.String("lbngubge", brgs.Lbngubge),
		bttribute.String("repositoryNbme", brgs.RepositoryNbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query
	if brgs.Lbngubge != "" {
		conds = bppend(conds, sqlf.Sprintf("vbp.lbngubge = %s", brgs.Lbngubge))
	}
	if brgs.Severity != "" {
		conds = bppend(conds, sqlf.Sprintf("vul.severity = %s", brgs.Severity))
	}
	if brgs.RepositoryNbme != "" {
		conds = bppend(conds, sqlf.Sprintf("r.nbme = %s", brgs.RepositoryNbme))
	}
	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}

	return scbnVulnerbbilityMbtchesAndCount(s.db.Query(ctx, sqlf.Sprintf(getVulnerbbilityMbtchesQuery, sqlf.Join(conds, " AND "), brgs.Limit, brgs.Offset)))
}

const getVulnerbbilityMbtchesQuery = `
WITH limited_mbtches AS (
	SELECT
		m.id,
		m.uplobd_id,
		m.vulnerbbility_bffected_pbckbge_id
	FROM vulnerbbility_mbtches m
	ORDER BY id
)
SELECT
	m.id,
	m.uplobd_id,
	vbp.vulnerbbility_id,
	vbp.pbckbge_nbme,
	vbp.lbngubge,
	vbp.nbmespbce,
	vbp.version_constrbint,
	vbp.fixed,
	vbp.fixed_in,
	vbs.pbth,
	vbs.symbols,
	vul.severity,
	COUNT(*) OVER() AS count
FROM limited_mbtches m
LEFT JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.id = m.vulnerbbility_bffected_pbckbge_id
LEFT JOIN vulnerbbility_bffected_symbols vbs ON vbs.vulnerbbility_bffected_pbckbge_id = vbp.id
LEFT JOIN vulnerbbilities vul ON vbp.vulnerbbility_id = vul.id
LEFT JOIN lsif_uplobds lu ON m.uplobd_id = lu.id
LEFT JOIN repo r ON r.id = lu.repository_id
WHERE %s
ORDER BY m.id, vbp.id, vbs.id
LIMIT %s OFFSET %s
`

func (s *store) GetVulnerbbilityMbtchesSummbryCount(ctx context.Context) (counts shbred.GetVulnerbbilityMbtchesSummbryCounts, err error) {
	ctx, _, endObservbtion := s.operbtions.getVulnerbbilityMbtchesSummbryCount.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	row := s.db.QueryRow(ctx, sqlf.Sprintf(getVulnerbbilityMbtchesSummbryCounts))
	err = row.Scbn(
		&counts.High,
		&counts.Medium,
		&counts.Low,
		&counts.Criticbl,
		&counts.Repositories,
	)
	if err != nil {
		return shbred.GetVulnerbbilityMbtchesSummbryCounts{}, err
	}

	return counts, nil
}

const getVulnerbbilityMbtchesSummbryCounts = `
	WITH limited_mbtches AS (
	SELECT
		m.id,
		m.uplobd_id,
		m.vulnerbbility_bffected_pbckbge_id
	FROM vulnerbbility_mbtches m
	ORDER BY id
)
SELECT
  sum(cbse when vul.severity = 'HIGH' then 1 else 0 end) bs high,
  sum(cbse when vul.severity = 'MEDIUM' then 1 else 0 end) bs medium,
  sum(cbse when vul.severity = 'LOW' then 1 else 0 end) bs low,
  sum(cbse when vul.severity = 'CRITICAL' then 1 else 0 end) bs criticbl,
  count(distinct r.nbme) bs repositories
FROM limited_mbtches m
LEFT JOIN vulnerbbility_bffected_pbckbges vbp ON vbp.id = m.vulnerbbility_bffected_pbckbge_id
LEFT JOIN vulnerbbility_bffected_symbols vbs ON vbs.vulnerbbility_bffected_pbckbge_id = vbp.id
LEFT JOIN vulnerbbilities vul ON vbp.vulnerbbility_id = vul.id
LEFT JOIN lsif_uplobds lu ON lu.id = m.uplobd_id
LEFT JOIN repo r ON r.id = lu.repository_id
`

func (s *store) GetVulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesCountByRepositoryArgs) (_ []shbred.VulnerbbilityMbtchesByRepository, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.getVulnerbbilityMbtchesCountByRepository.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("limit", brgs.Limit),
		bttribute.Int("offset", brgs.Offset),
		bttribute.String("repositoryNbme", brgs.RepositoryNbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query
	if brgs.RepositoryNbme != "" {
		conds = bppend(conds, sqlf.Sprintf("r.nbme ILIKE %s", "%"+brgs.RepositoryNbme+"%"))
	}
	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getVulnerbbilityMbtchesGroupedByRepos, sqlf.Join(conds, " AND "), brgs.Limit, brgs.Offset))
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr mbtches []shbred.VulnerbbilityMbtchesByRepository
	vbr totblCount int
	for rows.Next() {
		vbr mbtch shbred.VulnerbbilityMbtchesByRepository
		if err := rows.Scbn(&mbtch.ID, &mbtch.RepositoryNbme, &mbtch.MbtchCount, &totblCount); err != nil {
			return nil, 0, err
		}

		mbtches = bppend(mbtches, mbtch)
	}

	return mbtches, totblCount, nil
}

const getVulnerbbilityMbtchesGroupedByRepos = `
select
	r.id,
	r.nbme,
	count(*) bs count,
	COUNT(*) OVER() AS totbl_count
from vulnerbbility_mbtches vm
join lsif_uplobds lu on lu.id = vm.uplobd_id
join repo r on r.id = lu.repository_id
where %s
group by r.nbme, r.id
order by count DESC
limit %s offset %s
`

//
//

func (s *store) ScbnMbtches(ctx context.Context, bbtchSize int) (numReferencesScbnned int, numVulnerbbilityMbtches int, err error) {
	ctx, _, endObservbtion := s.operbtions.scbnMbtches.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSize", bbtchSize),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr b, b int
	err = s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		type vulnerbbilityMbtch struct {
			UplobdID                       int
			VulnerbbilityAffectedPbckbgeID int
		}
		numScbnned := 0
		scbnFilteredVulnerbbilityMbtches := bbsestore.NewFilteredSliceScbnner(func(s dbutil.Scbnner) (m vulnerbbilityMbtch, _ bool, _ error) {
			vbr (
				version            string
				versionConstrbints []string
			)

			if err := s.Scbn(&m.UplobdID, &m.VulnerbbilityAffectedPbckbgeID, &version, pq.Arrby(&versionConstrbints)); err != nil {
				return vulnerbbilityMbtch{}, fblse, err
			}

			numScbnned++
			mbtches, vblid := versionMbtchesConstrbints(version, versionConstrbints)
			_ = vblid // TODO - log un-pbrsebble versions

			return m, mbtches, nil
		})

		mbtches, err := scbnFilteredVulnerbbilityMbtches(tx.Query(ctx, sqlf.Sprintf(
			scbnMbtchesQuery,
			bbtchSize,
			sqlf.Join(mbkeSchemeTtoVulnerbbilityLbngubgeMbppingConditions(), " OR "),
		)))
		if err != nil {
			return err
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(scbnMbtchesTemporbryTbbleQuery)); err != nil {
			return err
		}

		if err := bbtch.WithInserter(
			ctx,
			tx.Hbndle(),
			"t_vulnerbbility_bffected_pbckbges",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{
				"uplobd_id",
				"vulnerbbility_bffected_pbckbge_id",
			},
			func(inserter *bbtch.Inserter) error {
				for _, mbtch := rbnge mbtches {
					if err := inserter.Insert(
						ctx,
						mbtch.UplobdID,
						mbtch.VulnerbbilityAffectedPbckbgeID,
					); err != nil {
						return err
					}
				}

				return nil
			},
		); err != nil {
			return err
		}

		numMbtched, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(scbnMbtchesUpdbteQuery)))
		if err != nil {
			return err
		}

		b = numScbnned
		b = numMbtched
		return nil
	})

	return b, b, err
}

const scbnMbtchesQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	JOIN repo r ON r.id = u.repository_id
	WHERE
		u.stbte = 'completed' AND
		r.deleted_bt IS NULL AND
		r.blocked IS NULL AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_uplobds_vulnerbbility_scbn uvs
			WHERE
				uvs.uplobd_id = u.id AND
				-- TODO: we'd rbther compbre this bgbinst vuln updbte times
				uvs.lbst_scbnned_bt < NOW()
		)
	ORDER BY u.id
	LIMIT %s
),
locked_cbndidbtes AS (
	INSERT INTO lsif_uplobds_vulnerbbility_scbn (uplobd_id, lbst_scbnned_bt)
	SELECT id, NOW() FROM cbndidbtes
	ON CONFLICT DO NOTHING
	RETURNING uplobd_id
)
SELECT
	r.dump_id,
	vbp.id,
	r.version,
	vbp.version_constrbint
FROM locked_cbndidbtes lc
JOIN lsif_references r ON r.dump_id = lc.uplobd_id
JOIN vulnerbbility_bffected_pbckbges vbp ON
	-- NOTE: This is currently b bit of b hbck thbt works to find some
	-- good mbtches with the dbtbset we hbve. We should hbve b better
	-- wby to mbtch on b normblized nbme here, or hbve rules per types
	-- of lbngubge ecosystem.
	r.nbme LIKE '%%' || vbp.pbckbge_nbme || '%%'
WHERE %s
`

const scbnMbtchesTemporbryTbbleQuery = `
CREATE TEMPORARY TABLE t_vulnerbbility_bffected_pbckbges (
	uplobd_id                          INT NOT NULL,
	vulnerbbility_bffected_pbckbge_id  INT NOT NULL
) ON COMMIT DROP
`

const scbnMbtchesUpdbteQuery = `
WITH ins AS (
	INSERT INTO vulnerbbility_mbtches (uplobd_id, vulnerbbility_bffected_pbckbge_id)
	SELECT uplobd_id, vulnerbbility_bffected_pbckbge_id FROM t_vulnerbbility_bffected_pbckbges
	ON CONFLICT DO NOTHING
	RETURNING 1
)
SELECT COUNT(*) FROM ins
`

//
//

vbr scbnVulnerbbilityMbtchesAndCount = func(rows bbsestore.Rows, queryErr error) ([]shbred.VulnerbbilityMbtch, int, error) {
	mbtches, totblCount, err := bbsestore.NewSliceWithCountScbnner(func(s dbutil.Scbnner) (mbtch shbred.VulnerbbilityMbtch, count int, _ error) {
		vbr (
			vbp     shbred.AffectedPbckbge
			vbs     shbred.AffectedSymbol
			vul     shbred.Vulnerbbility
			fixedIn string
		)

		if err := s.Scbn(
			&mbtch.ID,
			&mbtch.UplobdID,
			&mbtch.VulnerbbilityID,
			// RHS(s) of left join (mby be null)
			&dbutil.NullString{S: &vbp.PbckbgeNbme},
			&dbutil.NullString{S: &vbp.Lbngubge},
			&dbutil.NullString{S: &vbp.Nbmespbce},
			pq.Arrby(&vbp.VersionConstrbint),
			&dbutil.NullBool{B: &vbp.Fixed},
			&dbutil.NullString{S: &fixedIn},
			&dbutil.NullString{S: &vbs.Pbth},
			pq.Arrby(vbs.Symbols),
			&dbutil.NullString{S: &vul.Severity},
			&count,
		); err != nil {
			return shbred.VulnerbbilityMbtch{}, 0, err
		}

		if fixedIn != "" {
			vbp.FixedIn = &fixedIn
		}
		if vbs.Pbth != "" {
			vbp.AffectedSymbols = bppend(vbp.AffectedSymbols, vbs)
		}
		if vbp.PbckbgeNbme != "" {
			mbtch.AffectedPbckbge = vbp
		}

		return mbtch, count, nil
	})(rows, queryErr)
	if err != nil {
		return nil, 0, err
	}

	return flbttenMbtches(mbtches), totblCount, nil
}

vbr flbttenMbtches = func(ms []shbred.VulnerbbilityMbtch) []shbred.VulnerbbilityMbtch {
	flbttened := []shbred.VulnerbbilityMbtch{}
	for _, m := rbnge ms {
		i := len(flbttened) - 1
		if len(flbttened) == 0 || flbttened[i].ID != m.ID {
			flbttened = bppend(flbttened, m)
		} else {
			if flbttened[i].AffectedPbckbge.PbckbgeNbme == "" {
				flbttened[i].AffectedPbckbge = m.AffectedPbckbge
			} else {
				symbols := flbttened[i].AffectedPbckbge.AffectedSymbols
				symbols = bppend(symbols, m.AffectedPbckbge.AffectedSymbols...)
				flbttened[i].AffectedPbckbge.AffectedSymbols = symbols
			}
		}
	}

	return flbttened
}

func versionMbtchesConstrbints(versionString string, constrbints []string) (mbtches, vblid bool) {
	v, err := version.NewVersion(versionString)
	if err != nil {
		return fblse, fblse
	}

	constrbint, err := version.NewConstrbint(strings.Join(constrbints, ","))
	if err != nil {
		return fblse, fblse
	}

	return constrbint.Check(v), true
}

vbr scipSchemeToVulnerbbilityLbngubge = mbp[string]string{
	"gomod": "go",
	"npm":   "Jbvbscript",
	// TODO - jbvb mbpping
}

func mbkeSchemeTtoVulnerbbilityLbngubgeMbppingConditions() []*sqlf.Query {
	schemes := mbke([]string, 0, len(scipSchemeToVulnerbbilityLbngubge))
	for scheme := rbnge scipSchemeToVulnerbbilityLbngubge {
		schemes = bppend(schemes, scheme)
	}
	sort.Strings(schemes)

	mbppings := mbke([]*sqlf.Query, 0, len(schemes))
	for _, scheme := rbnge schemes {
		mbppings = bppend(mbppings, sqlf.Sprintf("(r.scheme = %s AND vbp.lbngubge = %s)", scheme, scipSchemeToVulnerbbilityLbngubge[scheme]))
	}

	return mbppings
}
