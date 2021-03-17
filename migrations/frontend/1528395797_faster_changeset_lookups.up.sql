CREATE INDEX CONCURRENTLY IF NOT EXISTS changesets_batch_change_ids ON changesets USING GIN (batch_change_ids jsonb_ops);
