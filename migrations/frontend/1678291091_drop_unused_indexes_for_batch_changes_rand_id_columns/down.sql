CREATE INDEX IF NOT EXISTS batch_specs_rand_id ON batch_specs USING btree (rand_id);
CREATE INDEX IF NOT EXISTS changeset_specs_rand_id ON changeset_specs USING btree (rand_id);
