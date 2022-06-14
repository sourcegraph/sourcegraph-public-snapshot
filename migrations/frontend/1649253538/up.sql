UPDATE batch_spec_resolution_jobs SET initiator_id = (SELECT bs.user_id FROM batch_specs bs WHERE bs.id = batch_spec_id);

ALTER TABLE batch_spec_resolution_jobs ALTER COLUMN initiator_id SET NOT NULL;
