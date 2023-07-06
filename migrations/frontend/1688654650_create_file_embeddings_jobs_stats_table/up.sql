CREATE TABLE IF NOT EXISTS file_embedding_job_stats (
   job_id INTEGER PRIMARY KEY REFERENCES file_embedding_jobs(id) ON DELETE CASCADE DEFERRABLE,
   is_incremental BOOLEAN NOT NULL DEFAULT FALSE,
   files_total INTEGER NOT NULL DEFAULT 0,
   files_embedded INTEGER NOT NULL DEFAULT 0,
   chunks_embedded INTEGER NOT NULL DEFAULT 0,
   files_skipped JSONB NOT NULL DEFAULT '{}',
   bytes_embedded INTEGER NOT NULL DEFAULT 0
);
