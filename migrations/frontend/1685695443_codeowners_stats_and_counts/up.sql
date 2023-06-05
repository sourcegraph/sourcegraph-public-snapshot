CREATE TABLE IF NOT EXISTS codeowners_owners (
    id SERIAL NOT NULL PRIMARY KEY,
    reference TEXT NOT NULL
);

COMMENT ON TABLE codeowners_owners IS 'Text reference in CODEOWNERS entry to use in codeowners_individual_stats. Reference is either email or handle without @ in front.';
COMMENT ON COLUMN codeowners_owners.reference IS 'We just keep the reference as opposed to splitting it to handle or email
since the distinction is not relevant for query, and this makes indexing way easier.';

CREATE INDEX IF NOT EXISTS codeowners_owners_reference ON codeowners_owners USING btree (reference);

CREATE TABLE IF NOT EXISTS ownership_path_stats (
    file_path_id INTEGER NOT NULL PRIMARY KEY REFERENCES repo_paths(id),
    tree_codeowned_files_count INTEGER NULL,
    last_updated_at TIMESTAMP NOT NULL
);

COMMENT ON TABLE ownership_path_stats IS 'Data on how many files in given tree are owned by anyone.

We choose to have a table for `ownership_path_stats` - more general than for CODEOWNERS,
with a specific tree_codeowned_files_count CODEOWNERS column. The reason for that
is that we aim at expanding path stats by including total owned files (via CODEOWNERS
or assigned ownership), and perhaps files count by assigned ownership only.';
COMMENT ON COLUMN ownership_path_stats.last_updated_at IS 'When the last background job updating counts run.';


CREATE TABLE IF NOT EXISTS codeowners_individual_stats (
    file_path_id INTEGER NOT NULL REFERENCES repo_paths(id),
    owner_id INTEGER NOT NULL REFERENCES codeowners_owners(id),
    tree_owned_files_count INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    -- We choose a compound primary key as the counts are looked up by either file_path_id only
    -- or by the pair.
    PRIMARY KEY (file_path_id, owner_id)
);

COMMENT ON TABLE codeowners_individual_stats IS 'Data on how many files in given tree are owned by given owner.

As opposed to ownership-general `ownership_path_stats` table, the individual <path x owner> stats
are stored in CODEOWNERS-specific table `codeowners_individual_stats`. The reason for that is that
we are also indexing on owner_id which is CODEOWNERS-specific.';
COMMENT ON COLUMN codeowners_individual_stats.tree_owned_files_count IS 'Total owned file count by given owner at given file tree.';
COMMENT ON COLUMN codeowners_individual_stats.updated_at IS 'When the last background job updating counts run.';

ALTER TABLE IF EXISTS repo_paths
ADD COLUMN IF NOT EXISTS tree_files_count INTEGER NULL,
    ADD COLUMN IF NOT EXISTS tree_files_counts_updated_at TIMESTAMP NULL;

COMMENT ON COLUMN repo_paths.tree_files_count IS 'Total count of files in the file tree rooted at the path. 1 for files.';
COMMENT ON COLUMN repo_paths.tree_files_counts_updated_at IS 'Timestamp of the job that updated the file counts';