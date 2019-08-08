BEGIN;

ALTER TABLE discussion_threads ADD COLUMN settings text;

ALTER TABLE discussion_threads ADD COLUMN type text;
UPDATE discussion_threads SET type='THREAD';
ALTER TABLE discussion_threads ALTER COLUMN type SET NOT NULL;

ALTER TABLE discussion_threads ADD COLUMN status text;
UPDATE discussion_threads SET status=(CASE WHEN archived_at IS NULL THEN 'OPEN_ACTIVE' ELSE 'CLOSED' END);
ALTER TABLE discussion_threads ALTER COLUMN status SET NOT NULL;

ALTER TABLE discussion_threads_target_repo ADD COLUMN is_ignored boolean NOT NULL DEFAULT false;

CREATE TABLE projects (
       id bigserial PRIMARY KEY,
       namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE,
       namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,
       name citext NOT NULL
);
ALTER TABLE projects ADD CONSTRAINT projects_has_1_namespace CHECK ((namespace_user_id IS NULL) != (namespace_org_id IS NULL));
CREATE INDEX projects_namespace_user_id ON projects(namespace_user_id);
CREATE INDEX projects_namespace_org_id ON projects(namespace_org_id);
CREATE UNIQUE INDEX projects_name ON projects(name);

-- TODO!(sqs): what to do with threads that are not in a project? NEEDS MIGRATION
ALTER TABLE discussion_threads ADD COLUMN project_id integer NOT NULL REFERENCES projects(id) ON DELETE CASCADE;
CREATE INDEX discussion_threads_project_id ON discussion_threads(project_id);

COMMIT;
