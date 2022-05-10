ALTER TABLE
    IF EXISTS lsif_uploads_audit_logs
    DROP COLUMN IF EXISTS reason,
    ADD COLUMN IF NOT EXISTS committed_at timestamptz;

-- Revert functions to previous version

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

-- CREATE OR REPLACE only supported in Postgres 14+
DROP TRIGGER IF EXISTS trigger_lsif_uploads_update ON lsif_uploads;
DROP TRIGGER IF EXISTS trigger_lsif_uploads_insert ON lsif_uploads;

CREATE TRIGGER trigger_lsif_uploads_update
BEFORE UPDATE OF state, num_resets, num_failures, worker_hostname, expired, committed_at
ON lsif_uploads
FOR EACH ROW
EXECUTE FUNCTION func_lsif_uploads_update();

CREATE TRIGGER trigger_lsif_uploads_insert
AFTER INSERT
ON lsif_uploads
FOR EACH ROW
EXECUTE FUNCTION func_lsif_uploads_insert();
