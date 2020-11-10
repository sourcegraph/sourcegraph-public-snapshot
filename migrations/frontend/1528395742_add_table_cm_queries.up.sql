BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE IF NOT EXISTS cm_queries
(
    id         BIGSERIAL PRIMARY KEY,
    monitor    int8        NOT NULL,
    query      text        NOT NULL,
    created_by int4        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    changed_by int4        NOT NULL,
    changed_at timestamptz NOT NULL DEFAULT now(),
    constraint cm_triggers_monitor
        foreign key (monitor)
            REFERENCES cm_monitors (id)
            ON DELETE CASCADE,
    constraint cm_triggers_created_by_fk
        foreign key (created_by)
            REFERENCES users (id)
            ON DELETE CASCADE,
    constraint cm_triggers_changed_by_fk
        foreign key (changed_by)
            REFERENCES users (id)
            ON DELETE cascade
);
COMMIT;
