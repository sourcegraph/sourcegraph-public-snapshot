CREATE TABLE access_tokens (
       id bigserial NOT NULL PRIMARY KEY,
       user_id integer NOT NULL REFERENCES users (id),
       value_sha256 bytea NOT NULL UNIQUE,
       note text NOT NULL,
       created_at timestamp with time zone NOT NULL DEFAULT now(),
       last_used_at timestamp with time zone,
       deleted_at timestamp with time zone
);
CREATE INDEX access_tokens_lookup ON access_tokens USING hash(value_sha256) WHERE deleted_at IS NULL;
