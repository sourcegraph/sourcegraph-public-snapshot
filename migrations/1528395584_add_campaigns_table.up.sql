BEGIN;

CREATE TABLE campaigns (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
  description text,
  author_id integer NOT NULL REFERENCES users(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  namespace_user_id integer REFERENCES users(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  namespace_org_id integer REFERENCES orgs(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now()
);

ALTER TABLE campaigns ADD CONSTRAINT campaigns_has_1_namespace
CHECK ((namespace_user_id IS NULL) != (namespace_org_id IS NULL));

CREATE INDEX campaigns_namespace_user_id ON campaigns(namespace_user_id);
CREATE INDEX campaigns_namespace_org_id ON campaigns(namespace_org_id);

COMMIT;
