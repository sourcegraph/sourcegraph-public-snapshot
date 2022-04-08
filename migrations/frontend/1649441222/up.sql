CREATE TABLE IF NOT EXISTS lsif_uploads_audit_logs (
    instant             timestamptz DEFAULT NOW(),
    upload_id           integer not null,
    commit              text not null,
    root                text not null,
    repository_id       integer not null,
    uploaded_at         timestamptz not null,
    indexer             text not null,
    indexer_version     text,
    upload_size         integer,
    associated_index_id integer,
    committed_at        timestamptz,
    transition_columns  hstore[]
);

CREATE INDEX lsif_uploads_audit_logs_timestamp ON lsif_uploads_audit_logs USING brin (instant);

CREATE OR REPLACE FUNCTION func_lsif_uploads_update() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO lsif_uploads_audit_logs
        (upload_id, commit, root, repository_id, uploaded_at,
        indexer, indexer_version, upload_size, associated_index_id,
        committed_at, transition_columns)
        VALUES (
            NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
            NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
            NEW.committed_at,
            -- array || NULL should be a noop, but that doesn't seem to be happening
            -- hence array_remove here
            array_remove(
                ARRAY[]::hstore[] ||
                CASE WHEN OLD.state IS DISTINCT FROM NEW.state THEN
                    hstore(ARRAY['column', 'state', 'old', OLD.state, 'new', NEW.state])
                    ELSE NULL
                END ||
                CASE WHEN OLD.expired IS DISTINCT FROM NEW.expired THEN
                    hstore(ARRAY['column', 'expired', 'old', OLD.expired::text, 'new', NEW.expired::text])
                    ELSE NULL
                END ||
                CASE WHEN OLD.num_resets IS DISTINCT FROM NEW.num_resets THEN
                    hstore(ARRAY['column', 'num_resets', 'old', OLD.num_resets::text, 'new', NEW.num_resets::text])
                    ELSE NULL
                END ||
                CASE WHEN OLD.num_failures IS DISTINCT FROM NEW.num_failures THEN
                    hstore(ARRAY['column', 'num_failures', 'old', OLD.num_failures::text, 'new', NEW.num_failures::text])
                    ELSE NULL
                END ||
                CASE WHEN OLD.worker_hostname IS DISTINCT FROM NEW.worker_hostname THEN
                    hstore(ARRAY['column', 'worker_hostname', 'old', OLD.worker_hostname, 'new', NEW.worker_hostname])
                    ELSE NULL
                END ||
                CASE WHEN OLD.committed_at IS DISTINCT FROM NEW.committed_at THEN
                    hstore(ARRAY['column', 'committed_at', 'old', OLD.committed_at::text, 'new', NEW.committed_at::text])
                    ELSE NULL
                END,
            NULL)
        );
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

-- CREATE OR REPLACE only supported in Postgres 14+
DROP TRIGGER IF EXISTS trigger_lsif_uploads_update on lsif_uploads;

CREATE TRIGGER trigger_lsif_uploads_update
AFTER UPDATE OF state, num_resets, num_failures, worker_hostname, expired, committed_at
ON lsif_uploads
FOR EACH ROW
EXECUTE FUNCTION func_lsif_uploads_update();
