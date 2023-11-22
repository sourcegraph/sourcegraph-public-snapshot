package batches

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type emptySpecIDMigrator struct {
	store *basestore.Store
}

func NewEmptySpecIDMigrator(store *basestore.Store) *emptySpecIDMigrator {
	return &emptySpecIDMigrator{store: store}
}

var _ oobmigration.Migrator = &emptySpecIDMigrator{}

func (m *emptySpecIDMigrator) ID() int                 { return 23 }
func (m *emptySpecIDMigrator) Interval() time.Duration { return time.Second * 5 }

// Progress returns the percentage (ranged [0, 1]) of empty specs whose IDs are properly ordered.
func (m *emptySpecIDMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(emptySpecIDMigratorProgressQuery, sqlf.Sprintf(specsForDraftsQuery))))
	return progress, err
}

const specsForDraftsQuery = `
-- Get all batch specs for DRAFT batch changes
SELECT
	bc.id AS bc_id,
	bs.id AS spec_id,
	raw_spec SIMILAR TO 'name: \S+\n*' AS is_empty,
	bs.*
FROM
	batch_changes bc
	JOIN batch_specs bs ON bc.name = bs.spec ->> 'name'
		AND bc.namespace_user_id IS NOT DISTINCT FROM bs.namespace_user_id
		AND bc.namespace_org_id IS NOT DISTINCT FROM bs.namespace_org_id
	-- We filter to specs that belong to batch changes where last_applied_at IS NULL, as this
	-- is our way of distinguishing DRAFT batch changes. The migrations that caused this
	-- regression deleted and then reconstructed batch specs based on this condition, so we
	-- know the empty ones we want to re-ID are included in here.
	WHERE bc.last_applied_at IS NULL
	-- The ordering doesn't actually matter, this just makes the behavior more predictable.
	ORDER BY
		bc.id DESC,
		spec_id DESC`

// This query compares the count empty batch specs whose IDs are the min spec ID for their
// batch change vs. the total count of empty batch specs.
const emptySpecIDMigratorProgressQuery = `
WITH specs AS (%s)
SELECT
	CASE total.count WHEN 0 THEN 1 ELSE
		CAST(migrated.count AS float) / CAST(total.count AS float)
	END
FROM
(SELECT COUNT(*) FROM specs WHERE is_empty) AS total,
(SELECT COUNT(*) FROM specs
	JOIN (
		SELECT bc_id, MIN(spec_id) AS min_spec_id
		FROM specs
		GROUP BY bc_id
	) AS mids ON mids.bc_id = specs.bc_id
	WHERE is_empty
	AND spec_id = min_spec_id) migrated;`

func (m *emptySpecIDMigrator) Up(ctx context.Context) (err error) {
	// If a batch change has multiple empty batch specs associated with it, first we clear
	// out the dupes, so that we're only needing to reorder one record.
	if err = m.store.Exec(ctx, sqlf.Sprintf(deleteDupEmptySpecsQuery, sqlf.Sprintf(specsForDraftsQuery))); err != nil {
		return err
	}
	if err = m.store.Exec(ctx, sqlf.Sprintf(nextAvailableIDFunctionQuery)); err != nil {
		return err
	}
	if err = m.store.Exec(ctx, sqlf.Sprintf(emptySpecIDMigratorUpdateQuery, sqlf.Sprintf(specsForDraftsQuery))); err != nil {
		return err
	}

	return nil
}

const deleteDupEmptySpecsQuery = `
WITH specs AS (%s),
-- From just the empty specs...
empty_specs AS (SELECT * FROM specs WHERE is_empty)
DELETE FROM batch_specs
	-- ...delete any empty spec that shares a batch change ID with another empty spec...
	WHERE batch_specs.id IN (SELECT spec_id FROM empty_specs
		WHERE EXISTS (
			SELECT 1 FROM empty_specs s
			WHERE s.bc_id = empty_specs.bc_id
		)
	-- ...so long as it's not the one that's currently applied to the batch change.
	) AND batch_specs.id NOT IN (SELECT batch_spec_id FROM batch_changes)`

