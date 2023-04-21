DROP TRIGGER IF EXISTS update_own_aggregate_recent_contribution ON own_signal_recent_contribution;
DROP FUNCTION IF EXISTS update_own_aggregate_recent_contribution();

DROP INDEX IF EXISTS own_aggregate_recent_contribution_file_author;
DROP TABLE IF EXISTS own_aggregate_recent_contribution;

DROP TABLE IF EXISTS own_signal_recent_contribution;

DROP INDEX IF EXISTS commit_authors_email_name;
DROP TABLE IF EXISTS commit_authors;

DROP INDEX IF EXISTS repo_paths_index_absolute_path;
DROP TABLE IF EXISTS repo_paths;
