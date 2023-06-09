-- Undo the changes made in the up migration

ALTER TABLE repo_embedding_jobs
DROP COLUMN IF EXISTS stat_has_ranks,
DROP COLUMN IF EXISTS stat_is_incremental,
DROP COLUMN IF EXISTS stat_code_files_total,
DROP COLUMN IF EXISTS stat_code_files_embedded,
DROP COLUMN IF EXISTS stat_code_chunks_embedded,
DROP COLUMN IF EXISTS stat_code_files_skipped,
DROP COLUMN IF EXISTS stat_code_bytes_skipped,
DROP COLUMN IF EXISTS stat_code_bytes_embedded,
DROP COLUMN IF EXISTS stat_text_files_total,
DROP COLUMN IF EXISTS stat_text_files_embedded,
DROP COLUMN IF EXISTS stat_text_chunks_embedded,
DROP COLUMN IF EXISTS stat_text_files_skipped,
DROP COLUMN IF EXISTS stat_text_bytes_skipped,
DROP COLUMN IF EXISTS stat_text_bytes_embedded;
