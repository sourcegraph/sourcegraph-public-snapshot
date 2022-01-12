-- We have a hand-created but non-codified index in our Cloud environment
-- called repo_deleted_at_idx which is a partial btree index over deleted_at
-- where deleted_at is null. This effectively creates a btree index whose only
-- value is NULL.
--
-- Instead, we'll make a partial index on useful fields that can at least help
-- cover queries that select only/filter by id and/or name.

CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_non_deleted_id_name_idx ON repo(id, name) WHERE deleted_at IS NULL;
