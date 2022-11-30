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

CREATE INDEX IF NOT EXISTS repo_directories_index_absolute_path ON repo_directories USING btree (repo_id, absolute_path);

CREATE TABLE IF NOT EXISTS repo_file_contents (
    id SERIAL PRIMARY KEY,
    text_contents TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS repo_files (
    id SERIAL PRIMARY KEY,
    directory_id integer NOT NULL REFERENCES repo_directories(id) ON DELETE CASCADE DEFERRABLE,
    version_id integer NOT NULL REFERENCES repo_versions(id) ON DELETE CASCADE DEFERRABLE,
    base_name text NOT NULL,
    content_id integer NOT NULL REFERENCES repo_file_contents(id) ON DELETE CASCADE DEFERRABLE
);

CREATE INDEX IF NOT EXISTS repo_files_directory ON repo_files USING btree (directory_id);
CREATE INDEX IF NOT EXISTS repo_files_version ON repo_files USING btree (version_id);
