BEGIN;

CREATE TABLE rules (
	id bigserial PRIMARY KEY,
	project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
	settings text NOT NULL
);
CREATE INDEX rules_project_id ON rules(project_id);

COMMIT;
