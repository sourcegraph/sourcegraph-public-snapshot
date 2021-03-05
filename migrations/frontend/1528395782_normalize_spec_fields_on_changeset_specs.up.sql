BEGIN;

ALTER TABLE changeset_specs ADD COLUMN head_ref text DEFAULT NULL;
ALTER TABLE changeset_specs ADD COLUMN title text DEFAULT NULL;
ALTER TABLE changeset_specs ADD COLUMN external_id text DEFAULT NULL;

CREATE INDEX changeset_specs_external_id ON changeset_specs (external_id);
CREATE INDEX changeset_specs_head_ref ON changeset_specs (head_ref);
CREATE INDEX changeset_specs_title ON changeset_specs (title);

UPDATE changeset_specs SET head_ref = changeset_specs.spec->>'headRef'::text WHERE changeset_specs.spec->>'externalID'::text IS NULL;
UPDATE changeset_specs SET title = changeset_specs.spec->>'title'::text;
UPDATE changeset_specs SET external_id = changeset_specs.spec->>'externalID'::text;

-- These are the same views as before, except that they now use the new
-- columns, which is why we need to re-create them.
DROP VIEW IF EXISTS branch_changeset_specs_and_changesets;
CREATE VIEW branch_changeset_specs_and_changesets AS (
    SELECT
        changeset_specs.id AS changeset_spec_id,
        COALESCE(changesets.id, 0::bigint) AS changeset_id,
        changeset_specs.repo_id,
        changeset_specs.campaign_spec_id,
        changesets.owned_by_campaign_id AS owner_campaign_id,
        repo.name AS repo_name,
        changeset_specs.title AS changeset_name
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

DROP VIEW IF EXISTS tracking_changeset_specs_and_changesets;
CREATE VIEW tracking_changeset_specs_and_changesets AS (  
    SELECT
        changeset_specs.id AS changeset_spec_id,
        COALESCE(changesets.id, 0::bigint) AS changeset_id,
        changeset_specs.repo_id,
        changeset_specs.campaign_spec_id,
        repo.name AS repo_name,
        COALESCE(changesets.metadata ->> 'Title'::text, changesets.metadata ->> 'title'::text) AS changeset_name
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

COMMIT;
