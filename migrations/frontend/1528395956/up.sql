CREATE TABLE IF NOT EXISTS org_stats
(
    org_id INTEGER
        REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE
            PRIMARY KEY,
    code_host_repo_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE org_stats IS 'Business statistics for organizations';
COMMENT ON COLUMN org_stats.org_id IS 'Org ID that the stats relate to.';
COMMENT ON COLUMN org_stats.code_host_repo_count IS 'Count of repositories accessible on all code hosts for this organization.';
