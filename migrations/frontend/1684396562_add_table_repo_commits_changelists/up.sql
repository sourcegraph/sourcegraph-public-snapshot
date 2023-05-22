CREATE TABLE IF NOT EXISTS repo_commits_changelists (
    id SERIAL PRIMARY KEY,
    repo_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
    commit_sha bytea NOT NULL,
    perforce_changelist_id integer NOT NULL,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS repo_id_perforce_changelist_id_unique ON repo_commits_changelists USING btree (repo_id, perforce_changelist_id);
