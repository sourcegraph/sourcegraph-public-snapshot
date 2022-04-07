-- Undo the changes made in the up migration

ALTER TABLE cm_trigger_jobs
    ADD COLUMN IF NOT EXISTS results BOOLEAN;
ALTER TABLE cm_trigger_jobs
    ADD COLUMN IF NOT EXISTS num_results INTEGER;
