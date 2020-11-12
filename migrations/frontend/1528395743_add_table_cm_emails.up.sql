BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TYPE cm_email_priority AS ENUM ('NORMAL', 'CRITICAL');

CREATE TABLE IF NOT EXISTS cm_emails
(
    id                BIGSERIAL PRIMARY KEY,
    monitor           int8              NOT NULL,
    enabled           boolean           NOT NULL,
    priority          cm_email_priority NOT NULL,
    header            text              NOT NULL,
    created_by        int4              NOT NULL,
    created_at        timestamptz       NOT NULL DEFAULT now(),
    changed_by        int4              NOT NULL,
    changed_at        timestamptz       NOT NULL DEFAULT now(),
    constraint cm_emails_monitor
        foreign key (monitor)
            REFERENCES cm_monitors (id)
            ON DELETE CASCADE,
    constraint cm_emails_created_by_fk
        foreign key (created_by)
            REFERENCES users (id)
            ON DELETE CASCADE,
    constraint cm_emails_changed_by_fk
        foreign key (changed_by)
            REFERENCES users (id)
            ON DELETE cascade
);
COMMIT;
