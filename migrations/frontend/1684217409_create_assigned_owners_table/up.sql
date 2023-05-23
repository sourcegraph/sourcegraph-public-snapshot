CREATE TABLE IF NOT EXISTS assigned_owners
(
    id                   SERIAL PRIMARY KEY,
    owner_user_id        INTEGER   NOT NULL REFERENCES users (id) ON DELETE CASCADE DEFERRABLE,
    file_path_id         INTEGER   NOT NULL REFERENCES repo_paths (id),
    who_assigned_user_id INTEGER   NULL REFERENCES users (id) ON DELETE SET NULL DEFERRABLE,
    assigned_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS assigned_owners_file_path
    ON assigned_owners
        USING btree (file_path_id);

COMMENT ON TABLE assigned_owners
    IS 'Table for ownership assignments, one entry contains an assigned user ID, which repo_path is assigned and the date and user who assigned the owner.';
