BEGIN;

CREATE TABLE IF NOT EXISTS insight_view_grants
(
    id              SERIAL
        CONSTRAINT insight_view_grants_pk
            PRIMARY KEY,
    insight_view_id INTEGER NOT NULL
        CONSTRAINT insight_view_grants_insight_view_id_fk
            REFERENCES insight_view
            ON DELETE CASCADE, -- These grants only have meaning in the context of a parent view.
    user_id         INTEGER,
    org_id          INTEGER,
    global          BOOLEAN
);

COMMENT ON TABLE insight_view_grants IS 'Permission grants for insight views. Each row should represent a unique principal (user, org, etc).';
COMMENT ON COLUMN insight_view_grants.user_id IS 'User ID that that receives this grant.';
COMMENT ON COLUMN insight_view_grants.org_id IS 'Org ID that that receives this grant.';
COMMENT ON COLUMN insight_view_grants.global IS 'Grant that does not belong to any specific principal and is granted to all users.';

CREATE INDEX IF NOT EXISTS insight_view_grants_insight_view_id_index
    ON insight_view_grants (insight_view_id);

CREATE INDEX IF NOT EXISTS insight_view_grants_user_id_idx
    ON insight_view_grants (user_id);

CREATE INDEX IF NOT EXISTS insight_view_grants_org_id_idx
    ON insight_view_grants (org_id);

CREATE INDEX IF NOT EXISTS insight_view_grants_global_idx
    ON insight_view_grants (global) WHERE global IS TRUE;


-- This series join table is completely dependent on the existence of a parent view. So to simplify db operations
-- and avoid dangling rows, adding cascade deletes to the insight view FK.
ALTER TABLE insight_view_series
    DROP CONSTRAINT IF EXISTS insight_view_series_insight_view_id_fkey;

ALTER TABLE insight_view_series
    ADD CONSTRAINT insight_view_series_insight_view_id_fkey
        FOREIGN KEY (insight_view_id) REFERENCES insight_view
            ON DELETE CASCADE;

COMMIT;
