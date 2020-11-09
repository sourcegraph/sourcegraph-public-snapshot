BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE IF NOT EXISTS cm_monitors
(
    id                BIGSERIAL PRIMARY KEY,
    created_by        int4        NOT NULL,
    created_at        timestamptz NOT NULL DEFAULT now(),
    description       text        NOT NULL,
    changed_at        timestamptz NOT NULL DEFAULT now(),
    changed_by        int4        NOT NULL,
    enabled           boolean     NOT NULL DEFAULT TRUE,
    namespace_user_id int4,
    namespace_org_id  int4,
    constraint cm_monitors_user_id_fk
        foreign key (namespace_user_id)
            REFERENCES users (id)
            ON DELETE CASCADE,
    constraint cm_monitors_org_id_fk
        foreign key (namespace_org_id)
            REFERENCES orgs (id)
            ON DELETE CASCADE,
    constraint cm_monitors_created_by_fk
        foreign key (created_by)
            REFERENCES users (id)
            ON DELETE CASCADE,
    constraint cm_monitors_changed_by_fk
        foreign key (changed_by)
            REFERENCES users (id)
            ON DELETE cascade
);
COMMIT;
