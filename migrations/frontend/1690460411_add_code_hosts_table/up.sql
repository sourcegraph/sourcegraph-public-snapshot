CREATE TABLE IF NOT EXISTS code_hosts (
    id SERIAL PRIMARY KEY,
    kind text NOT NULL,
    url text NOT NULL UNIQUE,
    api_rate_limit_quota integer,
    api_rate_limit_interval_seconds integer,
    git_rate_limit_quota integer,
    git_rate_limit_interval_seconds integer,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

ALTER TABLE external_services
    ADD COLUMN IF NOT EXISTS code_host_id integer REFERENCES code_hosts(id) ON DELETE SET NULL ON UPDATE CASCADE DEFERRABLE INITIALLY DEFERRED;
