ALTER TABLE lsif_uploads_audit_logs
ADD COLUMN IF NOT EXISTS reason text DEFAULT '',
DROP COLUMN IF EXISTS committed_at;

COMMENT ON COLUMN lsif_uploads_audit_logs.reason IS 'The reason/source for this entry.';

CREATE OR REPLACE FUNCTION func_lsif_uploads_update() RETURNS TRIGGER AS $$
    DECLARE
        diff hstore[];
    BEGIN
        diff = func_lsif_uploads_transition_columns_diff(
            func_row_to_lsif_uploads_transition_columns(OLD),
            func_row_to_lsif_uploads_transition_columns(NEW)
        );

        IF (array_length(diff, 1) > 0) THEN
            INSERT INTO lsif_uploads_audit_logs
            (reason, upload_id, commit, root, repository_id, uploaded_at,
            indexer, indexer_version, upload_size, associated_index_id,
            transition_columns)
            VALUES (
                COALESCE(current_setting('codeintel.lsif_uploads_audit.reason', true), ''),
                NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
                NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
                diff
            );
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION func_lsif_uploads_insert() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO lsif_uploads_audit_logs
        (upload_id, commit, root, repository_id, uploaded_at,
        indexer, indexer_version, upload_size, associated_index_id,
        transition_columns)
        VALUES (
            NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
            NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
            func_lsif_uploads_transition_columns_diff(
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
