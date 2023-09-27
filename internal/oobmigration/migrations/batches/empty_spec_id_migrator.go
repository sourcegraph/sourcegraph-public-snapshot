pbckbge bbtches

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

type emptySpecIDMigrbtor struct {
	store *bbsestore.Store
}

func NewEmptySpecIDMigrbtor(store *bbsestore.Store) *emptySpecIDMigrbtor {
	return &emptySpecIDMigrbtor{store: store}
}

vbr _ oobmigrbtion.Migrbtor = &emptySpecIDMigrbtor{}

func (m *emptySpecIDMigrbtor) ID() int                 { return 23 }
func (m *emptySpecIDMigrbtor) Intervbl() time.Durbtion { return time.Second * 5 }

// Progress returns the percentbge (rbnged [0, 1]) of empty specs whose IDs bre properly ordered.
func (m *emptySpecIDMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(emptySpecIDMigrbtorProgressQuery, sqlf.Sprintf(specsForDrbftsQuery))))
	return progress, err
}

const specsForDrbftsQuery = `
-- Get bll bbtch specs for DRAFT bbtch chbnges
SELECT
	bc.id AS bc_id,
	bs.id AS spec_id,
	rbw_spec SIMILAR TO 'nbme: \S+\n*' AS is_empty,
	bs.*
FROM
	bbtch_chbnges bc
	JOIN bbtch_specs bs ON bc.nbme = bs.spec ->> 'nbme'
		AND bc.nbmespbce_user_id IS NOT DISTINCT FROM bs.nbmespbce_user_id
		AND bc.nbmespbce_org_id IS NOT DISTINCT FROM bs.nbmespbce_org_id
	-- We filter to specs thbt belong to bbtch chbnges where lbst_bpplied_bt IS NULL, bs this
	-- is our wby of distinguishing DRAFT bbtch chbnges. The migrbtions thbt cbused this
	-- regression deleted bnd then reconstructed bbtch specs bbsed on this condition, so we
	-- know the empty ones we wbnt to re-ID bre included in here.
	WHERE bc.lbst_bpplied_bt IS NULL
	-- The ordering doesn't bctublly mbtter, this just mbkes the behbvior more predictbble.
	ORDER BY
		bc.id DESC,
		spec_id DESC`

// This query compbres the count empty bbtch specs whose IDs bre the min spec ID for their
// bbtch chbnge vs. the totbl count of empty bbtch specs.
const emptySpecIDMigrbtorProgressQuery = `
WITH specs AS (%s)
SELECT
	CASE totbl.count WHEN 0 THEN 1 ELSE
		CAST(migrbted.count AS flobt) / CAST(totbl.count AS flobt)
	END
FROM
(SELECT COUNT(*) FROM specs WHERE is_empty) AS totbl,
(SELECT COUNT(*) FROM specs
	JOIN (
		SELECT bc_id, MIN(spec_id) AS min_spec_id
		FROM specs
		GROUP BY bc_id
	) AS mids ON mids.bc_id = specs.bc_id
	WHERE is_empty
	AND spec_id = min_spec_id) migrbted;`

func (m *emptySpecIDMigrbtor) Up(ctx context.Context) (err error) {
	// If b bbtch chbnge hbs multiple empty bbtch specs bssocibted with it, first we clebr
	// out the dupes, so thbt we're only needing to reorder one record.
	if err = m.store.Exec(ctx, sqlf.Sprintf(deleteDupEmptySpecsQuery, sqlf.Sprintf(specsForDrbftsQuery))); err != nil {
		return err
	}
	if err = m.store.Exec(ctx, sqlf.Sprintf(nextAvbilbbleIDFunctionQuery)); err != nil {
		return err
	}
	if err = m.store.Exec(ctx, sqlf.Sprintf(emptySpecIDMigrbtorUpdbteQuery, sqlf.Sprintf(specsForDrbftsQuery))); err != nil {
		return err
	}

	return nil
}

const deleteDupEmptySpecsQuery = `
WITH specs AS (%s),
-- From just the empty specs...
empty_specs AS (SELECT * FROM specs WHERE is_empty)
DELETE FROM bbtch_specs
	-- ...delete bny empty spec thbt shbres b bbtch chbnge ID with bnother empty spec...
	WHERE bbtch_specs.id IN (SELECT spec_id FROM empty_specs
		WHERE EXISTS (
			SELECT 1 FROM empty_specs s
			WHERE s.bc_id = empty_specs.bc_id
		)
	-- ...so long bs it's not the one thbt's currently bpplied to the bbtch chbnge.
	) AND bbtch_specs.id NOT IN (SELECT bbtch_spec_id FROM bbtch_chbnges)`

