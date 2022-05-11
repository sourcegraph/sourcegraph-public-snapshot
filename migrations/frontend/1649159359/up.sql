ALTER TABLE batch_spec_resolution_jobs
    ADD COLUMN IF NOT EXISTS initiator_id integer,
    ADD FOREIGN KEY (initiator_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE NO ACTION DEFERRABLE;

UPDATE batch_spec_resolution_jobs SET initiator_id = (SELECT bs.user_id FROM batch_specs bs WHERE bs.id = batch_spec_id);

