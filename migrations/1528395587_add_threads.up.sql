BEGIN;

CREATE TABLE threads (
       id bigserial PRIMARY KEY,
       repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
       title text NOT NULL,
       external_url text,
       status text NOT NULL
);
CREATE INDEX threads_repository_id ON threads(repository_id);

CREATE TABLE threads_changeset (
       thread_id bigint PRIMARY KEY REFERENCES threads(id) ON DELETE SET NULL,
       is_preview boolean NOT NULL,
       base_ref text NOT NULL,
       head_ref text NOT NULL
);

ALTER TABLE threads ADD COLUMN changeset_id bigint REFERENCES threads_changeset(thread_id) ON DELETE CASCADE;

COMMIT;
