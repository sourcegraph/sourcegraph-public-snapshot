-- Note: DROP INDEX CONCURRENTLY cannot run inside a transaction block

DROP INDEX CONCURRENTLY IF EXISTS repos_created_at;

