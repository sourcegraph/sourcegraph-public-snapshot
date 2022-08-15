CREATE TABLE IF NOT EXISTS batch_spec_mounts
(
    id            serial PRIMARY KEY,
    rand_id       text                                   NOT NULL,
    batch_spec_id bigint                                 NOT NULL,
    filename      text                                   NOT NULL,
    path          text                                   NOT NULL,
    size          bigint                                 NOT NULL,
    modified      timestamp with time zone               NOT NULL,
    created_at    timestamp with time zone DEFAULT now() NOT NULL,
    updated_at    timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY batch_spec_mounts
    ADD CONSTRAINT batch_spec_mounts_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs (id);

CREATE INDEX batch_spec_mounts_rand_id ON batch_spec_mounts USING btree (rand_id);

CREATE UNIQUE INDEX IF NOT EXISTS batch_spec_mounts_batch_spec_id_filename_path ON batch_spec_mounts (batch_spec_id, filename, path);
