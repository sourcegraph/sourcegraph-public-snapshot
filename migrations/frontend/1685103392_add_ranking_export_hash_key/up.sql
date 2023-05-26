ALTER TABLE codeintel_ranking_exports ADD COLUMN IF NOT EXISTS upload_key TEXT;

UPDATE codeintel_ranking_exports SET upload_key = (
    SELECT md5(u.repository_id || ':' || u.root || ':' || u.indexer)
    FROM lsif_uploads u
    WHERE u.id = upload_id
) WHERE upload_key IS NULL;
