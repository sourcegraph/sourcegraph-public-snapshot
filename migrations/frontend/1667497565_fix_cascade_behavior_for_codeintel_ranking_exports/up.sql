-- Change primary key
ALTER TABLE codeintel_ranking_exports DROP CONSTRAINT IF EXISTS codeintel_ranking_exports_pkey;
ALTER TABLE codeintel_ranking_exports ADD COLUMN IF NOT EXISTS id SERIAL NOT NULL;
ALTER TABLE codeintel_ranking_exports ADD PRIMARY KEY (id);

-- Make column + foreign key nullable
ALTER TABLE codeintel_ranking_exports DROP CONSTRAINT IF EXISTS codeintel_ranking_exports_upload_id_fkey;
ALTER TABLE codeintel_ranking_exports ALTER COLUMN upload_id DROP NOT NULL;
ALTER TABLE codeintel_ranking_exports ADD CONSTRAINT codeintel_ranking_exports_upload_id_fkey FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE SET NULL;

-- Emulate old primary key index
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_ranking_exports_upload_id_graph_key ON codeintel_ranking_exports(upload_id, graph_key);

-- Add prefix to control authoritative GCS paths as we might set upload_id to NULL now
ALTER TABLE codeintel_ranking_exports ADD COLUMN IF NOT EXISTS object_prefix TEXT;
