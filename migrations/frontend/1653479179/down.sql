ALTER TABLE lsif_uploads_audit_logs
DROP COLUMN IF EXISTS sequence,
DROP COLUMN IF EXISTS operation;

DROP SEQUENCE IF EXISTS lsif_uploads_audit_logs_seq;

ALTER TABLE configuration_policies_audit_logs
DROP COLUMN IF EXISTS sequence,
DROP COLUMN IF EXISTS operation;

DROP SEQUENCE IF EXISTS configuration_policies_audit_logs_seq;

DROP TYPE IF EXISTS audit_log_operation;

CREATE OR REPLACE FUNCTION func_configuration_policies_update() RETURNS TRIGGER AS $$
    DECLARE
        diff hstore[];
    BEGIN
        diff = func_configuration_policies_transition_columns_diff(
            func_row_to_configuration_policies_transition_columns(OLD),
            func_row_to_configuration_policies_transition_columns(NEW)
        );

        IF (array_length(diff, 1) > 0) THEN
            INSERT INTO configuration_policies_audit_logs
            (policy_id, transition_columns)
            VALUES (
                NEW.id,
                diff
            );
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION func_configuration_policies_insert() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO configuration_policies_audit_logs
        (policy_id, transition_columns)
        VALUES (
            NEW.id,
            func_configuration_policies_transition_columns_diff(
                (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
                func_row_to_configuration_policies_transition_columns(NEW)
            )
        );
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

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
