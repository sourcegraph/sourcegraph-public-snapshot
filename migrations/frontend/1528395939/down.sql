BEGIN;

ALTER TABLE changeset_specs
    DROP CONSTRAINT changeset_specs_batch_spec_id_fkey,
    ADD CONSTRAINT changeset_specs_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs (id) DEFERRABLE;

COMMIT;
