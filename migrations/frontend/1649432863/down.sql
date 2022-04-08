-- Undo the changes made in the up migration

ALTER TABLE cm_last_searched
    DROP CONSTRAINT cm_last_searched_pkey,
    DROP COLUMN repo_id,
    ADD COLUMN IF NOT EXISTS args_hash bigint NOT NULL,
    ADD PRIMARY KEY (monitor_id, args_hash);
