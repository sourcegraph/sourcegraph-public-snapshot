ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS last_scanned_at timestamp with time zone;
ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS last_scanned_at timestamp with time zone;
