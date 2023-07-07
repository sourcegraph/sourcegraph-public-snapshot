CREATE TABLE IF NOT EXISTS file_embedding_jobs (
   id SERIAL PRIMARY KEY,

   -- worker-related columns
   state TEXT DEFAULT 'queued',
   failure_message TEXT,
   queued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
   started_at TIMESTAMP WITH TIME ZONE,
   finished_at TIMESTAMP WITH TIME ZONE,
   process_after TIMESTAMP WITH TIME ZONE,
   num_resets INTEGER NOT NULL DEFAULT 0,
   num_failures INTEGER NOT NULL DEFAULT 0,
   last_heartbeat_at TIMESTAMP WITH TIME ZONE,
   execution_logs JSON [],
   worker_hostname TEXT NOT NULL DEFAULT '',
   cancel boolean DEFAULT FALSE NOT NULL,

   -- file metadata columns
   archive_id TEXT NOT NULL, -- this should be text, depending on the blob store we use it might not always be numeric
   file_type TEXT DEFAULT 'html' NOT NULL
);

CREATE TABLE IF NOT EXISTS file_embedding_job_stats (
   job_id INTEGER PRIMARY KEY REFERENCES file_embedding_jobs(id) ON DELETE CASCADE DEFERRABLE,
   is_incremental BOOLEAN NOT NULL DEFAULT FALSE,
   files_total INTEGER NOT NULL DEFAULT 0,
   files_embedded INTEGER NOT NULL DEFAULT 0,
   chunks_embedded INTEGER NOT NULL DEFAULT 0,
   files_skipped JSONB NOT NULL DEFAULT '{}',
   bytes_embedded INTEGER NOT NULL DEFAULT 0
);
