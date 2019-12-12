CREATE TABLE public.campaign_jobs (
    id bigint NOT NULL,
    campaign_plan_id bigint NOT NULL,
    repo_id bigint NOT NULL,
    rev text NOT NULL,
    diff text NOT NULL,
    error text NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    base_ref text NOT NULL,
    description text,
    CONSTRAINT campaign_jobs_base_ref_check CHECK ((base_ref <> ''::text))
);
CREATE TABLE public.campaign_plans (
    id bigint NOT NULL,
    campaign_type text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    arguments text NOT NULL,
    canceled_at timestamp with time zone,
    CONSTRAINT campaign_plans_campaign_type_check CHECK ((campaign_type <> ''::text))
);
CREATE TABLE public.campaigns (
    id bigint NOT NULL,
    name text NOT NULL,
    description text,
    author_id integer NOT NULL,
    namespace_user_id integer,
    namespace_org_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    changeset_ids jsonb DEFAULT '{}'::jsonb NOT NULL,
    campaign_plan_id integer,
    closed_at timestamp with time zone,
    CONSTRAINT campaigns_changeset_ids_check CHECK ((jsonb_typeof(changeset_ids) = 'object'::text)),
    CONSTRAINT campaigns_has_1_namespace CHECK (((namespace_user_id IS NULL) <> (namespace_org_id IS NULL)))
);
CREATE TABLE public.changeset_events (
    id bigint NOT NULL,
    changeset_id bigint NOT NULL,
    kind text NOT NULL,
    key text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT changeset_events_key_check CHECK ((key <> ''::text)),
    CONSTRAINT changeset_events_kind_check CHECK ((kind <> ''::text)),
    CONSTRAINT changeset_events_metadata_check CHECK ((jsonb_typeof(metadata) = 'object'::text))
);
CREATE TABLE public.changeset_jobs (
    id bigint NOT NULL,
    campaign_id bigint NOT NULL,
    campaign_job_id bigint NOT NULL,
    changeset_id bigint,
    error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone
);
CREATE TABLE public.changesets (
    id bigint NOT NULL,
    campaign_ids jsonb DEFAULT '{}'::jsonb NOT NULL,
    repo_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    external_id text NOT NULL,
    external_service_type text NOT NULL,
    CONSTRAINT changesets_campaign_ids_check CHECK ((jsonb_typeof(campaign_ids) = 'object'::text)),
    CONSTRAINT changesets_external_id_check CHECK ((external_id <> ''::text)),
    CONSTRAINT changesets_external_service_type_not_blank CHECK ((external_service_type <> ''::text)),
    CONSTRAINT changesets_metadata_check CHECK ((jsonb_typeof(metadata) = 'object'::text))
);
