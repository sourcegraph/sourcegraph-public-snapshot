BEGIN;

CREATE TABLE threads (
       id bigserial PRIMARY KEY,
       repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
       title text NOT NULL,
       external_url text,
       settings text,
       type text NOT NULL,
       status text NOT NULL
);
CREATE INDEX threads_repository_id ON threads(repository_id);

COMMIT;
