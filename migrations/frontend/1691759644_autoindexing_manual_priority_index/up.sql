CREATE INDEX CONCURRENTLY IF NOT EXISTS
lsif_indexes_dequeue_order_idx ON lsif_indexes
USING btree ((enqueuer_user_id > 0) DESC, queued_at DESC, id)
WHERE (state = 'queued' OR state = 'errored');
