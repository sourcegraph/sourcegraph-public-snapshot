DROP VIEW IF EXISTS lsif_uploads_with_repository_name;
ALTER TABLE lsif_uploads ADD COLUMN IF NOT EXISTS content_type TEXT NOT NULL DEFAULT 'application/x-ndjson+lsif';
ALTER TABLE lsif_uploads_audit_logs ADD COLUMN IF NOT EXISTS content_type TEXT NOT NULL DEFAULT 'application/x-ndjson+lsif';

COMMENT ON COLUMN lsif_uploads.content_type IS 'The content type of the upload record. For now, the default value is `application/x-ndjson+lsif` to backfill existing records. This will change as we remove LSIF support.';

CREATE VIEW lsif_uploads_with_repository_name AS
SELECT u.id,
    u.commit,
    u.root,
    u.queued_at,
    u.uploaded_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.indexer,
    u.indexer_version,
    u.num_parts,
    u.uploaded_parts,
    u.process_after,
    u.num_resets,
    u.upload_size,
    u.num_failures,
    u.associated_index_id,
    u.content_type,
    u.expired,
    u.last_retention_scan_at,
    r.name AS repository_name,
    u.uncompressed_size
FROM lsif_uploads u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;

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
            content_type,
            operation, transition_columns)
            VALUES (
                COALESCE(current_setting('codeintel.lsif_uploads_audit.reason', true), ''),
                NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
                NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
                NEW.content_type,
                'modify', diff
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
        content_type,
        operation, transition_columns)
        VALUES (
            NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
            NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
            NEW.content_type,
            'create', func_lsif_uploads_transition_columns_diff(
                (NULL, NULL, NULL, NULL, NULL, NULL),
                func_row_to_lsif_uploads_transition_columns(NEW)
            )
        );
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;
