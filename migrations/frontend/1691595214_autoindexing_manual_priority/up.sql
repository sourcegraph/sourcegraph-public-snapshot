ALTER TABLE lsif_indexes
ADD COLUMN IF NOT EXISTS enqueuer_user_id integer NOT NULL DEFAULT 0;
