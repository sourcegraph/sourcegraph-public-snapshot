-- CREATE TABLE IF NOT EXISTS own_stats (
--     file_path_id INTEGER NOT NULL PRIMARY KEY REFERENCES repo_paths(id) ON DELETE CASCADE DEFERRABLE,
--     codeowners_id INTEGER NULL REFERENCES commit_authors(id),
--     assigned_owner_id INTEGER NULL REFERENCES users(id),
--     deep_files_owned_count INTEGER,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
-- );
-- CREATE UNIQUE INDEX IF NOT EXISTS own_stats_file_codeowners
-- ON own_stats
-- USING btree (file_path_id, codeowners_id)
-- WHERE codeowners_id IS NOT NULL;
-- CREATE UNIQUE INDEX IF NOT EXISTS own_stats_file_assigned
-- ON own_stats
-- USING btree (file_path_od, assigned_owner_id)
-- WHERE
ALTER TABLE IF EXISTS repo_paths
ADD COLUMN IF NOT EXISTS deep_file_count INTEGER NULL;
