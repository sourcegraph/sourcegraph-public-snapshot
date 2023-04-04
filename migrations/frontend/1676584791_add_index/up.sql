CREATE INDEX CONCURRENTLY IF NOT EXISTS lsif_indexes ON lsif_indexes(repository_id, commit, root, indexer) WHERE state = 'completed';
