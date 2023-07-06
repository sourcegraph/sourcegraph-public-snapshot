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
   file_id TEXT NOT NULL, -- this should be text, depending on the blob store we use it might not always be numeric
   file_type TEXT DEFAULT 'html' NOT NULL
);
