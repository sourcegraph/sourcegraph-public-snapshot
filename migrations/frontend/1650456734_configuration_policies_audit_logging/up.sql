CREATE TABLE IF NOT EXISTS configuration_policies_audit_logs (
    -- log entry columns
    log_timestamp      timestamptz DEFAULT clock_timestamp(),
    record_deleted_at  timestamptz,
    -- associated object columns
    policy_id          integer not null,
    transition_columns hstore[]
);

COMMENT ON COLUMN configuration_policies_audit_logs.log_timestamp IS 'Timestamp for this log entry.';
COMMENT ON COLUMN configuration_policies_audit_logs.record_deleted_at IS 'Set once the upload this entry is associated with is deleted. Once NOW() - record_deleted_at is above a certain threshold, this log entry will be deleted.';
COMMENT ON COLUMN configuration_policies_audit_logs.transition_columns IS 'Array of changes that occurred to the upload for this entry, in the form of {"column"=>"<column name>", "old"=>"<previous value>", "new"=>"<new value>"}.';

CREATE INDEX IF NOT EXISTS configuration_policies_audit_logs_policy_id ON configuration_policies_audit_logs USING btree (policy_id);
CREATE INDEX IF NOT EXISTS configuration_policies_audit_logs_timestamp ON configuration_policies_audit_logs USING brin (log_timestamp);

-- CREATE TYPE IF NOT EXISTS is not a thing, so we must drop it here,
-- but also all the functions that reference the type...
DROP FUNCTION IF EXISTS func_row_to_configuration_policies_transition_columns;
DROP FUNCTION IF EXISTS func_configuration_policies_transition_columns_diff;
DROP TYPE IF EXISTS configuration_policies_transition_columns;

CREATE TYPE configuration_policies_transition_columns AS (
    name                        text,
    type                        text,
    pattern                     text,
    retention_enabled           boolean,
    retention_duration_hours    integer,
    retain_intermediate_commits boolean,
    indexing_enabled            boolean,
    index_commit_max_age_hours  integer,
    index_intermediate_commits  boolean,
    protected                   boolean,
    repository_patterns         text[]
);

COMMENT ON TYPE configuration_policies_transition_columns IS 'A type containing the columns that make-up the set of tracked transition columns. Primarily used to create a nulled record due to `OLD` being unset in INSERT queries, and creating a nulled record with a subquery is not allowed.';

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

CREATE OR REPLACE FUNCTION func_configuration_policies_delete() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE configuration_policies_audit_logs
        SET record_deleted_at = NOW()
        WHERE policy_id IN (
            SELECT id FROM OLD
        );

        RETURN NULL;
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


CREATE OR REPLACE FUNCTION func_row_to_configuration_policies_transition_columns(
    rec record
) RETURNS configuration_policies_transition_columns AS $$
    BEGIN
        RETURN (
            rec.name, rec.type, rec.pattern,
            rec.retention_enabled, rec.retention_duration_hours, rec.retain_intermediate_commits,
            rec.indexing_enabled, rec.index_commit_max_age_hours, rec.index_intermediate_commits,
            rec.protected, rec.repository_patterns);
    END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION func_configuration_policies_insert IS 'Transforms a record from the configuration_policies table into an `configuration_policies_transition_columns` type variable.';

CREATE OR REPLACE FUNCTION func_configuration_policies_transition_columns_diff(
    old configuration_policies_transition_columns, new configuration_policies_transition_columns
) RETURNS hstore[] AS $$
    BEGIN
        -- array || NULL should be a noop, but that doesn't seem to be happening
        -- hence array_remove here
        RETURN array_remove(
            ARRAY[]::hstore[] ||
            CASE WHEN old.name IS DISTINCT FROM new.name THEN
                hstore(ARRAY['column', 'name', 'old', old.name, 'new', new.name])
                ELSE NULL
            END ||
            CASE WHEN old.type IS DISTINCT FROM new.type THEN
                hstore(ARRAY['column', 'type', 'old', old.type, 'new', new.type])
                ELSE NULL
            END ||
            CASE WHEN old.pattern IS DISTINCT FROM new.pattern THEN
                hstore(ARRAY['column', 'pattern', 'old', old.pattern, 'new', new.pattern])
                ELSE NULL
            END ||
            CASE WHEN old.retention_enabled IS DISTINCT FROM new.retention_enabled THEN
                hstore(ARRAY['column', 'retention_enabled', 'old', old.retention_enabled::text, 'new', new.retention_enabled::text])
                ELSE NULL
            END ||
            CASE WHEN old.retention_duration_hours IS DISTINCT FROM new.retention_duration_hours THEN
                hstore(ARRAY['column', 'retention_duration_hours', 'old', old.retention_duration_hours::text, 'new', new.retention_duration_hours::text])
                ELSE NULL
            END ||
            CASE WHEN old.indexing_enabled IS DISTINCT FROM new.indexing_enabled THEN
                hstore(ARRAY['column', 'indexing_enabled', 'old', old.indexing_enabled::text, 'new', new.indexing_enabled::text])
                ELSE NULL
            END ||
            CASE WHEN old.index_commit_max_age_hours IS DISTINCT FROM new.index_commit_max_age_hours THEN
                hstore(ARRAY['column', 'index_commit_max_age_hours', 'old', old.index_commit_max_age_hours::text, 'new', new.index_commit_max_age_hours::text])
                ELSE NULL
            END ||
            CASE WHEN old.index_intermediate_commits IS DISTINCT FROM new.index_intermediate_commits THEN
                hstore(ARRAY['column', 'index_intermediate_commits', 'old', old.index_intermediate_commits::text, 'new', new.index_intermediate_commits::text])
                ELSE NULL
            END ||
            CASE WHEN old.protected IS DISTINCT FROM new.protected THEN
                hstore(ARRAY['column', 'protected', 'old', old.protected::text, 'new', new.protected::text])
                ELSE NULL
            END ||
            CASE WHEN old.repository_patterns IS DISTINCT FROM new.repository_patterns THEN
                hstore(ARRAY['column', 'repository_patterns', 'old', old.repository_patterns::text, 'new', new.repository_patterns::text])
                ELSE NULL
            END,
        NULL);
    END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION func_configuration_policies_transition_columns_diff IS 'Diffs two `configuration_policies_transition_columns` values into an array of hstores, where each hstore is in the format {"column"=>"<column name>", "old"=>"<previous value>", "new"=>"<new value>"}.';

-- CREATE OR REPLACE only supported in Postgres 14+
DROP TRIGGER IF EXISTS trigger_configuration_policies_update ON lsif_configuration_policies;
DROP TRIGGER IF EXISTS trigger_configuration_policies_delete ON lsif_configuration_policies;
DROP TRIGGER IF EXISTS trigger_configuration_policies_insert ON lsif_configuration_policies;

CREATE TRIGGER trigger_configuration_policies_update
BEFORE UPDATE OF name, pattern, retention_enabled, retention_duration_hours, type, retain_intermediate_commits
ON lsif_configuration_policies
FOR EACH ROW
EXECUTE FUNCTION func_configuration_policies_update();

CREATE TRIGGER trigger_configuration_policies_delete
AFTER DELETE
ON lsif_configuration_policies
REFERENCING OLD TABLE AS OLD
FOR EACH STATEMENT
EXECUTE FUNCTION func_configuration_policies_delete();

CREATE TRIGGER trigger_configuration_policies_insert
AFTER INSERT
ON lsif_configuration_policies
FOR EACH ROW
EXECUTE FUNCTION func_configuration_policies_insert();
