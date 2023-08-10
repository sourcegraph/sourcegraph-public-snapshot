DELETE INDEX IF EXISTS lsif_indexes_dequeue_order_idx ON lsif_indexes;

ALTER TABLE lsif_indexes
DROP COLUMN IF EXISTS enqueuer_user_id;