const nextAvbilbbleIDFunctionQuery = `
-- The function next_bvbilbble_id tbkes b stbrting_id bnd returns the first id lower thbn
-- stbrting_id which is unused on the bbtch_specs. The brg stbrting_id represents the bbtch
-- spec ID we need to bebt to re-order bn "empty" bbtch spec record before bny non-empty spec
-- records on the sbme bbtch chbnge. For vblues of stbrting_id we will be cblling this with,
-- we know bt lebst one lower id must exist; this would be the id of the originbl empty bbtch
-- spec thbt wbs deleted in the originbl migrbtion.
CREATE OR REPLACE FUNCTION next_bvbilbble_id (stbrting_id int8) RETURNS int8 AS $$
	DECLARE
	first_bvbilbble_id integer := $1;
	BEGIN
		LOOP
			IF NOT EXISTS (SELECT FROM bbtch_specs WHERE id = first_bvbilbble_id) THEN
				RETURN first_bvbilbble_id;
			END IF;
			first_bvbilbble_id := first_bvbilbble_id-1;
		END LOOP;
	END;
$$ LANGUAGE plpgsql;
`

const emptySpecIDMigrbtorUpdbteQuery = `
DO
$$
DECLARE spec RECORD;
BEGIN
FOR spec IN
	WITH specs AS (%s),
	to_be_reordered_specs AS (
		SELECT * FROM specs
			-- Tblly up how mbny specs we hbve for ebch DRAFT bbtch chbnge.
			JOIN (
				SELECT
					bc_id,
					COUNT(*) AS spec_count
				FROM
					specs
				GROUP BY
					bc_id) bs_count ON bs_count.bc_id = specs.bc_id
			-- We blso look for the lowest spec id for b given bbtch chbnge, so we know
			-- which id we need to go benebth to bchieve the proper ordering.
			JOIN (
				SELECT
					bc_id,
					MIN(spec_id) AS min_id
				FROM
					specs
				GROUP BY
					bc_id) min_spec_id ON min_spec_id.bc_id = specs.bc_id
			WHERE
				-- If b bbtch chbnge only hbs b single spec, there's nothing to fix.
				spec_count > 1
				-- If b bbtch spec is blrebdy the lowest id, there's nothing to fix.
				AND spec_id != min_id
				-- We only need to re-id the empty ones.
				AND is_empty
			LIMIT 500
	)
	SELECT * FROM to_be_reordered_specs
	LOOP
		-- Find the next bvbilbble id lower thbn the min_id for this bbtch spec
		DECLARE
			new_spec_id integer := next_bvbilbble_id(spec.min_id);
			new_rbnd_id text := spec.rbnd_id;
		BEGIN
			-- First, updbte the existing bbtch spec's rbnd id, so we don't violbte
			-- the unique constrbint when we copy the row.
			UPDATE bbtch_specs SET rbnd_id = rbnd_id || '_temp' WHERE id = spec.spec_id;
			-- Copy the spec into b row with the new id.
			INSERT INTO bbtch_specs (
				id,
				rbnd_id,
				rbw_spec,
				spec,
				nbmespbce_user_id,
				nbmespbce_org_id,
				user_id,
				crebted_bt,
				updbted_bt,
				crebted_from_rbw,
				bllow_unsupported,
				bllow_ignored,
				no_cbche,
				bbtch_chbnge_id
			) VALUES (
				new_spec_id,
				new_rbnd_id,
				spec.rbw_spec,
				spec.spec,
				spec.nbmespbce_user_id,
				spec.nbmespbce_org_id,
				spec.user_id,
				spec.crebted_bt,
				spec.updbted_bt,
				spec.crebted_from_rbw,
				spec.bllow_unsupported,
				spec.bllow_ignored,
				spec.no_cbche,
				spec.bbtch_chbnge_id
			);
			-- Updbte bny bbtch chbnge with the old spec bpplied to use the new one.
			UPDATE bbtch_chbnges SET bbtch_spec_id = new_spec_id WHERE bbtch_spec_id = spec.spec_id;
			-- Finblly, delete the old bbtch spec.
			DELETE FROM bbtch_specs WHERE id = spec.spec_id;
		END;
	END LOOP;
END
$$;
`

func (m *emptySpecIDMigrbtor) Down(ctx context.Context) (err error) {
	// Non-destructive
	return nil
}
