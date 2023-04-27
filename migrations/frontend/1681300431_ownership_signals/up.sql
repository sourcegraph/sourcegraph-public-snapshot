CREATE TABLE IF NOT EXISTS repo_paths (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
    absolute_path TEXT NOT NULL,
    parent_id INTEGER NULL REFERENCES repo_paths(id)
);

COMMENT ON COLUMN repo_paths.absolute_path
IS 'Absolute path does not start or end with forward slash. Example: "a/b/c". Root directory is empty path "".';

CREATE UNIQUE INDEX IF NOT EXISTS repo_paths_index_absolute_path
ON repo_paths
USING btree (repo_id, absolute_path);

CREATE TABLE IF NOT EXISTS commit_authors (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS commit_authors_email_name
ON commit_authors
USING btree (email, name);

CREATE TABLE IF NOT EXISTS own_signal_recent_contribution (
    id SERIAL PRIMARY KEY,
    commit_author_id INTEGER NOT NULL REFERENCES commit_authors(id),
    changed_file_path_id INTEGER NOT NULL REFERENCES repo_paths(id),
    commit_timestamp TIMESTAMP NOT NULL,
    commit_id bytea NOT NULL
);

COMMENT ON TABLE own_signal_recent_contribution
IS 'One entry per file changed in every commit that classifies as a contribution signal.';

CREATE TABLE IF NOT EXISTS own_aggregate_recent_contribution (
    id SERIAL PRIMARY KEY,
    commit_author_id INTEGER NOT NULL REFERENCES commit_authors(id),
    changed_file_path_id INTEGER NOT NULL REFERENCES repo_paths(id),
    contributions_count INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS own_aggregate_recent_contribution_file_author
ON own_aggregate_recent_contribution
USING btree (changed_file_path_id, commit_author_id);

CREATE OR REPLACE FUNCTION update_own_aggregate_recent_contribution() RETURNS TRIGGER AS $$
BEGIN
    WITH RECURSIVE ancestors AS (
        SELECT id, parent_id, 1 AS level
        FROM repo_paths
        WHERE id = NEW.changed_file_path_id
        UNION ALL
        SELECT p.id, p.parent_id, a.level + 1
        FROM repo_paths p
        JOIN ancestors a ON p.id = a.parent_id
    )
    UPDATE own_aggregate_recent_contribution
    SET contributions_count = contributions_count + 1
    WHERE commit_author_id = NEW.commit_author_id AND changed_file_path_id IN (
        SELECT id FROM ancestors
    );

    WITH RECURSIVE ancestors AS (
        SELECT id, parent_id, 1 AS level
        FROM repo_paths
        WHERE id = NEW.changed_file_path_id
        UNION ALL
        SELECT p.id, p.parent_id, a.level + 1
        FROM repo_paths p
        JOIN ancestors a ON p.id = a.parent_id
    )
    INSERT INTO own_aggregate_recent_contribution (commit_author_id, changed_file_path_id, contributions_count)
    SELECT NEW.commit_author_id, id, 1
    FROM ancestors
    WHERE NOT EXISTS (
        SELECT 1 FROM own_aggregate_recent_contribution
        WHERE commit_author_id = NEW.commit_author_id AND changed_file_path_id = ancestors.id
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    trigger_exists INTEGER;
BEGIN
    -- Check if the trigger already exists
    SELECT COUNT(*)
    INTO trigger_exists
    FROM pg_trigger
    WHERE tgname = 'update_own_aggregate_recent_contribution';

    -- If the trigger exists, drop it
    IF trigger_exists > 0 THEN
        EXECUTE 'DROP TRIGGER update_own_aggregate_recent_contribution ON own_signal_recent_contribution';
    END IF;

    -- Create the trigger
    EXECUTE 'CREATE TRIGGER update_own_aggregate_recent_contribution
        AFTER INSERT
        ON own_signal_recent_contribution
        FOR EACH ROW
        EXECUTE FUNCTION update_own_aggregate_recent_contribution()';
END $$;
