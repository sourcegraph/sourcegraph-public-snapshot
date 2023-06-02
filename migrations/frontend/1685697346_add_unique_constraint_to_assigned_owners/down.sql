DROP INDEX IF EXISTS assigned_owners_file_path_owner;

CREATE INDEX IF NOT EXISTS assigned_owners_file_path
    ON assigned_owners
        USING btree (file_path_id);
