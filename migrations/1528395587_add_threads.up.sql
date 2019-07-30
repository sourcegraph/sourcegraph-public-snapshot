BEGIN;

CREATE TYPE thread_type AS enum ('THREAD', 'ISSUE', 'CHANGESET');

CREATE TABLE threads (
       id bigserial PRIMARY KEY,
       type thread_type NOT NULL,
       repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
       title text NOT NULL,
       external_url text,
       status text NOT NULL,

       -- type == CHANGESET
       is_preview boolean,
       base_ref text,
       head_ref text

);
CREATE INDEX threads_repository_id ON threads(repository_id);

COMMIT;
