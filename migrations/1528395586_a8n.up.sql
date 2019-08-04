BEGIN;

CREATE TABLE threads (
	id bigserial PRIMARY KEY,
	type text NOT NULL,
	repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
	title text NOT NULL,
	external_url text,
	status text NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),

	-- type == CHANGESET
	is_preview boolean,
	base_ref text,
	head_ref text
);
CREATE INDEX threads_repository_id ON threads(repository_id);

-----------------

CREATE TABLE campaigns (
	id bigserial PRIMARY KEY,
    namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,
	name text NOT NULL,
    is_preview boolean NOT NULL DEFAULT false,
    rules text NOT NULL DEFAULT '[]',

    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);
ALTER TABLE campaigns ADD CONSTRAINT campaigns_has_1_namespace CHECK ((namespace_user_id IS NULL) != (namespace_org_id IS NULL));
CREATE INDEX campaigns_namespace_user_id ON campaigns(namespace_user_id);
CREATE INDEX campaigns_namespace_org_id ON campaigns(namespace_org_id);

CREATE TABLE campaigns_threads (
	campaign_id bigint NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
	thread_id bigint NOT NULL REFERENCES threads(id) ON DELETE CASCADE
);
CREATE INDEX campaigns_threads_campaign_id ON campaigns_threads(campaign_id);
CREATE INDEX campaigns_threads_thread_id ON campaigns_threads(thread_id) WHERE thread_id IS NOT NULL;
CREATE UNIQUE INDEX campaigns_threads_uniq ON campaigns_threads(campaign_id, thread_id);

-----------------

CREATE TABLE comments (
    id bigserial PRIMARY KEY,
    author_user_id integer REFERENCES users(id) ON DELETE SET NULL,
    body text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),

	parent_comment_id bigint REFERENCES comments(id) ON DELETE CASCADE,
    thread_id bigint REFERENCES threads(id) ON DELETE CASCADE,
    campaign_id bigint REFERENCES campaigns(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX comments_thread_id ON comments(thread_id);
CREATE UNIQUE INDEX comments_campaign_id ON comments(campaign_id);
CREATE INDEX comments_author_user_id ON comments(author_user_id);

-- Ensure every thread and campaign has a primary comment (the "top comment" whose body is the
-- description the object).
ALTER TABLE threads ADD COLUMN primary_comment_id bigint NOT NULL REFERENCES comments(id) ON DELETE RESTRICT;
ALTER TABLE campaigns ADD COLUMN primary_comment_id bigint NOT NULL REFERENCES comments(id) ON DELETE RESTRICT;

-----------------

CREATE TABLE rules (
	id bigserial PRIMARY KEY,
	project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
	settings text NOT NULL
);
CREATE INDEX rules_project_id ON rules(project_id);

-----------------

CREATE TABLE events (
	id bigserial PRIMARY KEY,
	type text NOT NULL,
	actor_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL,

	-- The various event types give their own meanings to these columns.
    data jsonb,
	thread_id bigint REFERENCES threads(id) ON DELETE CASCADE,
	campaign_id bigint REFERENCES campaigns(id) ON DELETE CASCADE,
	comment_id bigint REFERENCES comments(id) ON DELETE CASCADE,
	rule_id bigint REFERENCES rules(id) ON DELETE CASCADE,
    repository_id integer REFERENCES repo(id) ON DELETE CASCADE,
    user_id integer REFERENCES users(id) ON DELETE CASCADE,
    organization_id integer REFERENCES orgs(id) ON DELETE CASCADE,
    registry_extension_id integer REFERENCES registry_extensions(id) ON DELETE CASCADE,

    imported_from_external_service_id bigint REFERENCES external_services(id) ON DELETE CASCADE
);
CREATE INDEX events_thread_id ON events(thread_id, created_at ASC) WHERE thread_id IS NOT NULL;
CREATE INDEX events_campaign_id ON events(campaign_id, created_at ASC) WHERE thread_id IS NOT NULL;
CREATE INDEX events_imported_from_external_service_id ON events(imported_from_external_service_id);

COMMIT;
