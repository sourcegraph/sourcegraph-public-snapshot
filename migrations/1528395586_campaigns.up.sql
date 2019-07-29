BEGIN;

CREATE TABLE campaigns (
	id bigserial PRIMARY KEY,
    namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
    settings text NOT NULL DEFAULT '{}'
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

COMMIT;
