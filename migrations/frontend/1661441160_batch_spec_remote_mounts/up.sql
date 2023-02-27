CREATE TABLE IF NOT EXISTS batch_spec_workspace_files
(
    id            serial PRIMARY KEY,
    rand_id       text                                   NOT NULL,
    batch_spec_id bigint                                 NOT NULL,
    filename      text                                   NOT NULL,
    path          text                                   NOT NULL,
    size          bigint                                 NOT NULL,
    content       bytea                                  NOT NULL,
    modified_at   timestamp with time zone               NOT NULL,
    created_at    timestamp with time zone DEFAULT now() NOT NULL,
    updated_at    timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE batch_spec_workspace_files
    DROP CONSTRAINT IF EXISTS batch_spec_workspace_files_batch_spec_id_fkey;

ALTER TABLE ONLY batch_spec_workspace_files
    ADD CONSTRAINT batch_spec_workspace_files_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS batch_spec_workspace_files_rand_id ON batch_spec_workspace_files USING btree (rand_id);

CREATE UNIQUE INDEX IF NOT EXISTS batch_spec_workspace_files_batch_spec_id_filename_path ON batch_spec_workspace_files (batch_spec_id, filename, path);
