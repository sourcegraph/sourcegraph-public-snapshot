ALTER TABLE lsif_uploads ADD COLUMN IF NOT EXISTS should_reindex boolean NOT NULL DEFAULT false;
