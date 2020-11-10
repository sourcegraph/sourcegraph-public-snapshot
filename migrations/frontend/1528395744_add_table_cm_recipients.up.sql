BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE IF NOT EXISTS cm_recipients
(
    id                BIGSERIAL PRIMARY KEY,
    email             int8 NOT NULL,
    namespace_user_id int4,
    namespace_org_id  int4,
    constraint cm_recipients_emails
        foreign key (email)
            REFERENCES cm_emails (id)
            ON DELETE CASCADE,
    constraint cm_recipients_user_id_fk
        foreign key (namespace_user_id)
            REFERENCES users (id)
            ON DELETE CASCADE,
    constraint cm_recipients_org_id_fk
        foreign key (namespace_org_id)
            REFERENCES orgs (id)
            ON DELETE CASCADE
);
COMMIT;
