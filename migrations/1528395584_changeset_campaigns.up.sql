BEGIN;

CREATE TABLE changeset_campaigns (
	id bigserial PRIMARY KEY,
	project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text
);
CREATE INDEX changeset_campaigns_project_id ON changeset_campaigns(project_id);

COMMIT;
