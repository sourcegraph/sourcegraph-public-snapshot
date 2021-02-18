BEGIN;

DROP VIEW IF EXISTS branch_changeset_specs_and_changesets;
DROP VIEW IF EXISTS tracking_changeset_specs_and_changesets;

CREATE VIEW tracking_changeset_specs_and_changesets AS (
    SELECT
        changeset_specs.id AS changeset_spec_id,
        COALESCE(changesets.id, 0::bigint) AS changeset_id,
        changeset_specs.repo_id,
        changeset_specs.campaign_spec_id,
        repo.name AS repo_name,
        COALESCE(changesets.metadata ->> 'Title'::text, changesets.metadata ->> 'title'::text) AS changeset_name,
        changesets.external_state,
        changesets.publication_state,
        changesets.reconciler_state
    FROM
        changeset_specs
    LEFT JOIN
        changesets
    ON
        changesets.repo_id = changeset_specs.repo_id
        AND
        changesets.external_id = changeset_specs.external_id
    JOIN
        repo
    ON
        changeset_specs.repo_id = repo.id
    WHERE
        changeset_specs.external_id IS NOT NULL
        AND
        repo.deleted_at IS NULL
);

CREATE VIEW branch_changeset_specs_and_changesets AS (
    SELECT
        changeset_specs.id AS changeset_spec_id,
        COALESCE(changesets.id, 0::bigint) AS changeset_id,
        changeset_specs.repo_id,
        changeset_specs.campaign_spec_id,
        changesets.owned_by_campaign_id AS owner_campaign_id,
        repo.name AS repo_name,
        changeset_specs.title AS changeset_name,
        changesets.external_state,
        changesets.publication_state,
        changesets.reconciler_state
    FROM changeset_specs
    LEFT JOIN
        changesets
    ON
        changesets.repo_id = changeset_specs.repo_id
        AND
        changesets.current_spec_id IS NOT NULL
        AND
        EXISTS (
            SELECT 1
            FROM changeset_specs changeset_specs_1
            WHERE
                changeset_specs_1.id = changesets.current_spec_id
                AND
                changeset_specs_1.head_ref = changeset_specs.head_ref
        )
    JOIN
        repo
    ON
        changeset_specs.repo_id = repo.id
    WHERE
        changeset_specs.external_id IS NULL
        AND
        repo.deleted_at IS NULL
);

CREATE INDEX IF NOT EXISTS
    changesets_external_state_idx
ON
    changesets (external_state);

CREATE INDEX IF NOT EXISTS
    changesets_publication_state_idx
ON
    changesets (publication_state);

CREATE INDEX IF NOT EXISTS
    changesets_reconciler_state_idx
ON
    changesets (reconciler_state);

COMMIT;
