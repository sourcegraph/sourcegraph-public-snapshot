ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS reference_cursor_export_deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS reference_cursor_export_id INTEGER;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS path_cursor_deleted_export_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS path_cursor_export_id INTEGER;
