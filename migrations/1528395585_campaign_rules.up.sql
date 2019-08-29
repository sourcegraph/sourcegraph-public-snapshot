BEGIN;

ALTER TABLE threads ADD COLUMN is_draft boolean NOT NULL DEFAULT false;

ALTER TABLE campaigns ADD COLUMN template_id text;
ALTER TABLE campaigns ADD COLUMN template_context text;
ALTER TABLE campaigns ADD COLUMN is_draft boolean NOT NULL DEFAULT false;
ALTER TABLE campaigns ADD COLUMN start_date timestamp with time zone;
ALTER TABLE campaigns ADD COLUMN due_date timestamp with time zone;

CREATE TABLE rules (
	id bigserial PRIMARY KEY,
	container_campaign_id bigint REFERENCES campaigns(id) ON DELETE CASCADE,
	container_thread_id bigint REFERENCES threads(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
	definition text NOT NULL,
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	updated_at timestamp with time zone NOT NULL DEFAULT now()
);
CREATE INDEX rules_container_campaign_id ON rules(container_campaign_id);
CREATE INDEX rules_container_thread_id ON rules(container_thread_id);

ALTER TABLE events ADD COLUMN rule_id bigint REFERENCES rules(id) ON DELETE CASCADE;

COMMIT;
