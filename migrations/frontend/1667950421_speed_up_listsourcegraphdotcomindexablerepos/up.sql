CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_dotcom_indexable_repos_idx ON repo (stars DESC NULLS LAST) INCLUDE (id, name)
WHERE
    deleted_at IS NULL AND blocked IS NULL AND (
        (repo.stars >= 5 AND NOT COALESCE(fork, false) AND NOT archived)
        OR
        (lower(repo.name) ~ '^(src\.fedoraproject\.org|maven|npm|jdk)')
    );
