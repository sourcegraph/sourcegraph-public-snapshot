SELECT remote_exec('_lsif', '
    DROP VIEW IF EXISTS "commitWithLsifMarkers";
    DROP INDEX IF EXISTS "commits_repo_commit_parentCommit_unique";
    DROP INDEX IF EXISTS "commits_repo_commit";
    DROP INDEX IF EXISTS "commits_repo_parentCommit";
    DROP TABLE IF EXISTS "commits";
    DROP TABLE IF EXISTS "lsifDataMarkers";
');
