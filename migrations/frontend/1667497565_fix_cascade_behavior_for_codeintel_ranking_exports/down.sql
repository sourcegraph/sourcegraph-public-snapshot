-- Drop dependent constraints
ALTER TABLE codeintel_ranking_exports DROP CONSTRAINT IF EXISTS codeintel_ranking_exports_pkey;

-- Make column + foreign key non-nullable
ALTER TABLE codeintel_ranking_exports DROP CONSTRAINT IF EXISTS codeintel_ranking_exports_upload_id_fkey;
DELETE FROM codeintel_ranking_exports WHERE upload_id IS NULL;
ALTER TABLE codeintel_ranking_exports ALTER COLUMN upload_id SET NOT NULL;
ALTER TABLE codeintel_ranking_exports ADD CONSTRAINT codeintel_ranking_exports_upload_id_fkey FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;

-- Restore the old primary key
ALTER TABLE codeintel_ranking_exports DROP COLUMN IF EXISTS id;
ALTER TABLE codeintel_ranking_exports ADD PRIMARY KEY (upload_id, graph_key);

-- Drop now duplicate index
DROP INDEX IF EXISTS codeintel_ranking_exports_upload_id_graph_key;

-- Drop new column
ALTER TABLE codeintel_ranking_exports DROP COLUMN IF EXISTS object_prefix;
