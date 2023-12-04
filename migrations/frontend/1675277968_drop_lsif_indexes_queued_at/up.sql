CREATE INDEX CONCURRENTLY IF NOT EXISTS lsif_indexes_queued_at_id ON lsif_indexes(queued_at DESC, id);
