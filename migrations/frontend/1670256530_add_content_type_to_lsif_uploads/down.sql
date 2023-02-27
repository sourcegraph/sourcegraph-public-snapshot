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
            operation, transition_columns)
            VALUES (
                COALESCE(current_setting('codeintel.lsif_uploads_audit.reason', true), ''),
                NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
                NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
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
        operation, transition_columns)
        VALUES (
            NEW.id, NEW.commit, NEW.root, NEW.repository_id, NEW.uploaded_at,
            NEW.indexer, NEW.indexer_version, NEW.upload_size, NEW.associated_index_id,
            'create', func_lsif_uploads_transition_columns_diff(
                (NULL, NULL, NULL, NULL, NULL, NULL),
                func_row_to_lsif_uploads_transition_columns(NEW)
            )
        );
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

DROP VIEW IF EXISTS lsif_uploads_with_repository_name;
ALTER TABLE lsif_uploads DROP COLUMN IF EXISTS content_type;
ALTER TABLE lsif_uploads_audit_logs DROP COLUMN IF EXISTS content_type;

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
    u.expired,
    u.last_retention_scan_at,
    r.name AS repository_name,
    u.uncompressed_size
FROM lsif_uploads u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;
