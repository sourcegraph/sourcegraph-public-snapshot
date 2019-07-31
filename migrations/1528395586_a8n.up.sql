BEGIN;

CREATE TYPE thread_type AS enum ('THREAD', 'ISSUE', 'CHANGESET');
CREATE TABLE threads (
	id bigserial PRIMARY KEY,
	type thread_type NOT NULL,
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

CREATE TABLE comments (
    id bigserial PRIMARY KEY,
    thread_id bigint REFERENCES threads(id) ON DELETE CASCADE,
    author_user_id integer REFERENCES users(id) ON DELETE SET NULL,
    body text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);
CREATE INDEX comments_thread_id ON comments(thread_id);
CREATE INDEX comments_author_user_id ON comments(author_user_id);

-----------------

CREATE TABLE campaigns (
	id bigserial PRIMARY KEY,
    namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
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

CREATE TABLE rules (
	id bigserial PRIMARY KEY,
	project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
	settings text NOT NULL
);
CREATE INDEX rules_project_id ON rules(project_id);

COMMIT;
