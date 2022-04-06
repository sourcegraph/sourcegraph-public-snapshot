-- Undo the changes made in the up migration

ALTER TABLE cm_trigger_jobs
    ADD COLUMN results BOOLEAN;
ALTER TABLE cm_trigger_jobs
    ADD COLUMN num_results INTEGER;
