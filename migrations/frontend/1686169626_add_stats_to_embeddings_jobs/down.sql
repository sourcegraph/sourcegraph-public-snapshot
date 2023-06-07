-- Undo the changes made in the up migration

ALTER TABLE repo_embedding_jobs
DROP COLUMN stat_has_ranks,
DROP COLUMN stat_is_incremental,
DROP COLUMN stat_code_files_total,
DROP COLUMN stat_code_files_embedded,
DROP COLUMN stat_code_chunks_embedded,
DROP COLUMN stat_code_files_skipped,
DROP COLUMN stat_code_bytes_skipped,
DROP COLUMN stat_code_bytes_embedded,
DROP COLUMN stat_text_files_total,
DROP COLUMN stat_text_files_embedded,
DROP COLUMN stat_text_chunks_embedded,
DROP COLUMN stat_text_files_skipped,
DROP COLUMN stat_text_bytes_skipped,
DROP COLUMN stat_text_bytes_embedded;
