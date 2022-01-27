BEGIN;

CREATE TABLE IF NOT EXISTS dashboard
(
    id                 SERIAL                  NOT NULL CONSTRAINT dashboard_pk PRIMARY KEY,
    title              TEXT,
    created_at         TIMESTAMP DEFAULT NOW() NOT NULL,
    created_by_user_id INT,
    last_updated_at    TIMESTAMP DEFAULT NOW() NOT NULL,
    deleted_at         TIMESTAMP
);

COMMENT ON TABLE dashboard IS 'Metadata for dashboards of insights';

COMMENT ON COLUMN dashboard.title IS 'Title of the dashboard';

COMMENT ON COLUMN dashboard.created_at IS 'Timestamp the dashboard was initially created.';

COMMENT ON COLUMN dashboard.created_by_user_id IS 'User that created the dashboard, if available.';

COMMENT ON COLUMN dashboard.last_updated_at IS 'Time the dashboard was last updated, either metadata or insights.';

COMMENT ON COLUMN dashboard.deleted_at IS 'Set to the time the dashboard was soft deleted.';



CREATE TABLE IF NOT EXISTS dashboard_grants
(
    id           SERIAL CONSTRAINT dashboard_grants_pk PRIMARY KEY,
    dashboard_id INTEGER NOT NULL CONSTRAINT dashboard_grants_dashboard_id_fk REFERENCES dashboard ON DELETE CASCADE, -- These grants only have meaning in the context of a parent dashboard.
    user_id      INTEGER,
    org_id       INTEGER,
    global       BOOLEAN
);

COMMENT ON TABLE dashboard_grants IS 'Permission grants for dashboards. Each row should represent a unique principal (user, org, etc).';
COMMENT ON COLUMN dashboard_grants.user_id IS 'User ID that that receives this grant.';
COMMENT ON COLUMN dashboard_grants.org_id IS 'Org ID that that receives this grant.';
COMMENT ON COLUMN dashboard_grants.global IS 'Grant that does not belong to any specific principal and is granted to all users.';

CREATE INDEX IF NOT EXISTS dashboard_grants_dashboard_id_index
    ON dashboard_grants (dashboard_id);

CREATE INDEX IF NOT EXISTS dashboard_grants_user_id_idx
    ON dashboard_grants (user_id);

CREATE INDEX IF NOT EXISTS dashboard_grants_org_id_idx
    ON dashboard_grants (org_id);

CREATE INDEX IF NOT EXISTS dashboard_grants_global_idx
    ON dashboard_grants (global) WHERE global IS TRUE;

CREATE TABLE dashboard_insight_view
(
    id              SERIAL NOT NULL CONSTRAINT dashboard_insight_view_pk PRIMARY KEY,
    dashboard_id    INT    NOT NULL CONSTRAINT dashboard_insight_view_dashboard_id_fk REFERENCES dashboard (id) ON DELETE CASCADE,
    insight_view_id INT    NOT NULL CONSTRAINT dashboard_insight_view_insight_view_id_fk REFERENCES insight_view (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS dashboard_insight_view_insight_view_id_fk_idx ON dashboard_insight_view (insight_view_id);
CREATE INDEX IF NOT EXISTS dashboard_insight_view_dashboard_id_fk_idx ON dashboard_insight_view (dashboard_id);

COMMIT;
