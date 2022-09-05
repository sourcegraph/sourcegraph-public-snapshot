CREATE TABLE IF NOT EXISTS lsif_uploads_audit_logs (
    -- log entry columns
    log_timestamp       timestamptz DEFAULT NOW(),
    record_deleted_at   timestamptz,
    -- associated object columns
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

COMMENT ON COLUMN lsif_uploads_audit_logs.log_timestamp IS 'Timestamp for this log entry.';
COMMENT ON COLUMN lsif_uploads_audit_logs.record_deleted_at IS 'Set once the upload this entry is associated with is deleted. Once NOW() - record_deleted_at is above a certain threshold, this log entry will be deleted.';
COMMENT ON COLUMN lsif_uploads_audit_logs.transition_columns IS 'Array of changes that occurred to the upload for this entry, in the form of {"column"=>"<column name>", "old"=>"<previous value>", "new"=>"<new value>"}.';

CREATE INDEX IF NOT EXISTS lsif_uploads_audit_logs_upload_id ON lsif_uploads_audit_logs USING btree (upload_id);
CREATE INDEX IF NOT EXISTS lsif_uploads_audit_logs_timestamp ON lsif_uploads_audit_logs USING brin (log_timestamp);

-- CREATE TYPE IF NOT EXISTS is not a thing, so we must drop it here,
-- but also all the functions that reference the type...
DROP FUNCTION IF EXISTS func_row_to_lsif_uploads_transition_columns;
DROP FUNCTION IF EXISTS func_lsif_uploads_transition_columns_diff;
DROP TYPE IF EXISTS lsif_uploads_transition_columns;

CREATE TYPE lsif_uploads_transition_columns AS (
    state           text,
    expired         boolean,
    num_resets      integer,
    num_failures    integer,
    worker_hostname text,
    committed_at    timestamptz
);

COMMENT ON TYPE lsif_uploads_transition_columns IS 'A type containing the columns that make-up the set of tracked transition columns. Primarily used to create a nulled record due to `OLD` being unset in INSERT queries, and creating a nulled record with a subquery is not allowed.';

CREATE OR REPLACE FUNCTION func_lsif_uploads_update() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO lsif_uploads_audit_logs
        (upload_id, commit, root, repository_id, uploaded_at,
        indexer, indexer_version, upload_size, associated_index_id,
        committed_at, transition_columns)
        VALUES (
            NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
            NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
            NEW.committed_at, func_lsif_uploads_transition_columns_diff(
                func_row_to_lsif_uploads_transition_columns(OLD),
                func_row_to_lsif_uploads_transition_columns(NEW)
            )
        );

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION func_lsif_uploads_delete() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE lsif_uploads_audit_logs
        SET record_deleted_at = NOW()
        WHERE upload_id IN (
            SELECT id FROM OLD
        );

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION func_lsif_uploads_insert() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO lsif_uploads_audit_logs
        (upload_id, commit, root, repository_id, uploaded_at,
        indexer, indexer_version, upload_size, associated_index_id,
        committed_at, transition_columns)
        VALUES (
            NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
            NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
            NEW.committed_at, func_lsif_uploads_transition_columns_diff(
                (NULL, NULL, NULL, NULL, NULL, NULL),
                func_row_to_lsif_uploads_transition_columns(NEW)
            )
        );
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION func_row_to_lsif_uploads_transition_columns(
    rec record
) RETURNS lsif_uploads_transition_columns AS $$
    BEGIN
        RETURN (rec.state, rec.expired, rec.num_resets, rec.num_failures, rec.worker_hostname, rec.committed_at);
    END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION func_lsif_uploads_insert IS 'Transforms a record from the lsif_uploads table into an `lsif_uploads_transition_columns` type variable.';

CREATE OR REPLACE FUNCTION func_lsif_uploads_transition_columns_diff(
    old lsif_uploads_transition_columns, new lsif_uploads_transition_columns
) RETURNS hstore[] AS $$
    BEGIN
        -- array || NULL should be a noop, but that doesn't seem to be happening
        -- hence array_remove here
        RETURN array_remove(
            ARRAY[]::hstore[] ||
            CASE WHEN old.state IS DISTINCT FROM new.state THEN
                hstore(ARRAY['column', 'state', 'old', old.state, 'new', new.state])
                ELSE NULL
            END ||
            CASE WHEN old.expired IS DISTINCT FROM new.expired THEN
                hstore(ARRAY['column', 'expired', 'old', old.expired::text, 'new', new.expired::text])
                ELSE NULL
            END ||
            CASE WHEN old.num_resets IS DISTINCT FROM new.num_resets THEN
                hstore(ARRAY['column', 'num_resets', 'old', old.num_resets::text, 'new', new.num_resets::text])
                ELSE NULL
            END ||
            CASE WHEN old.num_failures IS DISTINCT FROM new.num_failures THEN
                hstore(ARRAY['column', 'num_failures', 'old', old.num_failures::text, 'new', new.num_failures::text])
                ELSE NULL
            END ||
            CASE WHEN old.worker_hostname IS DISTINCT FROM new.worker_hostname THEN
                hstore(ARRAY['column', 'worker_hostname', 'old', old.worker_hostname, 'new', new.worker_hostname])
                ELSE NULL
            END ||
            CASE WHEN old.committed_at IS DISTINCT FROM new.committed_at THEN
                hstore(ARRAY['column', 'committed_at', 'old', old.committed_at::text, 'new', new.committed_at::text])
                ELSE NULL
            END,
        NULL);
    END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION func_lsif_uploads_transition_columns_diff IS 'Diffs two `lsif_uploads_transition_columns` values into an array of hstores, where each hstore is in the format {"column"=>"<column name>", "old"=>"<previous value>", "new"=>"<new value>"}.';

-- CREATE OR REPLACE only supported in Postgres 14+
DROP TRIGGER IF EXISTS trigger_lsif_uploads_update ON lsif_uploads;
DROP TRIGGER IF EXISTS trigger_lsif_uploads_delete ON lsif_uploads;
DROP TRIGGER IF EXISTS trigger_lsif_uploads_insert ON lsif_uploads;

CREATE TRIGGER trigger_lsif_uploads_update
BEFORE UPDATE OF state, num_resets, num_failures, worker_hostname, expired, committed_at
ON lsif_uploads
FOR EACH ROW
EXECUTE FUNCTION func_lsif_uploads_update();

CREATE TRIGGER trigger_lsif_uploads_delete
AFTER DELETE
ON lsif_uploads
REFERENCING OLD TABLE AS OLD
FOR EACH STATEMENT
EXECUTE FUNCTION func_lsif_uploads_delete();

CREATE TRIGGER trigger_lsif_uploads_insert
AFTER INSERT
ON lsif_uploads
FOR EACH ROW
EXECUTE FUNCTION func_lsif_uploads_insert();
