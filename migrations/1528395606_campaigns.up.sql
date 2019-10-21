BEGIN;

CREATE TABLE threads (
	id bigserial PRIMARY KEY,
	repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
	title text NOT NULL,
	state text NOT NULL,

    assignee_user_id integer REFERENCES users(id) ON DELETE SET NULL,
    assignee_external_actor_username text,
    assignee_external_actor_url text,

    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),

	base_ref text,
    base_ref_oid text,
    head_repository_id integer REFERENCES repo(id) ON DELETE SET NULL,
	head_ref text,
    head_ref_oid text,

	-- TODO!(sqs): make external_services use hard-delete not soft-delete so that the deletions
	-- cascade to the threads and events rows, which will prevent orphaned data - or else make their
	-- queries omit results if the corresponding external_services row was soft-deleted.
    external_service_id bigint REFERENCES external_services(id) ON DELETE CASCADE,
    external_id text,
    external_metadata jsonb
);
CREATE INDEX threads_repository_id ON threads(repository_id);
CREATE INDEX threads_external_service_id ON threads(external_service_id);
ALTER TABLE threads ADD CONSTRAINT external_thread_has_id_and_data CHECK ((external_service_id IS NULL) = (external_id IS NULL) AND (external_id IS NULL) = (external_metadata IS NULL));
CREATE UNIQUE INDEX threads_external ON threads(external_service_id, external_id) WHERE external_service_id IS NOT NULL;

-----------------

CREATE TABLE exp_campaigns_threads (
	campaign_id bigint NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
	thread_id bigint NOT NULL REFERENCES threads(id) ON DELETE CASCADE
);
CREATE INDEX exp_campaigns_threads_campaign_id ON exp_campaigns_threads(campaign_id);
CREATE INDEX exp_campaigns_threads_thread_id ON exp_campaigns_threads(thread_id) WHERE thread_id IS NOT NULL;
CREATE UNIQUE INDEX exp_campaigns_threads_uniq ON exp_campaigns_threads(campaign_id, thread_id);

-----------------

CREATE TABLE events (
	id bigserial PRIMARY KEY,
	type text NOT NULL,
	actor_user_id integer REFERENCES users(id) ON DELETE SET NULL,
    external_actor_username text,
    external_actor_url text,
    created_at timestamp with time zone NOT NULL,

	-- The various event types give their own meanings to these columns.
    data jsonb,
	thread_id bigint REFERENCES threads(id) ON DELETE CASCADE,
	campaign_id bigint REFERENCES campaigns(id) ON DELETE CASCADE,
    repository_id integer REFERENCES repo(id) ON DELETE CASCADE,
    user_id integer REFERENCES users(id) ON DELETE CASCADE,
    organization_id integer REFERENCES orgs(id) ON DELETE CASCADE,
    registry_extension_id integer REFERENCES registry_extensions(id) ON DELETE CASCADE,

    external_service_id bigint REFERENCES external_services(id) ON DELETE CASCADE
);
CREATE INDEX events_thread_id ON events(thread_id, created_at ASC) WHERE thread_id IS NOT NULL;
CREATE INDEX events_campaign_id ON events(campaign_id, created_at ASC) WHERE thread_id IS NOT NULL;
CREATE INDEX events_external_service_id ON events(external_service_id);

-----------------

ALTER TABLE threads ADD COLUMN is_draft boolean NOT NULL DEFAULT false;
ALTER TABLE threads ADD COLUMN is_pending_external_creation boolean NOT NULL DEFAULT false;
ALTER TABLE threads ADD COLUMN pending_patch text;

ALTER TABLE campaigns ADD COLUMN is_draft boolean NOT NULL DEFAULT false;
ALTER TABLE campaigns ADD COLUMN start_date timestamp with time zone;
ALTER TABLE campaigns ADD COLUMN due_date timestamp with time zone;
ALTER TABLE campaigns ADD COLUMN workflow_jsonc text;
ALTER TABLE campaigns ADD COLUMN extension_data jsonb;

CREATE TABLE rules (
	id bigserial PRIMARY KEY,
	container_campaign_id bigint REFERENCES campaigns(id) ON DELETE CASCADE,
	container_thread_id bigint REFERENCES threads(id) ON DELETE CASCADE,
	name text NOT NULL,
    template_id text,
    template_context text,
	description text,
	definition text NOT NULL,
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	updated_at timestamp with time zone NOT NULL DEFAULT now()
);
CREATE INDEX rules_container_campaign_id ON rules(container_campaign_id);
CREATE INDEX rules_container_thread_id ON rules(container_thread_id);

ALTER TABLE events ADD COLUMN rule_id bigint REFERENCES rules(id) ON DELETE CASCADE;

-----------------

CREATE TABLE thread_diagnostic_edges (
	id bigserial PRIMARY KEY,
	thread_id bigint NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
	--TODO!(sqs) location_repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
	type text NOT NULL,
	data jsonb NOT NULL
);
--TODO!(sqs) CREATE INDEX threads_diagnostics_location_repository_id ON threads_diagnostics(location_repository_id);
CREATE INDEX thread_diagnostic_edges_thread_id ON thread_diagnostic_edges(thread_id);

ALTER TABLE events ADD COLUMN thread_diagnostic_edge_id bigint REFERENCES thread_diagnostic_edges(id) ON DELETE SET NULL;

-----------------

-- See https://gitlab.com/gitlab-org/gitlab-ce/issues/19997#note_16081366 for why we don't have
-- group labels, only per-repository. labels. Instead of group labels, we will support batch
-- editing of labels across multiple repositories.
CREATE TABLE labels (
	id bigserial PRIMARY KEY,
    repository_id integer REFERENCES repo(id) ON DELETE CASCADE,
	name citext NOT NULL,
	description text,
	color text NOT NULL
);
CREATE INDEX labels_repository_id ON labels(repository_id);
CREATE INDEX labels_name ON labels(name);
CREATE UNIQUE INDEX labels_name_project_uniq ON labels(name, repository_id);

CREATE TABLE labels_objects (
	label_id bigint NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
	thread_id bigint REFERENCES threads(id) ON DELETE CASCADE
);
CREATE INDEX labels_objects_label_id ON labels_objects(label_id);
CREATE INDEX labels_objects_thread_id ON labels_objects(thread_id) WHERE thread_id IS NOT NULL;
CREATE UNIQUE INDEX labels_objects_uniq ON labels_objects(label_id, thread_id);

COMMIT;
