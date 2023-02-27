DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_log_operation') THEN
        -- delete is known by record_deleted_at
        CREATE TYPE audit_log_operation AS ENUM('create', 'modify', 'delete');
    END IF;
END
$$;

CREATE SEQUENCE IF NOT EXISTS lsif_uploads_audit_logs_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE IF NOT EXISTS configuration_policies_audit_logs_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- table not been used yet
TRUNCATE TABLE lsif_uploads_audit_logs;
TRUNCATE TABLE configuration_policies_audit_logs;

ALTER TABLE lsif_uploads_audit_logs
ADD COLUMN IF NOT EXISTS sequence BIGINT NOT NULL DEFAULT nextval('lsif_uploads_audit_logs_seq'::regclass),
ADD COLUMN IF NOT EXISTS operation audit_log_operation NOT NULL;

ALTER TABLE configuration_policies_audit_logs
ADD COLUMN IF NOT EXISTS sequence BIGINT NOT NULL DEFAULT nextval('configuration_policies_audit_logs_seq'::regclass),
ADD COLUMN IF NOT EXISTS operation audit_log_operation NOT NULL;

ALTER SEQUENCE lsif_uploads_audit_logs_seq OWNED BY lsif_uploads_audit_logs.sequence;
ALTER SEQUENCE configuration_policies_audit_logs_seq OWNED BY configuration_policies_audit_logs.sequence;

-- Start replace triggers

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
            (policy_id, operation, transition_columns)
            VALUES (NEW.id, 'modify', diff);
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION func_configuration_policies_insert() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO configuration_policies_audit_logs
        (policy_id, operation, transition_columns)
        VALUES (
            NEW.id, 'create',
            func_configuration_policies_transition_columns_diff(
                (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
                func_row_to_configuration_policies_transition_columns(NEW)
            )
        );
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

-- End replace triggers
