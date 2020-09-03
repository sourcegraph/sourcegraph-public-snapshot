BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm
DROP TRIGGER IF EXISTS trig_invalidate_session_on_password_change_or_reset ON users;
DROP FUNCTION IF EXISTS invalidate_session_for_userid_on_password_change_or_reset;
ALTER TABLE IF EXISTS users DROP COLUMN IF EXISTS invalidated_sessions_at;

COMMIT;
