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
    imported_from_external_service_id bigint REFERENCES external_services(id) ON DELETE CASCADE,
    external_id text,
    external_metadata jsonb
);
CREATE INDEX threads_repository_id ON threads(repository_id);
CREATE INDEX threads_imported_from_external_service_id ON threads(imported_from_external_service_id);
ALTER TABLE threads ADD CONSTRAINT external_thread_has_id_and_data CHECK ((imported_from_external_service_id IS NULL) = (external_id IS NULL) AND (external_id IS NULL) = (external_metadata IS NULL));
CREATE UNIQUE INDEX threads_external ON threads(imported_from_external_service_id, external_id) WHERE imported_from_external_service_id IS NOT NULL;

-----------------

CREATE TABLE campaigns (
	id bigserial PRIMARY KEY,
    namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,
	name text NOT NULL,

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
    author_external_actor_username text,
    author_external_actor_url text,
    body text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),

    campaign_id bigint REFERENCES campaigns(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX comments_campaign_id ON comments(campaign_id);
CREATE INDEX comments_author_user_id ON comments(author_user_id);

-- Ensure every campaign has a primary comment (the "top comment" whose body is the description of
-- the object).
ALTER TABLE campaigns ADD COLUMN primary_comment_id bigint NOT NULL REFERENCES comments(id) ON DELETE RESTRICT;

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
	comment_id bigint REFERENCES comments(id) ON DELETE CASCADE,
    repository_id integer REFERENCES repo(id) ON DELETE CASCADE,
    user_id integer REFERENCES users(id) ON DELETE CASCADE,
    organization_id integer REFERENCES orgs(id) ON DELETE CASCADE,
    registry_extension_id integer REFERENCES registry_extensions(id) ON DELETE CASCADE,

    imported_from_external_service_id bigint REFERENCES external_services(id) ON DELETE CASCADE
);
CREATE INDEX events_thread_id ON events(thread_id, created_at ASC) WHERE thread_id IS NOT NULL;
CREATE INDEX events_campaign_id ON events(campaign_id, created_at ASC) WHERE thread_id IS NOT NULL;
CREATE INDEX events_imported_from_external_service_id ON events(imported_from_external_service_id);

-----------------

CREATE TABLE commit_status_contexts (
	id bigserial PRIMARY KEY,
    repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
    commit_oid text NOT NULL,
    context text NOT NULL,
    state text NOT NULL,
    description text,
    target_url text,
	actor_user_id integer REFERENCES users(id) ON DELETE SET NULL,
    external_actor_username text,
    external_actor_url text,
    created_at timestamp with time zone NOT NULL DEFAULT now(),

	-- TODO!(sqs): make external_services use hard-delete not soft-delete so that the deletions
	-- cascade to the threads and events rows, which will prevent orphaned data - or else make their
	-- queries omit results if the corresponding external_services row was soft-deleted.
	imported_from_external_service_id bigint REFERENCES external_services(id) ON DELETE CASCADE,
	external_id text,
	external_metadata jsonb
);
CREATE INDEX commit_status_contexts_repository_commit_context ON commit_status_contexts(repository_id, commit_oid, context);

COMMIT;
