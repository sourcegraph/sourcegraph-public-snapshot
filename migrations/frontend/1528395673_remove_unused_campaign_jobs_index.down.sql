BEGIN;

CREATE INDEX patches_patch_set_id ON patches USING btree (patch_set_id);

COMMIT;
