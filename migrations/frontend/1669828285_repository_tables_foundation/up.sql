CREATE TABLE IF NOT EXISTS repo_versions (
    id SERIAL PRIMARY KEY,
    repo_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
    -- Externally referencable identifier like a git sha
    external_id text NOT NULL,
    -- For computing reachability using path cover
    path_cover_color integer NOT NULL,
    path_cover_index integer NOT NULL,
    path_cover_reachability jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

-- TODO should this possibly be just on external_id ?
CREATE UNIQUE INDEX IF NOT EXISTS repo_versions_external_id ON repo_versions USING btree (repo_id, external_id);

CREATE TABLE IF NOT EXISTS repo_directories (
    id SERIAL PRIMARY KEY,
    repo_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
    -- Absolute path does not start with a forward slash, and does not end with a slash, so a/b/c is an absolute path.
    absolute_path text NOT NULL,
    -- TODO should this be on delete cascade as well?
    parent_id integer NULL REFERENCES repo_directories(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS repo_directories_index_absolute_path ON repo_directories USING btree (repo_id, absolute_path);

CREATE TABLE IF NOT EXISTS repo_file_contents (
    id SERIAL PRIMARY KEY,
    text_contents TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS repo_files (
    id SERIAL PRIMARY KEY,
    directory_id integer NOT NULL REFERENCES repo_directories(id) ON DELETE CASCADE DEFERRABLE,
    version_id integer NOT NULL REFERENCES repo_versions(id) ON DELETE CASCADE DEFERRABLE,
    topological_order integer NOT NULL,
    base_name text NOT NULL,
    content_id integer NOT NULL REFERENCES repo_file_contents(id) ON DELETE CASCADE DEFERRABLE
);

CREATE UNIQUE INDEX IF NOT EXISTS repo_files_uq ON repo_files USING btree (directory_id, version_id, base_name);
CREATE INDEX IF NOT EXISTS repo_files_directory ON repo_files USING btree (directory_id);
CREATE INDEX IF NOT EXISTS repo_files_version ON repo_files USING btree (version_id);


-- Example of how to find all files and their specific commits reachable from given version
-- SELECT DISTINCT ON (f.directory_id, f.base_name)
--   f.directory_id,
--   f.base_name,
--   fv.external_id
-- FROM
--   repo_files AS f
--   INNER JOIN repo_versions AS fv ON f.version_id = fv.id
--   INNER JOIN repo_versions AS v ON (v.path_cover_reachability ->> fv.path_cover_color)::integer >= fv.path_cover_index
-- WHERE
--   v.repo_id = 1
--   AND v.external_id = 'foo'
-- ORDER BY
--   f.directory_id,
--   f.base_name,
--   f.topological_order DESC;