const nextAvailableIDFunctionQuery = `
-- The function next_available_id takes a starting_id and returns the first id lower than
-- starting_id which is unused on the batch_specs. The arg starting_id represents the batch
-- spec ID we need to beat to re-order an "empty" batch spec record before any non-empty spec
-- records on the same batch change. For values of starting_id we will be calling this with,
-- we know at least one lower id must exist; this would be the id of the original empty batch
-- spec that was deleted in the original migration.
CREATE OR REPLACE FUNCTION next_available_id (starting_id int8) RETURNS int8 AS $$
	DECLARE
	first_available_id integer := $1;
	BEGIN
		LOOP
			IF NOT EXISTS (SELECT FROM batch_specs WHERE id = first_available_id) THEN
				RETURN first_available_id;
			END IF;
			first_available_id := first_available_id-1;
		END LOOP;
	END;
$$ LANGUAGE plpgsql;
`

const emptySpecIDMigratorUpdateQuery = `
DO
$$
DECLARE spec RECORD;
BEGIN
FOR spec IN
	WITH specs AS (%s),
	to_be_reordered_specs AS (
		SELECT * FROM specs
			-- Tally up how many specs we have for each DRAFT batch change.
			JOIN (
				SELECT
					bc_id,
					COUNT(*) AS spec_count
				FROM
					specs
				GROUP BY
					bc_id) bs_count ON bs_count.bc_id = specs.bc_id
			-- We also look for the lowest spec id for a given batch change, so we know
			-- which id we need to go beneath to achieve the proper ordering.
			JOIN (
				SELECT
					bc_id,
					MIN(spec_id) AS min_id
				FROM
					specs
				GROUP BY
					bc_id) min_spec_id ON min_spec_id.bc_id = specs.bc_id
			WHERE
				-- If a batch change only has a single spec, there's nothing to fix.
				spec_count > 1
				-- If a batch spec is already the lowest id, there's nothing to fix.
				AND spec_id != min_id
				-- We only need to re-id the empty ones.
				AND is_empty
			LIMIT 500
	)
	SELECT * FROM to_be_reordered_specs
	LOOP
		-- Find the next available id lower than the min_id for this batch spec
		DECLARE
			new_spec_id integer := next_available_id(spec.min_id);
			new_rand_id text := spec.rand_id;
		BEGIN
			-- First, update the existing batch spec's rand id, so we don't violate
			-- the unique constraint when we copy the row.
			UPDATE batch_specs SET rand_id = rand_id || '_temp' WHERE id = spec.spec_id;
			-- Copy the spec into a row with the new id.
			INSERT INTO batch_specs (
				id,
				rand_id,
				raw_spec,
				spec,
				namespace_user_id,
				namespace_org_id,
				user_id,
				created_at,
				updated_at,
				created_from_raw,
				allow_unsupported,
				allow_ignored,
				no_cache,
				batch_change_id
			) VALUES (
				new_spec_id,
				new_rand_id,
				spec.raw_spec,
				spec.spec,
				spec.namespace_user_id,
				spec.namespace_org_id,
				spec.user_id,
				spec.created_at,
				spec.updated_at,
				spec.created_from_raw,
				spec.allow_unsupported,
				spec.allow_ignored,
				spec.no_cache,
				spec.batch_change_id
			);
			-- Update any batch change with the old spec applied to use the new one.
			UPDATE batch_changes SET batch_spec_id = new_spec_id WHERE batch_spec_id = spec.spec_id;
			-- Finally, delete the old batch spec.
			DELETE FROM batch_specs WHERE id = spec.spec_id;
		END;
	END LOOP;
END
$$;
`

func (m *emptySpecIDMigrator) Down(ctx context.Context) (err error) {
	// Non-destructive
	return nil
}
