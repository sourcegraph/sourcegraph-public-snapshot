BEGIN;

CREATE TABLE IF NOT EXISTS campaigns_old (
    id BIGINT,
    name TEXT,
    description TEXT,
    initial_applier_id INTEGER,
    namespace_user_id INTEGER,
    namespace_org_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    changeset_ids JSONB,
    closed_at TIMESTAMP WITH TIME ZONE,
    campaign_spec_id BIGINT,
    last_applier_id BIGINT,
    last_applied_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS changesets_old (
    id BIGINT,
    campaign_ids JSONB,
    repo_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    external_id TEXT,
    external_service_type TEXT,
    external_deleted_at TIMESTAMP WITH TIME ZONE,
    external_branch TEXT,
    external_updated_at TIMESTAMP WITH TIME ZONE,
    external_state TEXT,
    external_review_state TEXT,
    external_check_state TEXT,
    created_by_campaign BOOLEAN,
    added_to_campaign BOOLEAN,
    diff_stat_added INTEGER,
    diff_stat_changed INTEGER,
    diff_stat_deleted INTEGER,
    sync_state JSONB,
    current_spec_id BIGINT,
    previous_spec_id BIGINT,
    publication_state TEXT,
    owned_by_campaign_id BIGINT,
    reconciler_state TEXT,
    failure_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    process_after TIMESTAMP WITH TIME ZONE,
    num_resets INTEGER
);

COMMIT;
