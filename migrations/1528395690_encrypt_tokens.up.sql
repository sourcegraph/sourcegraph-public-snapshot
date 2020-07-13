BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE crypt_secrets (
    id bigint NOT NULL,
    source_type varying(50) NOT NULL,
    source_id bigint NOT NULL,
    value text NOT NULL,
);

CREATE INDEX source_secret_idx ON crypt_secrets USING (source_type, source_id);

COMMIT;
