DROP TABLE IF EXISTS codeintel_ranking_references_processed;
ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS processed BOOLEAN NOT NULL DEFAULT false;
