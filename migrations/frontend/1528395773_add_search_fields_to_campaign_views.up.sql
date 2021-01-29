BEGIN;

DROP VIEW IF EXISTS branch_changeset_specs_and_changesets;
DROP VIEW IF EXISTS tracking_changeset_specs_and_changesets;

CREATE VIEW tracking_changeset_specs_and_changesets AS (
    SELECT
        changeset_specs.id AS changeset_spec_id,
        COALESCE(changesets.id, 0) AS changeset_id,
        changeset_specs.repo_id AS repo_id,
        changeset_specs.campaign_spec_id AS campaign_spec_id,
        repo.name AS repo_name,
        COALESCE(changesets.metadata->>'Title', changesets.metadata->>'title') AS changeset_name
    FROM changeset_specs
    LEFT JOIN changesets ON
        changesets.repo_id = changeset_specs.repo_id AND
        changesets.external_id = changeset_specs.spec->>'externalID'
    INNER JOIN repo ON changeset_specs.repo_id = repo.id
    WHERE
        changeset_specs.spec->>'externalID' IS NOT NULL AND
        repo.deleted_at IS NULL
);

CREATE VIEW branch_changeset_specs_and_changesets AS (
    SELECT
        changeset_specs.id AS changeset_spec_id,
        COALESCE(changesets.id, 0) AS changeset_id,
        changeset_specs.repo_id AS repo_id,
        changeset_specs.campaign_spec_id AS campaign_spec_id,
        changesets.owned_by_campaign_id AS owner_campaign_id,
        repo.name AS repo_name,
        changeset_specs.spec->>'title' AS changeset_name
    FROM changeset_specs
    LEFT JOIN changesets ON
        changesets.repo_id = changeset_specs.repo_id AND
        changesets.current_spec_id IS NOT NULL AND
        (SELECT spec FROM changeset_specs WHERE changeset_specs.id = changesets.current_spec_id)->>'headRef' = changeset_specs.spec->>'headRef'
    INNER JOIN repo ON changeset_specs.repo_id = repo.id
    WHERE
        changeset_specs.spec->>'externalID' IS NULL AND
        repo.deleted_at IS NULL
);

COMMIT;
