-- +++
-- parent: 1528395877
-- +++

CREATE INDEX CONCURRENTLY IF NOT EXISTS lsif_indexes_repository_id_commit ON lsif_indexes (repository_id, commit);
