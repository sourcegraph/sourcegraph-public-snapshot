BEGIN;

-- Refer to the up migration for full details on what we're doing: we're just
-- doing all of that in reverse.

ALTER TABLE campaigns
    ALTER COLUMN last_applier_id DROP NOT NULL,
    ALTER COLUMN last_applied_at DROP NOT NULL,
    ALTER COLUMN campaign_spec_id DROP NOT NULL;

INSERT INTO
    changesets
SELECT
    *
FROM
    changesets_old;

INSERT INTO
    campaigns
SELECT
    *
FROM
    campaigns_old;

DROP TABLE IF EXISTS changesets_old;
DROP TABLE IF EXISTS campaigns_old;

COMMIT;
