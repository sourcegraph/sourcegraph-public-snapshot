CREATE TYPE batch_changes_changeset_ui_publication_state AS ENUM (
    'UNPUBLISHED',
    'DRAFT',
    'PUBLISHED'
);

CREATE TYPE cm_email_priority AS ENUM (
    'NORMAL',
    'CRITICAL'
);

CREATE TYPE critical_or_site AS ENUM (
    'critical',
    'site'
);

CREATE TYPE feature_flag_type AS ENUM (
    'bool',
    'rollout'
);

CREATE TYPE lsif_index_state AS ENUM (
    'queued',
    'processing',
    'completed',
    'errored',
    'failed'
);

CREATE TYPE lsif_upload_state AS ENUM (
    'uploading',
    'queued',
    'processing',
    'completed',
    'errored',
    'deleted',
    'failed'
);

CREATE TYPE persistmode AS ENUM (
    'record',
    'snapshot'
);

CREATE FUNCTION delete_batch_change_reference_on_changesets() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        UPDATE
          changesets
        SET
          batch_change_ids = changesets.batch_change_ids - OLD.id::text
        WHERE
          changesets.batch_change_ids ? OLD.id::text;

        RETURN OLD;
    END;
$$;

CREATE FUNCTION delete_repo_ref_on_external_service_repos() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        -- if a repo is soft-deleted, delete every row that references that repo
        IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
        DELETE FROM
            external_service_repos
        WHERE
            repo_id = OLD.id;
        END IF;

        RETURN OLD;
    END;
$$;

CREATE FUNCTION invalidate_session_for_userid_on_password_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        IF OLD.passwd != NEW.passwd THEN
            NEW.invalidated_sessions_at = now() + (1 * interval '1 second');
            RETURN NEW;
        END IF;
    RETURN NEW;
    END;
$$;

CREATE FUNCTION repo_block(reason text, at timestamp with time zone) RETURNS jsonb
    LANGUAGE sql IMMUTABLE STRICT
    AS $$
SELECT jsonb_build_object(
    'reason', reason,
    'at', extract(epoch from timezone('utc', at))::bigint
);
$$;

CREATE PROCEDURE set_repo_stars_null_to_zero()
    LANGUAGE plpgsql
    AS $$
DECLARE
  done boolean;
  total integer = 0;
  updated integer = 0;

BEGIN
  SELECT COUNT(*) INTO total FROM repo WHERE stars IS NULL;

  RAISE NOTICE 'repo_stars_null_to_zero: updating % rows', total;

  done := total = 0;

  WHILE NOT done LOOP
    UPDATE repo SET stars = 0
    FROM (
      SELECT id FROM repo
      WHERE stars IS NULL
      LIMIT 10000
      FOR UPDATE SKIP LOCKED
    ) s
    WHERE repo.id = s.id;

    COMMIT;

    SELECT COUNT(*) = 0 INTO done FROM repo WHERE stars IS NULL LIMIT 1;

    updated := updated + 10000;

    RAISE NOTICE 'repo_stars_null_to_zero: updated % of % rows', updated, total;
  END LOOP;
END
$$;

CREATE FUNCTION soft_delete_orphan_repo_by_external_service_repos() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- When an external service is soft or hard-deleted,
    -- performs a clean up to soft-delete orphan repositories.
    UPDATE
        repo
    SET
        name = soft_deleted_repository_name(name),
        deleted_at = transaction_timestamp()
    WHERE
      deleted_at IS NULL
      AND NOT EXISTS (
        SELECT FROM external_service_repos WHERE repo_id = repo.id
      );
END;
$$;

CREATE FUNCTION soft_delete_user_reference_on_external_service() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- If a user is soft-deleted, delete every row that references that user
    IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
        UPDATE external_services
        SET deleted_at = NOW()
        WHERE namespace_user_id = OLD.id;
    END IF;

    RETURN OLD;
END;
$$;

CREATE FUNCTION soft_deleted_repository_name(name text) RETURNS text
    LANGUAGE plpgsql STRICT
    AS $$
BEGIN
    RETURN 'DELETED-' || extract(epoch from transaction_timestamp()) || '-' || name;
END;
$$;

CREATE FUNCTION versions_insert_row_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.first_version = NEW.version;
    RETURN NEW;
END $$;

CREATE TABLE access_tokens (
    id bigint NOT NULL,
    subject_user_id integer NOT NULL,
    value_sha256 bytea NOT NULL,
    note text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone,
    deleted_at timestamp with time zone,
    creator_user_id integer NOT NULL,
    scopes text[] NOT NULL,
    internal boolean DEFAULT false
);

CREATE SEQUENCE access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE access_tokens_id_seq OWNED BY access_tokens.id;

CREATE TABLE batch_changes (
    id bigint NOT NULL,
    name text NOT NULL,
    description text,
    creator_id integer,
    namespace_user_id integer,
    namespace_org_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    closed_at timestamp with time zone,
    batch_spec_id bigint NOT NULL,
    last_applier_id bigint,
    last_applied_at timestamp with time zone,
    CONSTRAINT batch_changes_has_1_namespace CHECK (((namespace_user_id IS NULL) <> (namespace_org_id IS NULL))),
    CONSTRAINT batch_changes_name_not_blank CHECK ((name <> ''::text))
);

CREATE SEQUENCE batch_changes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_changes_id_seq OWNED BY batch_changes.id;

CREATE TABLE batch_changes_site_credentials (
    id bigint NOT NULL,
    external_service_type text NOT NULL,
    external_service_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    credential bytea NOT NULL,
    encryption_key_id text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE batch_changes_site_credentials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_changes_site_credentials_id_seq OWNED BY batch_changes_site_credentials.id;

CREATE TABLE batch_spec_execution_cache_entries (
    id bigint NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    version integer NOT NULL,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id integer NOT NULL
);

CREATE SEQUENCE batch_spec_execution_cache_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_spec_execution_cache_entries_id_seq OWNED BY batch_spec_execution_cache_entries.id;

CREATE TABLE batch_spec_resolution_jobs (
    id bigint NOT NULL,
    batch_spec_id integer,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    queued_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE batch_spec_resolution_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_spec_resolution_jobs_id_seq OWNED BY batch_spec_resolution_jobs.id;

CREATE TABLE batch_spec_workspace_execution_jobs (
    id bigint NOT NULL,
    batch_spec_workspace_id integer,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    cancel boolean DEFAULT false NOT NULL,
    access_token_id bigint,
    queued_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE batch_spec_workspace_execution_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_spec_workspace_execution_jobs_id_seq OWNED BY batch_spec_workspace_execution_jobs.id;

CREATE TABLE batch_spec_workspaces (
    id bigint NOT NULL,
    batch_spec_id integer,
    changeset_spec_ids jsonb DEFAULT '{}'::jsonb,
    repo_id integer,
    branch text NOT NULL,
    commit text NOT NULL,
    path text NOT NULL,
    file_matches text[] NOT NULL,
    only_fetch_workspace boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    ignored boolean DEFAULT false NOT NULL,
    unsupported boolean DEFAULT false NOT NULL,
    skipped boolean DEFAULT false NOT NULL,
    cached_result_found boolean DEFAULT false NOT NULL,
    step_cache_results jsonb DEFAULT '{}'::jsonb NOT NULL
);

CREATE SEQUENCE batch_spec_workspaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_spec_workspaces_id_seq OWNED BY batch_spec_workspaces.id;

CREATE TABLE batch_specs (
    id bigint NOT NULL,
    rand_id text NOT NULL,
    raw_spec text NOT NULL,
    spec jsonb DEFAULT '{}'::jsonb NOT NULL,
    namespace_user_id integer,
    namespace_org_id integer,
    user_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_from_raw boolean DEFAULT false NOT NULL,
    allow_unsupported boolean DEFAULT false NOT NULL,
    allow_ignored boolean DEFAULT false NOT NULL,
    no_cache boolean DEFAULT false NOT NULL,
    CONSTRAINT batch_specs_has_1_namespace CHECK (((namespace_user_id IS NULL) <> (namespace_org_id IS NULL)))
);

CREATE SEQUENCE batch_specs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE batch_specs_id_seq OWNED BY batch_specs.id;

CREATE TABLE changeset_specs (
    id bigint NOT NULL,
    rand_id text NOT NULL,
    spec jsonb DEFAULT '{}'::jsonb NOT NULL,
    batch_spec_id bigint,
    repo_id integer NOT NULL,
    user_id integer,
    diff_stat_added integer,
    diff_stat_changed integer,
    diff_stat_deleted integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    head_ref text,
    title text,
    external_id text,
    fork_namespace citext
);

CREATE TABLE changesets (
    id bigint NOT NULL,
    batch_change_ids jsonb DEFAULT '{}'::jsonb NOT NULL,
    repo_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    external_id text,
    external_service_type text NOT NULL,
    external_deleted_at timestamp with time zone,
    external_branch text,
    external_updated_at timestamp with time zone,
    external_state text,
    external_review_state text,
    external_check_state text,
    diff_stat_added integer,
    diff_stat_changed integer,
    diff_stat_deleted integer,
    sync_state jsonb DEFAULT '{}'::jsonb NOT NULL,
    current_spec_id bigint,
    previous_spec_id bigint,
    publication_state text DEFAULT 'UNPUBLISHED'::text,
    owned_by_batch_change_id bigint,
    reconciler_state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    closing boolean DEFAULT false NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    log_contents text,
    execution_logs json[],
    syncer_error text,
    external_title text,
    worker_hostname text DEFAULT ''::text NOT NULL,
    ui_publication_state batch_changes_changeset_ui_publication_state,
    last_heartbeat_at timestamp with time zone,
    external_fork_namespace citext,
    queued_at timestamp with time zone DEFAULT now(),
    CONSTRAINT changesets_batch_change_ids_check CHECK ((jsonb_typeof(batch_change_ids) = 'object'::text)),
    CONSTRAINT changesets_external_id_check CHECK ((external_id <> ''::text)),
    CONSTRAINT changesets_external_service_type_not_blank CHECK ((external_service_type <> ''::text)),
    CONSTRAINT changesets_metadata_check CHECK ((jsonb_typeof(metadata) = 'object'::text)),
    CONSTRAINT external_branch_ref_prefix CHECK ((external_branch ~~ 'refs/heads/%'::text))
);

COMMENT ON COLUMN changesets.external_title IS 'Normalized property generated on save using Changeset.Title()';

CREATE TABLE repo (
    id integer NOT NULL,
    name citext NOT NULL,
    description text,
    fork boolean,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    external_id text,
    external_service_type text,
    external_service_id text,
    archived boolean DEFAULT false NOT NULL,
    uri citext,
    deleted_at timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    private boolean DEFAULT false NOT NULL,
    stars integer DEFAULT 0 NOT NULL,
    blocked jsonb,
    CONSTRAINT check_name_nonempty CHECK ((name OPERATOR(<>) ''::citext)),
    CONSTRAINT repo_metadata_check CHECK ((jsonb_typeof(metadata) = 'object'::text))
);

CREATE VIEW branch_changeset_specs_and_changesets AS
 SELECT changeset_specs.id AS changeset_spec_id,
    COALESCE(changesets.id, (0)::bigint) AS changeset_id,
    changeset_specs.repo_id,
    changeset_specs.batch_spec_id,
    changesets.owned_by_batch_change_id AS owner_batch_change_id,
    repo.name AS repo_name,
    changeset_specs.title AS changeset_name,
    changesets.external_state,
    changesets.publication_state,
    changesets.reconciler_state
   FROM ((changeset_specs
     LEFT JOIN changesets ON (((changesets.repo_id = changeset_specs.repo_id) AND (changesets.current_spec_id IS NOT NULL) AND (EXISTS ( SELECT 1
           FROM changeset_specs changeset_specs_1
          WHERE ((changeset_specs_1.id = changesets.current_spec_id) AND (changeset_specs_1.head_ref = changeset_specs.head_ref)))))))
     JOIN repo ON ((changeset_specs.repo_id = repo.id)))
  WHERE ((changeset_specs.external_id IS NULL) AND (repo.deleted_at IS NULL));

CREATE TABLE changeset_events (
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

CREATE SEQUENCE changeset_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE changeset_events_id_seq OWNED BY changeset_events.id;

CREATE TABLE changeset_jobs (
    id bigint NOT NULL,
    bulk_group text NOT NULL,
    user_id integer NOT NULL,
    batch_change_id integer NOT NULL,
    changeset_id integer NOT NULL,
    job_type text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    queued_at timestamp with time zone DEFAULT now(),
    CONSTRAINT changeset_jobs_payload_check CHECK ((jsonb_typeof(payload) = 'object'::text))
);

CREATE SEQUENCE changeset_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE changeset_jobs_id_seq OWNED BY changeset_jobs.id;

CREATE SEQUENCE changeset_specs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE changeset_specs_id_seq OWNED BY changeset_specs.id;

CREATE SEQUENCE changesets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE changesets_id_seq OWNED BY changesets.id;

CREATE TABLE cm_action_jobs (
    id integer NOT NULL,
    email bigint,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    log_contents text,
    trigger_event integer,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    execution_logs json[],
    webhook bigint,
    slack_webhook bigint,
    queued_at timestamp with time zone DEFAULT now(),
    CONSTRAINT cm_action_jobs_only_one_action_type CHECK ((((
CASE
    WHEN (email IS NULL) THEN 0
    ELSE 1
END +
CASE
    WHEN (webhook IS NULL) THEN 0
    ELSE 1
END) +
CASE
    WHEN (slack_webhook IS NULL) THEN 0
    ELSE 1
END) = 1))
);

COMMENT ON COLUMN cm_action_jobs.email IS 'The ID of the cm_emails action to execute if this is an email job. Mutually exclusive with webhook and slack_webhook';

COMMENT ON COLUMN cm_action_jobs.webhook IS 'The ID of the cm_webhooks action to execute if this is a webhook job. Mutually exclusive with email and slack_webhook';

COMMENT ON COLUMN cm_action_jobs.slack_webhook IS 'The ID of the cm_slack_webhook action to execute if this is a slack webhook job. Mutually exclusive with email and webhook';

COMMENT ON CONSTRAINT cm_action_jobs_only_one_action_type ON cm_action_jobs IS 'Constrains that each queued code monitor action has exactly one action type';

CREATE SEQUENCE cm_action_jobs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_action_jobs_id_seq OWNED BY cm_action_jobs.id;

CREATE TABLE cm_emails (
    id bigint NOT NULL,
    monitor bigint NOT NULL,
    enabled boolean NOT NULL,
    priority cm_email_priority NOT NULL,
    header text NOT NULL,
    created_by integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by integer NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    include_results boolean DEFAULT false NOT NULL
);

CREATE SEQUENCE cm_emails_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_emails_id_seq OWNED BY cm_emails.id;

CREATE TABLE cm_last_searched (
    monitor_id bigint NOT NULL,
    args_hash bigint NOT NULL,
    commit_oids text[] NOT NULL
);

COMMENT ON TABLE cm_last_searched IS 'The last searched commit hashes for the given code monitor and unique set of search arguments';

COMMENT ON COLUMN cm_last_searched.args_hash IS 'A unique hash of the gitserver search arguments to identify this search job';

COMMENT ON COLUMN cm_last_searched.commit_oids IS 'The set of commit OIDs that was previously successfully searched and should be excluded on the next run';

CREATE TABLE cm_monitors (
    id bigint NOT NULL,
    created_by integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    description text NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by integer NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    namespace_user_id integer NOT NULL,
    namespace_org_id integer
);

COMMENT ON COLUMN cm_monitors.namespace_org_id IS 'DEPRECATED: code monitors cannot be owned by an org';

CREATE SEQUENCE cm_monitors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_monitors_id_seq OWNED BY cm_monitors.id;

CREATE TABLE cm_queries (
    id bigint NOT NULL,
    monitor bigint NOT NULL,
    query text NOT NULL,
    created_by integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by integer NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    next_run timestamp with time zone DEFAULT now(),
    latest_result timestamp with time zone
);

CREATE SEQUENCE cm_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_queries_id_seq OWNED BY cm_queries.id;

CREATE TABLE cm_recipients (
    id bigint NOT NULL,
    email bigint NOT NULL,
    namespace_user_id integer,
    namespace_org_id integer
);

CREATE SEQUENCE cm_recipients_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_recipients_id_seq OWNED BY cm_recipients.id;

CREATE TABLE cm_slack_webhooks (
    id bigint NOT NULL,
    monitor bigint NOT NULL,
    url text NOT NULL,
    enabled boolean NOT NULL,
    created_by integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by integer NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    include_results boolean DEFAULT false NOT NULL
);

COMMENT ON TABLE cm_slack_webhooks IS 'Slack webhook actions configured on code monitors';

COMMENT ON COLUMN cm_slack_webhooks.monitor IS 'The code monitor that the action is defined on';

COMMENT ON COLUMN cm_slack_webhooks.url IS 'The Slack webhook URL we send the code monitor event to';

CREATE SEQUENCE cm_slack_webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_slack_webhooks_id_seq OWNED BY cm_slack_webhooks.id;

CREATE TABLE cm_trigger_jobs (
    id integer NOT NULL,
    query bigint NOT NULL,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    log_contents text,
    query_string text,
    results boolean,
    num_results integer,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    execution_logs json[],
    search_results jsonb,
    queued_at timestamp with time zone DEFAULT now(),
    CONSTRAINT search_results_is_array CHECK ((jsonb_typeof(search_results) = 'array'::text))
);

COMMENT ON COLUMN cm_trigger_jobs.results IS 'DEPRECATED: replaced by len(search_results) > 0. Can be removed after version 3.37 release cut';

COMMENT ON COLUMN cm_trigger_jobs.num_results IS 'DEPRECATED: replaced by len(search_results). Can be removed after version 3.37 release cut';

CREATE SEQUENCE cm_trigger_jobs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_trigger_jobs_id_seq OWNED BY cm_trigger_jobs.id;

CREATE TABLE cm_webhooks (
    id bigint NOT NULL,
    monitor bigint NOT NULL,
    url text NOT NULL,
    enabled boolean NOT NULL,
    created_by integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by integer NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    include_results boolean DEFAULT false NOT NULL
);

COMMENT ON TABLE cm_webhooks IS 'Webhook actions configured on code monitors';

COMMENT ON COLUMN cm_webhooks.monitor IS 'The code monitor that the action is defined on';

COMMENT ON COLUMN cm_webhooks.url IS 'The webhook URL we send the code monitor event to';

COMMENT ON COLUMN cm_webhooks.enabled IS 'Whether this Slack webhook action is enabled. When not enabled, the action will not be run when its code monitor generates events';

CREATE SEQUENCE cm_webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE cm_webhooks_id_seq OWNED BY cm_webhooks.id;

CREATE TABLE critical_and_site_config (
    id integer NOT NULL,
    type critical_or_site NOT NULL,
    contents text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE critical_and_site_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE critical_and_site_config_id_seq OWNED BY critical_and_site_config.id;

CREATE TABLE discussion_comments (
    id bigint NOT NULL,
    thread_id bigint NOT NULL,
    author_user_id integer NOT NULL,
    contents text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    reports text[] DEFAULT '{}'::text[] NOT NULL
);

CREATE SEQUENCE discussion_comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE discussion_comments_id_seq OWNED BY discussion_comments.id;

CREATE TABLE discussion_mail_reply_tokens (
    token text NOT NULL,
    user_id integer NOT NULL,
    thread_id bigint NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE discussion_threads (
    id bigint NOT NULL,
    author_user_id integer NOT NULL,
    title text,
    target_repo_id bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE SEQUENCE discussion_threads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE discussion_threads_id_seq OWNED BY discussion_threads.id;

CREATE TABLE discussion_threads_target_repo (
    id bigint NOT NULL,
    thread_id bigint NOT NULL,
    repo_id integer NOT NULL,
    path text,
    branch text,
    revision text,
    start_line integer,
    end_line integer,
    start_character integer,
    end_character integer,
    lines_before text,
    lines text,
    lines_after text
);

CREATE SEQUENCE discussion_threads_target_repo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE discussion_threads_target_repo_id_seq OWNED BY discussion_threads_target_repo.id;

CREATE TABLE event_logs (
    id bigint NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    user_id integer NOT NULL,
    anonymous_user_id text NOT NULL,
    source text NOT NULL,
    argument jsonb NOT NULL,
    version text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    feature_flags jsonb,
    cohort_id date,
    public_argument jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT event_logs_check_has_user CHECK ((((user_id = 0) AND (anonymous_user_id <> ''::text)) OR ((user_id <> 0) AND (anonymous_user_id = ''::text)) OR ((user_id <> 0) AND (anonymous_user_id <> ''::text)))),
    CONSTRAINT event_logs_check_name_not_empty CHECK ((name <> ''::text)),
    CONSTRAINT event_logs_check_source_not_empty CHECK ((source <> ''::text)),
    CONSTRAINT event_logs_check_version_not_empty CHECK ((version <> ''::text))
);

CREATE SEQUENCE event_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE event_logs_id_seq OWNED BY event_logs.id;

CREATE TABLE executor_heartbeats (
    id integer NOT NULL,
    hostname text NOT NULL,
    queue_name text NOT NULL,
    os text NOT NULL,
    architecture text NOT NULL,
    docker_version text NOT NULL,
    executor_version text NOT NULL,
    git_version text NOT NULL,
    ignite_version text NOT NULL,
    src_cli_version text NOT NULL,
    first_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE executor_heartbeats IS 'Tracks the most recent activity of executors attached to this Sourcegraph instance.';

COMMENT ON COLUMN executor_heartbeats.hostname IS 'The uniquely identifying name of the executor.';

COMMENT ON COLUMN executor_heartbeats.queue_name IS 'The queue name that the executor polls for work.';

COMMENT ON COLUMN executor_heartbeats.os IS 'The operating system running the executor.';

COMMENT ON COLUMN executor_heartbeats.architecture IS 'The machine architure running the executor.';

COMMENT ON COLUMN executor_heartbeats.docker_version IS 'The version of Docker used by the executor.';

COMMENT ON COLUMN executor_heartbeats.executor_version IS 'The version of the executor.';

COMMENT ON COLUMN executor_heartbeats.git_version IS 'The version of Git used by the executor.';

COMMENT ON COLUMN executor_heartbeats.ignite_version IS 'The version of Ignite used by the executor.';

COMMENT ON COLUMN executor_heartbeats.src_cli_version IS 'The version of src-cli used by the executor.';

COMMENT ON COLUMN executor_heartbeats.first_seen_at IS 'The first time a heartbeat from the executor was received.';

COMMENT ON COLUMN executor_heartbeats.last_seen_at IS 'The last time a heartbeat from the executor was received.';

CREATE SEQUENCE executor_heartbeats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE executor_heartbeats_id_seq OWNED BY executor_heartbeats.id;

CREATE TABLE external_service_repos (
    external_service_id bigint NOT NULL,
    repo_id integer NOT NULL,
    clone_url text NOT NULL,
    user_id integer,
    org_id integer,
    created_at timestamp with time zone DEFAULT transaction_timestamp() NOT NULL
);

CREATE SEQUENCE external_service_sync_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE external_service_sync_jobs (
    id integer DEFAULT nextval('external_service_sync_jobs_id_seq'::regclass) NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    external_service_id bigint,
    num_failures integer DEFAULT 0 NOT NULL,
    log_contents text,
    execution_logs json[],
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    queued_at timestamp with time zone DEFAULT now()
);

CREATE TABLE external_services (
    id bigint NOT NULL,
    kind text NOT NULL,
    display_name text NOT NULL,
    config text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    last_sync_at timestamp with time zone,
    next_sync_at timestamp with time zone,
    namespace_user_id integer,
    unrestricted boolean DEFAULT false NOT NULL,
    cloud_default boolean DEFAULT false NOT NULL,
    encryption_key_id text DEFAULT ''::text NOT NULL,
    namespace_org_id integer,
    has_webhooks boolean,
    token_expires_at timestamp with time zone,
    CONSTRAINT check_non_empty_config CHECK ((btrim(config) <> ''::text)),
    CONSTRAINT external_services_max_1_namespace CHECK ((((namespace_user_id IS NULL) AND (namespace_org_id IS NULL)) OR ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL))))
);

CREATE VIEW external_service_sync_jobs_with_next_sync_at AS
 SELECT j.id,
    j.state,
    j.failure_message,
    j.queued_at,
    j.started_at,
    j.finished_at,
    j.process_after,
    j.num_resets,
    j.num_failures,
    j.execution_logs,
    j.external_service_id,
    e.next_sync_at
   FROM (external_services e
     JOIN external_service_sync_jobs j ON ((e.id = j.external_service_id)));

CREATE SEQUENCE external_services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE external_services_id_seq OWNED BY external_services.id;

CREATE TABLE feature_flag_overrides (
    namespace_org_id integer,
    namespace_user_id integer,
    flag_name text NOT NULL,
    flag_value boolean NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT feature_flag_overrides_has_org_or_user_id CHECK (((namespace_org_id IS NOT NULL) OR (namespace_user_id IS NOT NULL)))
);

CREATE TABLE feature_flags (
    flag_name text NOT NULL,
    flag_type feature_flag_type NOT NULL,
    bool_value boolean,
    rollout integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT feature_flags_rollout_check CHECK (((rollout >= 0) AND (rollout <= 10000))),
    CONSTRAINT required_bool_fields CHECK ((1 =
CASE
    WHEN ((flag_type = 'bool'::feature_flag_type) AND (bool_value IS NULL)) THEN 0
    WHEN ((flag_type <> 'bool'::feature_flag_type) AND (bool_value IS NOT NULL)) THEN 0
    ELSE 1
END)),
    CONSTRAINT required_rollout_fields CHECK ((1 =
CASE
    WHEN ((flag_type = 'rollout'::feature_flag_type) AND (rollout IS NULL)) THEN 0
    WHEN ((flag_type <> 'rollout'::feature_flag_type) AND (rollout IS NOT NULL)) THEN 0
    ELSE 1
END))
);

COMMENT ON COLUMN feature_flags.bool_value IS 'Bool value only defined when flag_type is bool';

COMMENT ON COLUMN feature_flags.rollout IS 'Rollout only defined when flag_type is rollout. Increments of 0.01%';

COMMENT ON CONSTRAINT required_bool_fields ON feature_flags IS 'Checks that bool_value is set IFF flag_type = bool';

COMMENT ON CONSTRAINT required_rollout_fields ON feature_flags IS 'Checks that rollout is set IFF flag_type = rollout';

CREATE TABLE gitserver_localclone_jobs (
    id integer NOT NULL,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    last_heartbeat_at timestamp with time zone,
    execution_logs json[],
    worker_hostname text DEFAULT ''::text NOT NULL,
    repo_id integer NOT NULL,
    source_hostname text NOT NULL,
    dest_hostname text NOT NULL,
    delete_source boolean DEFAULT false NOT NULL,
    queued_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE gitserver_localclone_jobs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE gitserver_localclone_jobs_id_seq OWNED BY gitserver_localclone_jobs.id;

CREATE VIEW gitserver_localclone_jobs_with_repo_name AS
 SELECT glj.id,
    glj.state,
    glj.failure_message,
    glj.started_at,
    glj.finished_at,
    glj.process_after,
    glj.num_resets,
    glj.num_failures,
    glj.last_heartbeat_at,
    glj.execution_logs,
    glj.worker_hostname,
    glj.repo_id,
    glj.source_hostname,
    glj.dest_hostname,
    glj.delete_source,
    glj.queued_at,
    r.name AS repo_name
   FROM (gitserver_localclone_jobs glj
     JOIN repo r ON ((r.id = glj.repo_id)));

CREATE TABLE gitserver_repos (
    repo_id integer NOT NULL,
    clone_status text DEFAULT 'not_cloned'::text NOT NULL,
    shard_id text NOT NULL,
    last_error text,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_fetched timestamp with time zone DEFAULT now() NOT NULL,
    last_changed timestamp with time zone DEFAULT now() NOT NULL,
    repo_size_bytes bigint
);

CREATE TABLE global_state (
    site_id uuid NOT NULL,
    initialized boolean DEFAULT false NOT NULL
);

CREATE TABLE insights_query_runner_jobs (
    id integer NOT NULL,
    series_id text NOT NULL,
    search_query text NOT NULL,
    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    record_time timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    priority integer DEFAULT 1 NOT NULL,
    cost integer DEFAULT 500 NOT NULL,
    persist_mode persistmode DEFAULT 'record'::persistmode NOT NULL,
    queued_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE insights_query_runner_jobs IS 'See [internal/insights/background/queryrunner/worker.go:Job](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:internal/insights/background/queryrunner/worker.go+type+Job&patternType=literal)';

COMMENT ON COLUMN insights_query_runner_jobs.priority IS 'Integer representing a category of priority for this query. Priority in this context is ambiguously defined for consumers to decide an interpretation.';

COMMENT ON COLUMN insights_query_runner_jobs.cost IS 'Integer representing a cost approximation of executing this search query.';

COMMENT ON COLUMN insights_query_runner_jobs.persist_mode IS 'The persistence level for this query. This value will determine the lifecycle of the resulting value.';

CREATE TABLE insights_query_runner_jobs_dependencies (
    id integer NOT NULL,
    job_id integer NOT NULL,
    recording_time timestamp without time zone NOT NULL
);

COMMENT ON TABLE insights_query_runner_jobs_dependencies IS 'Stores data points for a code insight that do not need to be queried directly, but depend on the result of a query at a different point';

COMMENT ON COLUMN insights_query_runner_jobs_dependencies.job_id IS 'Foreign key to the job that owns this record.';

COMMENT ON COLUMN insights_query_runner_jobs_dependencies.recording_time IS 'The time for which this dependency should be recorded at using the parents value.';

CREATE SEQUENCE insights_query_runner_jobs_dependencies_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insights_query_runner_jobs_dependencies_id_seq OWNED BY insights_query_runner_jobs_dependencies.id;

CREATE SEQUENCE insights_query_runner_jobs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insights_query_runner_jobs_id_seq OWNED BY insights_query_runner_jobs.id;

CREATE TABLE insights_settings_migration_jobs (
    id integer NOT NULL,
    user_id integer,
    org_id integer,
    global boolean,
    settings_id integer NOT NULL,
    total_insights integer DEFAULT 0 NOT NULL,
    migrated_insights integer DEFAULT 0 NOT NULL,
    total_dashboards integer DEFAULT 0 NOT NULL,
    migrated_dashboards integer DEFAULT 0 NOT NULL,
    runs integer DEFAULT 0 NOT NULL,
    completed_at timestamp without time zone
);

CREATE SEQUENCE insights_settings_migration_jobs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insights_settings_migration_jobs_id_seq OWNED BY insights_settings_migration_jobs.id;

CREATE TABLE lsif_configuration_policies (
    id integer NOT NULL,
    repository_id integer,
    name text,
    type text NOT NULL,
    pattern text NOT NULL,
    retention_enabled boolean NOT NULL,
    retention_duration_hours integer,
    retain_intermediate_commits boolean NOT NULL,
    indexing_enabled boolean NOT NULL,
    index_commit_max_age_hours integer,
    index_intermediate_commits boolean NOT NULL,
    protected boolean DEFAULT false NOT NULL,
    repository_patterns text[],
    last_resolved_at timestamp with time zone
);

COMMENT ON COLUMN lsif_configuration_policies.repository_id IS 'The identifier of the repository to which this configuration policy applies. If absent, this policy is applied globally.';

COMMENT ON COLUMN lsif_configuration_policies.type IS 'The type of Git object (e.g., COMMIT, BRANCH, TAG).';

COMMENT ON COLUMN lsif_configuration_policies.pattern IS 'A pattern used to match` names of the associated Git object type.';

COMMENT ON COLUMN lsif_configuration_policies.retention_enabled IS 'Whether or not this configuration policy affects data retention rules.';

COMMENT ON COLUMN lsif_configuration_policies.retention_duration_hours IS 'The max age of data retained by this configuration policy. If null, the age is unbounded.';

COMMENT ON COLUMN lsif_configuration_policies.retain_intermediate_commits IS 'If the matching Git object is a branch, setting this value to true will also retain all data used to resolve queries for any commit on the matching branches. Setting this value to false will only consider the tip of the branch.';

COMMENT ON COLUMN lsif_configuration_policies.indexing_enabled IS 'Whether or not this configuration policy affects auto-indexing schedules.';

COMMENT ON COLUMN lsif_configuration_policies.index_commit_max_age_hours IS 'The max age of commits indexed by this configuration policy. If null, the age is unbounded.';

COMMENT ON COLUMN lsif_configuration_policies.index_intermediate_commits IS 'If the matching Git object is a branch, setting this value to true will also index all commits on the matching branches. Setting this value to false will only consider the tip of the branch.';

COMMENT ON COLUMN lsif_configuration_policies.protected IS 'Whether or not this configuration policy is protected from modification of its data retention behavior (except for duration).';

COMMENT ON COLUMN lsif_configuration_policies.repository_patterns IS 'The name pattern matching repositories to which this configuration policy applies. If absent, all repositories are matched.';

CREATE SEQUENCE lsif_configuration_policies_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_configuration_policies_id_seq OWNED BY lsif_configuration_policies.id;

CREATE TABLE lsif_configuration_policies_repository_pattern_lookup (
    policy_id integer NOT NULL,
    repo_id integer NOT NULL
);

COMMENT ON TABLE lsif_configuration_policies_repository_pattern_lookup IS 'A lookup table to get all the repository patterns by repository id that apply to a configuration policy.';

COMMENT ON COLUMN lsif_configuration_policies_repository_pattern_lookup.policy_id IS 'The policy identifier associated with the repository.';

COMMENT ON COLUMN lsif_configuration_policies_repository_pattern_lookup.repo_id IS 'The repository identifier associated with the policy.';

CREATE TABLE lsif_dependency_indexing_jobs (
    id integer NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    queued_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    last_heartbeat_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL,
    upload_id integer,
    external_service_kind text DEFAULT ''::text NOT NULL,
    external_service_sync timestamp with time zone
);

COMMENT ON COLUMN lsif_dependency_indexing_jobs.external_service_kind IS 'Filter the external services for this kind to wait to have synced. If empty, external_service_sync is ignored and no external services are polled for their last sync time.';

COMMENT ON COLUMN lsif_dependency_indexing_jobs.external_service_sync IS 'The sync time after which external services of the given kind will have synced/created any repositories referenced by the LSIF upload that are resolvable.';

CREATE TABLE lsif_dependency_syncing_jobs (
    id integer NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    queued_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    upload_id integer,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone
);

COMMENT ON TABLE lsif_dependency_syncing_jobs IS 'Tracks jobs that scan imports of indexes to schedule auto-index jobs.';

COMMENT ON COLUMN lsif_dependency_syncing_jobs.upload_id IS 'The identifier of the triggering upload record.';

CREATE SEQUENCE lsif_dependency_indexing_jobs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_dependency_indexing_jobs_id_seq OWNED BY lsif_dependency_syncing_jobs.id;

CREATE SEQUENCE lsif_dependency_indexing_jobs_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_dependency_indexing_jobs_id_seq1 OWNED BY lsif_dependency_indexing_jobs.id;

CREATE TABLE lsif_dependency_repos (
    id bigint NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    scheme text NOT NULL
);

CREATE SEQUENCE lsif_dependency_repos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_dependency_repos_id_seq OWNED BY lsif_dependency_repos.id;

CREATE TABLE lsif_dirty_repositories (
    repository_id integer NOT NULL,
    dirty_token integer NOT NULL,
    update_token integer NOT NULL,
    updated_at timestamp with time zone
);

COMMENT ON TABLE lsif_dirty_repositories IS 'Stores whether or not the nearest upload data for a repository is out of date (when update_token > dirty_token).';

COMMENT ON COLUMN lsif_dirty_repositories.dirty_token IS 'Set to the value of update_token visible to the transaction that updates the commit graph. Updates of dirty_token during this time will cause a second update.';

COMMENT ON COLUMN lsif_dirty_repositories.update_token IS 'This value is incremented on each request to update the commit graph for the repository.';

COMMENT ON COLUMN lsif_dirty_repositories.updated_at IS 'The time the update_token value was last updated.';

CREATE TABLE lsif_uploads (
    id integer NOT NULL,
    commit text NOT NULL,
    root text DEFAULT ''::text NOT NULL,
    uploaded_at timestamp with time zone DEFAULT now() NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    repository_id integer NOT NULL,
    indexer text NOT NULL,
    num_parts integer NOT NULL,
    uploaded_parts integer[] NOT NULL,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    upload_size bigint,
    num_failures integer DEFAULT 0 NOT NULL,
    associated_index_id bigint,
    committed_at timestamp with time zone,
    commit_last_checked_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    execution_logs json[],
    num_references integer,
    expired boolean DEFAULT false NOT NULL,
    last_retention_scan_at timestamp with time zone,
    reference_count integer,
    indexer_version text,
    queued_at timestamp with time zone,
    CONSTRAINT lsif_uploads_commit_valid_chars CHECK ((commit ~ '^[a-z0-9]{40}$'::text))
);

COMMENT ON TABLE lsif_uploads IS 'Stores metadata about an LSIF index uploaded by a user.';

COMMENT ON COLUMN lsif_uploads.id IS 'Used as a logical foreign key with the (disjoint) codeintel database.';

COMMENT ON COLUMN lsif_uploads.commit IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN lsif_uploads.root IS 'The path for which the index can resolve code intelligence relative to the repository root.';

COMMENT ON COLUMN lsif_uploads.indexer IS 'The name of the indexer that produced the index file. If not supplied by the user it will be pulled from the index metadata.';

COMMENT ON COLUMN lsif_uploads.num_parts IS 'The number of parts src-cli split the upload file into.';

COMMENT ON COLUMN lsif_uploads.uploaded_parts IS 'The index of parts that have been successfully uploaded.';

COMMENT ON COLUMN lsif_uploads.upload_size IS 'The size of the index file (in bytes).';

COMMENT ON COLUMN lsif_uploads.num_references IS 'Deprecated in favor of reference_count.';

COMMENT ON COLUMN lsif_uploads.expired IS 'Whether or not this upload data is no longer protected by any data retention policy.';

COMMENT ON COLUMN lsif_uploads.last_retention_scan_at IS 'The last time this upload was checked against data retention policies.';

COMMENT ON COLUMN lsif_uploads.reference_count IS 'The number of references to this upload data from other upload records (via lsif_references).';

COMMENT ON COLUMN lsif_uploads.indexer_version IS 'The version of the indexer that produced the index file. If not supplied by the user it will be pulled from the index metadata.';

CREATE VIEW lsif_dumps AS
 SELECT u.id,
    u.commit,
    u.root,
    u.queued_at,
    u.uploaded_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.indexer,
    u.indexer_version,
    u.num_parts,
    u.uploaded_parts,
    u.process_after,
    u.num_resets,
    u.upload_size,
    u.num_failures,
    u.associated_index_id,
    u.expired,
    u.last_retention_scan_at,
    u.finished_at AS processed_at
   FROM lsif_uploads u
  WHERE ((u.state = 'completed'::text) OR (u.state = 'deleting'::text));

CREATE SEQUENCE lsif_dumps_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_dumps_id_seq OWNED BY lsif_uploads.id;

CREATE VIEW lsif_dumps_with_repository_name AS
 SELECT u.id,
    u.commit,
    u.root,
    u.queued_at,
    u.uploaded_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.indexer,
    u.indexer_version,
    u.num_parts,
    u.uploaded_parts,
    u.process_after,
    u.num_resets,
    u.upload_size,
    u.num_failures,
    u.associated_index_id,
    u.expired,
    u.last_retention_scan_at,
    u.processed_at,
    r.name AS repository_name
   FROM (lsif_dumps u
     JOIN repo r ON ((r.id = u.repository_id)))
  WHERE (r.deleted_at IS NULL);

CREATE TABLE lsif_index_configuration (
    id bigint NOT NULL,
    repository_id integer NOT NULL,
    data bytea NOT NULL,
    autoindex_enabled boolean DEFAULT true NOT NULL
);

COMMENT ON TABLE lsif_index_configuration IS 'Stores the configuration used for code intel index jobs for a repository.';

COMMENT ON COLUMN lsif_index_configuration.data IS 'The raw user-supplied [configuration](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/autoindex/config/types.go#L3:6) (encoded in JSONC).';

COMMENT ON COLUMN lsif_index_configuration.autoindex_enabled IS 'Whether or not auto-indexing should be attempted on this repo. Index jobs may be inferred from the repository contents if data is empty.';

CREATE SEQUENCE lsif_index_configuration_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_index_configuration_id_seq OWNED BY lsif_index_configuration.id;

CREATE TABLE lsif_indexes (
    id bigint NOT NULL,
    commit text NOT NULL,
    queued_at timestamp with time zone DEFAULT now() NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    repository_id integer NOT NULL,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    docker_steps jsonb[] NOT NULL,
    root text NOT NULL,
    indexer text NOT NULL,
    indexer_args text[] NOT NULL,
    outfile text NOT NULL,
    log_contents text,
    execution_logs json[],
    local_steps text[] NOT NULL,
    commit_last_checked_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    CONSTRAINT lsif_uploads_commit_valid_chars CHECK ((commit ~ '^[a-z0-9]{40}$'::text))
);

COMMENT ON TABLE lsif_indexes IS 'Stores metadata about a code intel index job.';

COMMENT ON COLUMN lsif_indexes.commit IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN lsif_indexes.docker_steps IS 'An array of pre-index [steps](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/stores/dbstore/docker_step.go#L9:6) to run.';

COMMENT ON COLUMN lsif_indexes.root IS 'The working directory of the indexer image relative to the repository root.';

COMMENT ON COLUMN lsif_indexes.indexer IS 'The docker image used to run the index command (e.g. sourcegraph/lsif-go).';

COMMENT ON COLUMN lsif_indexes.indexer_args IS 'The command run inside the indexer image to produce the index file (e.g. [''lsif-node'', ''-p'', ''.''])';

COMMENT ON COLUMN lsif_indexes.outfile IS 'The path to the index file produced by the index command relative to the working directory.';

COMMENT ON COLUMN lsif_indexes.log_contents IS '**Column deprecated in favor of execution_logs.**';

COMMENT ON COLUMN lsif_indexes.execution_logs IS 'An array of [log entries](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/internal/workerutil/store.go#L48:6) (encoded as JSON) from the most recent execution.';

COMMENT ON COLUMN lsif_indexes.local_steps IS 'A list of commands to run inside the indexer image prior to running the indexer command.';

CREATE SEQUENCE lsif_indexes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_indexes_id_seq OWNED BY lsif_indexes.id;

CREATE VIEW lsif_indexes_with_repository_name AS
 SELECT u.id,
    u.commit,
    u.queued_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.process_after,
    u.num_resets,
    u.num_failures,
    u.docker_steps,
    u.root,
    u.indexer,
    u.indexer_args,
    u.outfile,
    u.log_contents,
    u.execution_logs,
    u.local_steps,
    r.name AS repository_name
   FROM (lsif_indexes u
     JOIN repo r ON ((r.id = u.repository_id)))
  WHERE (r.deleted_at IS NULL);

CREATE TABLE lsif_last_index_scan (
    repository_id integer NOT NULL,
    last_index_scan_at timestamp with time zone NOT NULL
);

COMMENT ON TABLE lsif_last_index_scan IS 'Tracks the last time repository was checked for auto-indexing job scheduling.';

COMMENT ON COLUMN lsif_last_index_scan.last_index_scan_at IS 'The last time uploads of this repository were considered for auto-indexing job scheduling.';

CREATE TABLE lsif_last_retention_scan (
    repository_id integer NOT NULL,
    last_retention_scan_at timestamp with time zone NOT NULL
);

COMMENT ON TABLE lsif_last_retention_scan IS 'Tracks the last time uploads a repository were checked against data retention policies.';

COMMENT ON COLUMN lsif_last_retention_scan.last_retention_scan_at IS 'The last time uploads of this repository were checked against data retention policies.';

CREATE TABLE lsif_nearest_uploads (
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    uploads jsonb NOT NULL
);

COMMENT ON TABLE lsif_nearest_uploads IS 'Associates commits with the complete set of uploads visible from that commit. Every commit with upload data is present in this table.';

COMMENT ON COLUMN lsif_nearest_uploads.commit_bytea IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN lsif_nearest_uploads.uploads IS 'Encodes an {upload_id => distance} map that includes an entry for every upload visible from the commit. There is always at least one entry with a distance of zero.';

CREATE TABLE lsif_nearest_uploads_links (
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    ancestor_commit_bytea bytea NOT NULL,
    distance integer NOT NULL
);

COMMENT ON TABLE lsif_nearest_uploads_links IS 'Associates commits with the closest ancestor commit with usable upload data. Together, this table and lsif_nearest_uploads cover all commits with resolvable code intelligence.';

COMMENT ON COLUMN lsif_nearest_uploads_links.commit_bytea IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN lsif_nearest_uploads_links.ancestor_commit_bytea IS 'The 40-char revhash of the ancestor. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN lsif_nearest_uploads_links.distance IS 'The distance bewteen the commits. Parent = 1, Grandparent = 2, etc.';

CREATE TABLE lsif_packages (
    id integer NOT NULL,
    scheme text NOT NULL,
    name text NOT NULL,
    version text,
    dump_id integer NOT NULL
);

COMMENT ON TABLE lsif_packages IS 'Associates an upload with the set of packages they provide within a given packages management scheme.';

COMMENT ON COLUMN lsif_packages.scheme IS 'The (export) moniker scheme.';

COMMENT ON COLUMN lsif_packages.name IS 'The package name.';

COMMENT ON COLUMN lsif_packages.version IS 'The package version.';

COMMENT ON COLUMN lsif_packages.dump_id IS 'The identifier of the upload that provides the package.';

CREATE SEQUENCE lsif_packages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_packages_id_seq OWNED BY lsif_packages.id;

CREATE TABLE lsif_references (
    id integer NOT NULL,
    scheme text NOT NULL,
    name text NOT NULL,
    version text,
    filter bytea NOT NULL,
    dump_id integer NOT NULL
);

COMMENT ON TABLE lsif_references IS 'Associates an upload with the set of packages they require within a given packages management scheme.';

COMMENT ON COLUMN lsif_references.scheme IS 'The (import) moniker scheme.';

COMMENT ON COLUMN lsif_references.name IS 'The package name.';

COMMENT ON COLUMN lsif_references.version IS 'The package version.';

COMMENT ON COLUMN lsif_references.filter IS 'A [bloom filter](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/bloomfilter/bloom_filter.go#L27:6) encoded as gzipped JSON. This bloom filter stores the set of identifiers imported from the package.';

COMMENT ON COLUMN lsif_references.dump_id IS 'The identifier of the upload that references the package.';

CREATE SEQUENCE lsif_references_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_references_id_seq OWNED BY lsif_references.id;

CREATE TABLE lsif_retention_configuration (
    id integer NOT NULL,
    repository_id integer NOT NULL,
    max_age_for_non_stale_branches_seconds integer NOT NULL,
    max_age_for_non_stale_tags_seconds integer NOT NULL
);

COMMENT ON TABLE lsif_retention_configuration IS 'Stores the retention policy of code intellience data for a repository.';

COMMENT ON COLUMN lsif_retention_configuration.max_age_for_non_stale_branches_seconds IS 'The number of seconds since the last modification of a branch until it is considered stale.';

COMMENT ON COLUMN lsif_retention_configuration.max_age_for_non_stale_tags_seconds IS 'The nujmber of seconds since the commit date of a tagged commit until it is considered stale.';

CREATE SEQUENCE lsif_retention_configuration_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_retention_configuration_id_seq OWNED BY lsif_retention_configuration.id;

CREATE TABLE lsif_uploads_visible_at_tip (
    repository_id integer NOT NULL,
    upload_id integer NOT NULL,
    branch_or_tag_name text DEFAULT ''::text NOT NULL,
    is_default_branch boolean DEFAULT false NOT NULL
);

COMMENT ON TABLE lsif_uploads_visible_at_tip IS 'Associates a repository with the set of LSIF upload identifiers that can serve intelligence for the tip of the default branch.';

COMMENT ON COLUMN lsif_uploads_visible_at_tip.upload_id IS 'The identifier of the upload visible from the tip of the specified branch or tag.';

COMMENT ON COLUMN lsif_uploads_visible_at_tip.branch_or_tag_name IS 'The name of the branch or tag.';

COMMENT ON COLUMN lsif_uploads_visible_at_tip.is_default_branch IS 'Whether the specified branch is the default of the repository. Always false for tags.';

CREATE VIEW lsif_uploads_with_repository_name AS
 SELECT u.id,
    u.commit,
    u.root,
    u.queued_at,
    u.uploaded_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.indexer,
    u.indexer_version,
    u.num_parts,
    u.uploaded_parts,
    u.process_after,
    u.num_resets,
    u.upload_size,
    u.num_failures,
    u.associated_index_id,
    u.expired,
    u.last_retention_scan_at,
    r.name AS repository_name
   FROM (lsif_uploads u
     JOIN repo r ON ((r.id = u.repository_id)))
  WHERE (r.deleted_at IS NULL);

CREATE TABLE names (
    name citext NOT NULL,
    user_id integer,
    org_id integer,
    CONSTRAINT names_check CHECK (((user_id IS NOT NULL) OR (org_id IS NOT NULL)))
);

CREATE TABLE notebook_stars (
    notebook_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE notebooks (
    id bigint NOT NULL,
    title text NOT NULL,
    blocks jsonb DEFAULT '[]'::jsonb NOT NULL,
    public boolean NOT NULL,
    creator_user_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    blocks_tsvector tsvector GENERATED ALWAYS AS (jsonb_to_tsvector('english'::regconfig, blocks, '["string"]'::jsonb)) STORED,
    namespace_user_id integer,
    namespace_org_id integer,
    updater_user_id integer,
    CONSTRAINT blocks_is_array CHECK ((jsonb_typeof(blocks) = 'array'::text)),
    CONSTRAINT notebooks_has_max_1_namespace CHECK ((((namespace_user_id IS NULL) AND (namespace_org_id IS NULL)) OR ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL))))
);

CREATE SEQUENCE notebooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE notebooks_id_seq OWNED BY notebooks.id;

CREATE TABLE org_invitations (
    id bigint NOT NULL,
    org_id integer NOT NULL,
    sender_user_id integer NOT NULL,
    recipient_user_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    notified_at timestamp with time zone,
    responded_at timestamp with time zone,
    response_type boolean,
    revoked_at timestamp with time zone,
    deleted_at timestamp with time zone,
    recipient_email citext,
    expires_at timestamp with time zone,
    CONSTRAINT check_atomic_response CHECK (((responded_at IS NULL) = (response_type IS NULL))),
    CONSTRAINT check_single_use CHECK ((((responded_at IS NULL) AND (response_type IS NULL)) OR (revoked_at IS NULL))),
    CONSTRAINT either_user_id_or_email_defined CHECK (((recipient_user_id IS NULL) <> (recipient_email IS NULL)))
);

CREATE SEQUENCE org_invitations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE org_invitations_id_seq OWNED BY org_invitations.id;

CREATE TABLE org_members (
    id integer NOT NULL,
    org_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id integer NOT NULL
);

CREATE SEQUENCE org_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE org_members_id_seq OWNED BY org_members.id;

CREATE TABLE org_stats (
    org_id integer NOT NULL,
    code_host_repo_count integer DEFAULT 0,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE org_stats IS 'Business statistics for organizations';

COMMENT ON COLUMN org_stats.org_id IS 'Org ID that the stats relate to.';

COMMENT ON COLUMN org_stats.code_host_repo_count IS 'Count of repositories accessible on all code hosts for this organization.';

CREATE TABLE orgs (
    id integer NOT NULL,
    name citext NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_name text,
    slack_webhook_url text,
    deleted_at timestamp with time zone,
    CONSTRAINT orgs_display_name_max_length CHECK ((char_length(display_name) <= 255)),
    CONSTRAINT orgs_name_max_length CHECK ((char_length((name)::text) <= 255)),
    CONSTRAINT orgs_name_valid_chars CHECK ((name OPERATOR(~) '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'::citext))
);

CREATE SEQUENCE orgs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE orgs_id_seq OWNED BY orgs.id;

CREATE TABLE orgs_open_beta_stats (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id integer,
    org_id integer,
    created_at timestamp with time zone DEFAULT now(),
    data jsonb DEFAULT '{}'::jsonb NOT NULL
);

CREATE TABLE out_of_band_migrations (
    id integer NOT NULL,
    team text NOT NULL,
    component text NOT NULL,
    description text NOT NULL,
    progress double precision DEFAULT 0 NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    last_updated timestamp with time zone,
    non_destructive boolean NOT NULL,
    apply_reverse boolean DEFAULT false NOT NULL,
    is_enterprise boolean DEFAULT false NOT NULL,
    introduced_version_major integer NOT NULL,
    introduced_version_minor integer NOT NULL,
    deprecated_version_major integer,
    deprecated_version_minor integer,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT out_of_band_migrations_component_nonempty CHECK ((component <> ''::text)),
    CONSTRAINT out_of_band_migrations_description_nonempty CHECK ((description <> ''::text)),
    CONSTRAINT out_of_band_migrations_progress_range CHECK (((progress >= (0)::double precision) AND (progress <= (1)::double precision))),
    CONSTRAINT out_of_band_migrations_team_nonempty CHECK ((team <> ''::text))
);

COMMENT ON TABLE out_of_band_migrations IS 'Stores metadata and progress about an out-of-band migration routine.';

COMMENT ON COLUMN out_of_band_migrations.id IS 'A globally unique primary key for this migration. The same key is used consistently across all Sourcegraph instances for the same migration.';

COMMENT ON COLUMN out_of_band_migrations.team IS 'The name of the engineering team responsible for the migration.';

COMMENT ON COLUMN out_of_band_migrations.component IS 'The name of the component undergoing a migration.';

COMMENT ON COLUMN out_of_band_migrations.description IS 'A brief description about the migration.';

COMMENT ON COLUMN out_of_band_migrations.progress IS 'The percentage progress in the up direction (0=0%, 1=100%).';

COMMENT ON COLUMN out_of_band_migrations.created IS 'The date and time the migration was inserted into the database (via an upgrade).';

COMMENT ON COLUMN out_of_band_migrations.last_updated IS 'The date and time the migration was last updated.';

COMMENT ON COLUMN out_of_band_migrations.non_destructive IS 'Whether or not this migration alters data so it can no longer be read by the previous Sourcegraph instance.';

COMMENT ON COLUMN out_of_band_migrations.apply_reverse IS 'Whether this migration should run in the opposite direction (to support an upcoming downgrade).';

COMMENT ON COLUMN out_of_band_migrations.is_enterprise IS 'When true, these migrations are invisible to OSS mode.';

COMMENT ON COLUMN out_of_band_migrations.introduced_version_major IS 'The Sourcegraph version (major component) in which this migration was first introduced.';

COMMENT ON COLUMN out_of_band_migrations.introduced_version_minor IS 'The Sourcegraph version (minor component) in which this migration was first introduced.';

COMMENT ON COLUMN out_of_band_migrations.deprecated_version_major IS 'The lowest Sourcegraph version (major component) that assumes the migration has completed.';

COMMENT ON COLUMN out_of_band_migrations.deprecated_version_minor IS 'The lowest Sourcegraph version (minor component) that assumes the migration has completed.';

CREATE TABLE out_of_band_migrations_errors (
    id integer NOT NULL,
    migration_id integer NOT NULL,
    message text NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT out_of_band_migrations_errors_message_nonempty CHECK ((message <> ''::text))
);

COMMENT ON TABLE out_of_band_migrations_errors IS 'Stores errors that occurred while performing an out-of-band migration.';

COMMENT ON COLUMN out_of_band_migrations_errors.id IS 'A unique identifer.';

COMMENT ON COLUMN out_of_band_migrations_errors.migration_id IS 'The identifier of the migration.';

COMMENT ON COLUMN out_of_band_migrations_errors.message IS 'The error message.';

COMMENT ON COLUMN out_of_band_migrations_errors.created IS 'The date and time the error occurred.';

CREATE SEQUENCE out_of_band_migrations_errors_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE out_of_band_migrations_errors_id_seq OWNED BY out_of_band_migrations_errors.id;

CREATE SEQUENCE out_of_band_migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE out_of_band_migrations_id_seq OWNED BY out_of_band_migrations.id;

CREATE TABLE phabricator_repos (
    id integer NOT NULL,
    callsign citext NOT NULL,
    repo_name citext NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    url text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE phabricator_repos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE phabricator_repos_id_seq OWNED BY phabricator_repos.id;

CREATE TABLE product_licenses (
    id uuid NOT NULL,
    product_subscription_id uuid NOT NULL,
    license_key text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE product_subscriptions (
    id uuid NOT NULL,
    user_id integer NOT NULL,
    billing_subscription_id text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone
);

CREATE TABLE query_runner_state (
    query text,
    last_executed timestamp with time zone,
    latest_result timestamp with time zone,
    exec_duration_ns bigint
);

CREATE TABLE users (
    id integer NOT NULL,
    username citext NOT NULL,
    display_name text,
    avatar_url text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    invite_quota integer DEFAULT 15 NOT NULL,
    passwd text,
    passwd_reset_code text,
    passwd_reset_time timestamp with time zone,
    site_admin boolean DEFAULT false NOT NULL,
    page_views integer DEFAULT 0 NOT NULL,
    search_queries integer DEFAULT 0 NOT NULL,
    tags text[] DEFAULT '{}'::text[],
    billing_customer_id text,
    invalidated_sessions_at timestamp with time zone DEFAULT now() NOT NULL,
    tos_accepted boolean DEFAULT false NOT NULL,
    searchable boolean DEFAULT true NOT NULL,
    CONSTRAINT users_display_name_max_length CHECK ((char_length(display_name) <= 255)),
    CONSTRAINT users_username_max_length CHECK ((char_length((username)::text) <= 255)),
    CONSTRAINT users_username_valid_chars CHECK ((username OPERATOR(~) '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'::citext))
);

CREATE VIEW reconciler_changesets AS
 SELECT c.id,
    c.batch_change_ids,
    c.repo_id,
    c.queued_at,
    c.created_at,
    c.updated_at,
    c.metadata,
    c.external_id,
    c.external_service_type,
    c.external_deleted_at,
    c.external_branch,
    c.external_updated_at,
    c.external_state,
    c.external_review_state,
    c.external_check_state,
    c.diff_stat_added,
    c.diff_stat_changed,
    c.diff_stat_deleted,
    c.sync_state,
    c.current_spec_id,
    c.previous_spec_id,
    c.publication_state,
    c.owned_by_batch_change_id,
    c.reconciler_state,
    c.failure_message,
    c.started_at,
    c.finished_at,
    c.process_after,
    c.num_resets,
    c.closing,
    c.num_failures,
    c.log_contents,
    c.execution_logs,
    c.syncer_error,
    c.external_title,
    c.worker_hostname,
    c.ui_publication_state,
    c.last_heartbeat_at,
    c.external_fork_namespace
   FROM (changesets c
     JOIN repo r ON ((r.id = c.repo_id)))
  WHERE ((r.deleted_at IS NULL) AND (EXISTS ( SELECT 1
           FROM ((batch_changes
             LEFT JOIN users namespace_user ON ((batch_changes.namespace_user_id = namespace_user.id)))
             LEFT JOIN orgs namespace_org ON ((batch_changes.namespace_org_id = namespace_org.id)))
          WHERE ((c.batch_change_ids ? (batch_changes.id)::text) AND (namespace_user.deleted_at IS NULL) AND (namespace_org.deleted_at IS NULL)))));

CREATE TABLE registry_extension_releases (
    id bigint NOT NULL,
    registry_extension_id integer NOT NULL,
    creator_user_id integer NOT NULL,
    release_version citext,
    release_tag citext NOT NULL,
    manifest jsonb NOT NULL,
    bundle text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    source_map text
);

CREATE SEQUENCE registry_extension_releases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE registry_extension_releases_id_seq OWNED BY registry_extension_releases.id;

CREATE TABLE registry_extensions (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    publisher_user_id integer,
    publisher_org_id integer,
    name citext NOT NULL,
    manifest text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT registry_extensions_name_length CHECK (((char_length((name)::text) > 0) AND (char_length((name)::text) <= 128))),
    CONSTRAINT registry_extensions_name_valid_chars CHECK ((name OPERATOR(~) '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[_.-](?=[a-zA-Z0-9]))*$'::citext)),
    CONSTRAINT registry_extensions_single_publisher CHECK (((publisher_user_id IS NULL) <> (publisher_org_id IS NULL)))
);

CREATE SEQUENCE registry_extensions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE registry_extensions_id_seq OWNED BY registry_extensions.id;

CREATE SEQUENCE repo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE repo_id_seq OWNED BY repo.id;

CREATE TABLE repo_pending_permissions (
    repo_id integer NOT NULL,
    permission text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    user_ids_ints integer[] DEFAULT '{}'::integer[] NOT NULL
);

CREATE TABLE repo_permissions (
    repo_id integer NOT NULL,
    permission text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    synced_at timestamp with time zone,
    user_ids_ints integer[] DEFAULT '{}'::integer[] NOT NULL
);

CREATE TABLE saved_searches (
    id integer NOT NULL,
    description text NOT NULL,
    query text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    notify_owner boolean NOT NULL,
    notify_slack boolean NOT NULL,
    user_id integer,
    org_id integer,
    slack_webhook_url text,
    CONSTRAINT saved_searches_notifications_disabled CHECK (((notify_owner = false) AND (notify_slack = false))),
    CONSTRAINT user_or_org_id_not_null CHECK ((((user_id IS NOT NULL) AND (org_id IS NULL)) OR ((org_id IS NOT NULL) AND (user_id IS NULL))))
);

CREATE SEQUENCE saved_searches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE saved_searches_id_seq OWNED BY saved_searches.id;

CREATE TABLE search_context_repos (
    search_context_id bigint NOT NULL,
    repo_id integer NOT NULL,
    revision text NOT NULL
);

CREATE TABLE search_contexts (
    id bigint NOT NULL,
    name citext NOT NULL,
    description text NOT NULL,
    public boolean NOT NULL,
    namespace_user_id integer,
    namespace_org_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    query text,
    CONSTRAINT search_contexts_has_one_or_no_namespace CHECK (((namespace_user_id IS NULL) OR (namespace_org_id IS NULL)))
);

COMMENT ON COLUMN search_contexts.deleted_at IS 'This column is unused as of Sourcegraph 3.34. Do not refer to it anymore. It will be dropped in a future version.';

CREATE SEQUENCE search_contexts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE search_contexts_id_seq OWNED BY search_contexts.id;

CREATE TABLE security_event_logs (
    id bigint NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    user_id integer NOT NULL,
    anonymous_user_id text NOT NULL,
    source text NOT NULL,
    argument jsonb NOT NULL,
    version text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    CONSTRAINT security_event_logs_check_has_user CHECK ((((user_id = 0) AND (anonymous_user_id <> ''::text)) OR ((user_id <> 0) AND (anonymous_user_id = ''::text)) OR ((user_id <> 0) AND (anonymous_user_id <> ''::text)))),
    CONSTRAINT security_event_logs_check_name_not_empty CHECK ((name <> ''::text)),
    CONSTRAINT security_event_logs_check_source_not_empty CHECK ((source <> ''::text)),
    CONSTRAINT security_event_logs_check_version_not_empty CHECK ((version <> ''::text))
);

COMMENT ON TABLE security_event_logs IS 'Contains security-relevant events with a long time horizon for storage.';

COMMENT ON COLUMN security_event_logs.name IS 'The event name as a CAPITALIZED_SNAKE_CASE string.';

COMMENT ON COLUMN security_event_logs.url IS 'The URL within the Sourcegraph app which generated the event.';

COMMENT ON COLUMN security_event_logs.user_id IS 'The ID of the actor associated with the event.';

COMMENT ON COLUMN security_event_logs.anonymous_user_id IS 'The UUID of the actor associated with the event.';

COMMENT ON COLUMN security_event_logs.source IS 'The site section (WEB, BACKEND, etc.) that generated the event.';

COMMENT ON COLUMN security_event_logs.argument IS 'An arbitrary JSON blob containing event data.';

COMMENT ON COLUMN security_event_logs.version IS 'The version of Sourcegraph which generated the event.';

CREATE SEQUENCE security_event_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE security_event_logs_id_seq OWNED BY security_event_logs.id;

CREATE TABLE settings (
    id integer NOT NULL,
    org_id integer,
    contents text DEFAULT '{}'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id integer,
    author_user_id integer,
    CONSTRAINT settings_no_empty_contents CHECK ((contents <> ''::text))
);

CREATE SEQUENCE settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE settings_id_seq OWNED BY settings.id;

CREATE VIEW site_config AS
 SELECT global_state.site_id,
    global_state.initialized
   FROM global_state;

CREATE TABLE sub_repo_permissions (
    repo_id integer NOT NULL,
    user_id integer NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    path_includes text[],
    path_excludes text[],
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE sub_repo_permissions IS 'Responsible for storing permissions at a finer granularity than repo';

CREATE TABLE survey_responses (
    id bigint NOT NULL,
    user_id integer,
    email text,
    score integer NOT NULL,
    reason text,
    better text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE survey_responses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE survey_responses_id_seq OWNED BY survey_responses.id;

CREATE TABLE temporary_settings (
    id integer NOT NULL,
    user_id integer NOT NULL,
    contents jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE temporary_settings IS 'Stores per-user temporary settings used in the UI, for example, which modals have been dimissed or what theme is preferred.';

COMMENT ON COLUMN temporary_settings.user_id IS 'The ID of the user the settings will be saved for.';

COMMENT ON COLUMN temporary_settings.contents IS 'JSON-encoded temporary settings.';

CREATE SEQUENCE temporary_settings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE temporary_settings_id_seq OWNED BY temporary_settings.id;

CREATE VIEW tracking_changeset_specs_and_changesets AS
 SELECT changeset_specs.id AS changeset_spec_id,
    COALESCE(changesets.id, (0)::bigint) AS changeset_id,
    changeset_specs.repo_id,
    changeset_specs.batch_spec_id,
    repo.name AS repo_name,
    COALESCE((changesets.metadata ->> 'Title'::text), (changesets.metadata ->> 'title'::text)) AS changeset_name,
    changesets.external_state,
    changesets.publication_state,
    changesets.reconciler_state
   FROM ((changeset_specs
     LEFT JOIN changesets ON (((changesets.repo_id = changeset_specs.repo_id) AND (changesets.external_id = changeset_specs.external_id))))
     JOIN repo ON ((changeset_specs.repo_id = repo.id)))
  WHERE ((changeset_specs.external_id IS NOT NULL) AND (repo.deleted_at IS NULL));

CREATE TABLE user_credentials (
    id bigint NOT NULL,
    domain text NOT NULL,
    user_id integer NOT NULL,
    external_service_type text NOT NULL,
    external_service_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    credential bytea NOT NULL,
    ssh_migration_applied boolean DEFAULT false NOT NULL,
    encryption_key_id text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE user_credentials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE user_credentials_id_seq OWNED BY user_credentials.id;

CREATE TABLE user_emails (
    user_id integer NOT NULL,
    email citext NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    verification_code text,
    verified_at timestamp with time zone,
    last_verification_sent_at timestamp with time zone,
    is_primary boolean DEFAULT false NOT NULL
);

CREATE TABLE user_external_accounts (
    id integer NOT NULL,
    user_id integer NOT NULL,
    service_type text NOT NULL,
    service_id text NOT NULL,
    account_id text NOT NULL,
    auth_data text,
    account_data text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    client_id text NOT NULL,
    expired_at timestamp with time zone,
    last_valid_at timestamp with time zone,
    encryption_key_id text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE user_external_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE user_external_accounts_id_seq OWNED BY user_external_accounts.id;

CREATE TABLE user_pending_permissions (
    id integer NOT NULL,
    bind_id text NOT NULL,
    permission text NOT NULL,
    object_type text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    service_type text NOT NULL,
    service_id text NOT NULL,
    object_ids_ints integer[] DEFAULT '{}'::integer[] NOT NULL
);

CREATE SEQUENCE user_pending_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE user_pending_permissions_id_seq OWNED BY user_pending_permissions.id;

CREATE TABLE user_permissions (
    user_id integer NOT NULL,
    permission text NOT NULL,
    object_type text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    synced_at timestamp with time zone,
    object_ids_ints integer[] DEFAULT '{}'::integer[] NOT NULL
);

CREATE TABLE user_public_repos (
    user_id integer NOT NULL,
    repo_uri text NOT NULL,
    repo_id integer NOT NULL
);

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE users_id_seq OWNED BY users.id;

CREATE TABLE versions (
    service text NOT NULL,
    version text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    first_version text NOT NULL
);

CREATE TABLE webhook_logs (
    id bigint NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    external_service_id integer,
    status_code integer NOT NULL,
    request bytea NOT NULL,
    response bytea NOT NULL,
    encryption_key_id text NOT NULL
);

CREATE SEQUENCE webhook_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE webhook_logs_id_seq OWNED BY webhook_logs.id;

ALTER TABLE ONLY access_tokens ALTER COLUMN id SET DEFAULT nextval('access_tokens_id_seq'::regclass);

ALTER TABLE ONLY batch_changes ALTER COLUMN id SET DEFAULT nextval('batch_changes_id_seq'::regclass);

ALTER TABLE ONLY batch_changes_site_credentials ALTER COLUMN id SET DEFAULT nextval('batch_changes_site_credentials_id_seq'::regclass);

ALTER TABLE ONLY batch_spec_execution_cache_entries ALTER COLUMN id SET DEFAULT nextval('batch_spec_execution_cache_entries_id_seq'::regclass);

ALTER TABLE ONLY batch_spec_resolution_jobs ALTER COLUMN id SET DEFAULT nextval('batch_spec_resolution_jobs_id_seq'::regclass);

ALTER TABLE ONLY batch_spec_workspace_execution_jobs ALTER COLUMN id SET DEFAULT nextval('batch_spec_workspace_execution_jobs_id_seq'::regclass);

ALTER TABLE ONLY batch_spec_workspaces ALTER COLUMN id SET DEFAULT nextval('batch_spec_workspaces_id_seq'::regclass);

ALTER TABLE ONLY batch_specs ALTER COLUMN id SET DEFAULT nextval('batch_specs_id_seq'::regclass);

ALTER TABLE ONLY changeset_events ALTER COLUMN id SET DEFAULT nextval('changeset_events_id_seq'::regclass);

ALTER TABLE ONLY changeset_jobs ALTER COLUMN id SET DEFAULT nextval('changeset_jobs_id_seq'::regclass);

ALTER TABLE ONLY changeset_specs ALTER COLUMN id SET DEFAULT nextval('changeset_specs_id_seq'::regclass);

ALTER TABLE ONLY changesets ALTER COLUMN id SET DEFAULT nextval('changesets_id_seq'::regclass);

ALTER TABLE ONLY cm_action_jobs ALTER COLUMN id SET DEFAULT nextval('cm_action_jobs_id_seq'::regclass);

ALTER TABLE ONLY cm_emails ALTER COLUMN id SET DEFAULT nextval('cm_emails_id_seq'::regclass);

ALTER TABLE ONLY cm_monitors ALTER COLUMN id SET DEFAULT nextval('cm_monitors_id_seq'::regclass);

ALTER TABLE ONLY cm_queries ALTER COLUMN id SET DEFAULT nextval('cm_queries_id_seq'::regclass);

ALTER TABLE ONLY cm_recipients ALTER COLUMN id SET DEFAULT nextval('cm_recipients_id_seq'::regclass);

ALTER TABLE ONLY cm_slack_webhooks ALTER COLUMN id SET DEFAULT nextval('cm_slack_webhooks_id_seq'::regclass);

ALTER TABLE ONLY cm_trigger_jobs ALTER COLUMN id SET DEFAULT nextval('cm_trigger_jobs_id_seq'::regclass);

ALTER TABLE ONLY cm_webhooks ALTER COLUMN id SET DEFAULT nextval('cm_webhooks_id_seq'::regclass);

ALTER TABLE ONLY critical_and_site_config ALTER COLUMN id SET DEFAULT nextval('critical_and_site_config_id_seq'::regclass);

ALTER TABLE ONLY discussion_comments ALTER COLUMN id SET DEFAULT nextval('discussion_comments_id_seq'::regclass);

ALTER TABLE ONLY discussion_threads ALTER COLUMN id SET DEFAULT nextval('discussion_threads_id_seq'::regclass);

ALTER TABLE ONLY discussion_threads_target_repo ALTER COLUMN id SET DEFAULT nextval('discussion_threads_target_repo_id_seq'::regclass);

ALTER TABLE ONLY event_logs ALTER COLUMN id SET DEFAULT nextval('event_logs_id_seq'::regclass);

ALTER TABLE ONLY executor_heartbeats ALTER COLUMN id SET DEFAULT nextval('executor_heartbeats_id_seq'::regclass);

ALTER TABLE ONLY external_services ALTER COLUMN id SET DEFAULT nextval('external_services_id_seq'::regclass);

ALTER TABLE ONLY gitserver_localclone_jobs ALTER COLUMN id SET DEFAULT nextval('gitserver_localclone_jobs_id_seq'::regclass);

ALTER TABLE ONLY insights_query_runner_jobs ALTER COLUMN id SET DEFAULT nextval('insights_query_runner_jobs_id_seq'::regclass);

ALTER TABLE ONLY insights_query_runner_jobs_dependencies ALTER COLUMN id SET DEFAULT nextval('insights_query_runner_jobs_dependencies_id_seq'::regclass);

ALTER TABLE ONLY insights_settings_migration_jobs ALTER COLUMN id SET DEFAULT nextval('insights_settings_migration_jobs_id_seq'::regclass);

ALTER TABLE ONLY lsif_configuration_policies ALTER COLUMN id SET DEFAULT nextval('lsif_configuration_policies_id_seq'::regclass);

ALTER TABLE ONLY lsif_dependency_indexing_jobs ALTER COLUMN id SET DEFAULT nextval('lsif_dependency_indexing_jobs_id_seq1'::regclass);

ALTER TABLE ONLY lsif_dependency_repos ALTER COLUMN id SET DEFAULT nextval('lsif_dependency_repos_id_seq'::regclass);

ALTER TABLE ONLY lsif_dependency_syncing_jobs ALTER COLUMN id SET DEFAULT nextval('lsif_dependency_indexing_jobs_id_seq'::regclass);

ALTER TABLE ONLY lsif_index_configuration ALTER COLUMN id SET DEFAULT nextval('lsif_index_configuration_id_seq'::regclass);

ALTER TABLE ONLY lsif_indexes ALTER COLUMN id SET DEFAULT nextval('lsif_indexes_id_seq'::regclass);

ALTER TABLE ONLY lsif_packages ALTER COLUMN id SET DEFAULT nextval('lsif_packages_id_seq'::regclass);

ALTER TABLE ONLY lsif_references ALTER COLUMN id SET DEFAULT nextval('lsif_references_id_seq'::regclass);

ALTER TABLE ONLY lsif_retention_configuration ALTER COLUMN id SET DEFAULT nextval('lsif_retention_configuration_id_seq'::regclass);

ALTER TABLE ONLY lsif_uploads ALTER COLUMN id SET DEFAULT nextval('lsif_dumps_id_seq'::regclass);

ALTER TABLE ONLY notebooks ALTER COLUMN id SET DEFAULT nextval('notebooks_id_seq'::regclass);

ALTER TABLE ONLY org_invitations ALTER COLUMN id SET DEFAULT nextval('org_invitations_id_seq'::regclass);

ALTER TABLE ONLY org_members ALTER COLUMN id SET DEFAULT nextval('org_members_id_seq'::regclass);

ALTER TABLE ONLY orgs ALTER COLUMN id SET DEFAULT nextval('orgs_id_seq'::regclass);

ALTER TABLE ONLY out_of_band_migrations ALTER COLUMN id SET DEFAULT nextval('out_of_band_migrations_id_seq'::regclass);

ALTER TABLE ONLY out_of_band_migrations_errors ALTER COLUMN id SET DEFAULT nextval('out_of_band_migrations_errors_id_seq'::regclass);

ALTER TABLE ONLY phabricator_repos ALTER COLUMN id SET DEFAULT nextval('phabricator_repos_id_seq'::regclass);

ALTER TABLE ONLY registry_extension_releases ALTER COLUMN id SET DEFAULT nextval('registry_extension_releases_id_seq'::regclass);

ALTER TABLE ONLY registry_extensions ALTER COLUMN id SET DEFAULT nextval('registry_extensions_id_seq'::regclass);

ALTER TABLE ONLY repo ALTER COLUMN id SET DEFAULT nextval('repo_id_seq'::regclass);

ALTER TABLE ONLY saved_searches ALTER COLUMN id SET DEFAULT nextval('saved_searches_id_seq'::regclass);

ALTER TABLE ONLY search_contexts ALTER COLUMN id SET DEFAULT nextval('search_contexts_id_seq'::regclass);

ALTER TABLE ONLY security_event_logs ALTER COLUMN id SET DEFAULT nextval('security_event_logs_id_seq'::regclass);

ALTER TABLE ONLY settings ALTER COLUMN id SET DEFAULT nextval('settings_id_seq'::regclass);

ALTER TABLE ONLY survey_responses ALTER COLUMN id SET DEFAULT nextval('survey_responses_id_seq'::regclass);

ALTER TABLE ONLY temporary_settings ALTER COLUMN id SET DEFAULT nextval('temporary_settings_id_seq'::regclass);

ALTER TABLE ONLY user_credentials ALTER COLUMN id SET DEFAULT nextval('user_credentials_id_seq'::regclass);

ALTER TABLE ONLY user_external_accounts ALTER COLUMN id SET DEFAULT nextval('user_external_accounts_id_seq'::regclass);

ALTER TABLE ONLY user_pending_permissions ALTER COLUMN id SET DEFAULT nextval('user_pending_permissions_id_seq'::regclass);

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

ALTER TABLE ONLY webhook_logs ALTER COLUMN id SET DEFAULT nextval('webhook_logs_id_seq'::regclass);

ALTER TABLE ONLY access_tokens
    ADD CONSTRAINT access_tokens_pkey PRIMARY KEY (id);

ALTER TABLE ONLY access_tokens
    ADD CONSTRAINT access_tokens_value_sha256_key UNIQUE (value_sha256);

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_changes_pkey PRIMARY KEY (id);

ALTER TABLE ONLY batch_changes_site_credentials
    ADD CONSTRAINT batch_changes_site_credentials_pkey PRIMARY KEY (id);

ALTER TABLE ONLY batch_spec_execution_cache_entries
    ADD CONSTRAINT batch_spec_execution_cache_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY batch_spec_execution_cache_entries
    ADD CONSTRAINT batch_spec_execution_cache_entries_user_id_key_unique UNIQUE (user_id, key);

ALTER TABLE ONLY batch_spec_resolution_jobs
    ADD CONSTRAINT batch_spec_resolution_jobs_batch_spec_id_unique UNIQUE (batch_spec_id);

ALTER TABLE ONLY batch_spec_resolution_jobs
    ADD CONSTRAINT batch_spec_resolution_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY batch_spec_workspace_execution_jobs
    ADD CONSTRAINT batch_spec_workspace_execution_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY batch_spec_workspaces
    ADD CONSTRAINT batch_spec_workspaces_pkey PRIMARY KEY (id);

ALTER TABLE ONLY batch_specs
    ADD CONSTRAINT batch_specs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY changeset_events
    ADD CONSTRAINT changeset_events_changeset_id_kind_key_unique UNIQUE (changeset_id, kind, key);

ALTER TABLE ONLY changeset_events
    ADD CONSTRAINT changeset_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY changeset_jobs
    ADD CONSTRAINT changeset_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY changeset_specs
    ADD CONSTRAINT changeset_specs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY changesets
    ADD CONSTRAINT changesets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY changesets
    ADD CONSTRAINT changesets_repo_external_id_unique UNIQUE (repo_id, external_id);

ALTER TABLE ONLY cm_action_jobs
    ADD CONSTRAINT cm_action_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_emails
    ADD CONSTRAINT cm_emails_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_last_searched
    ADD CONSTRAINT cm_last_searched_pkey PRIMARY KEY (monitor_id, args_hash);

ALTER TABLE ONLY cm_monitors
    ADD CONSTRAINT cm_monitors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_queries
    ADD CONSTRAINT cm_queries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_recipients
    ADD CONSTRAINT cm_recipients_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_slack_webhooks
    ADD CONSTRAINT cm_slack_webhooks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_trigger_jobs
    ADD CONSTRAINT cm_trigger_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY cm_webhooks
    ADD CONSTRAINT cm_webhooks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY critical_and_site_config
    ADD CONSTRAINT critical_and_site_config_pkey PRIMARY KEY (id);

ALTER TABLE ONLY discussion_comments
    ADD CONSTRAINT discussion_comments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY discussion_mail_reply_tokens
    ADD CONSTRAINT discussion_mail_reply_tokens_pkey PRIMARY KEY (token);

ALTER TABLE ONLY discussion_threads
    ADD CONSTRAINT discussion_threads_pkey PRIMARY KEY (id);

ALTER TABLE ONLY discussion_threads_target_repo
    ADD CONSTRAINT discussion_threads_target_repo_pkey PRIMARY KEY (id);

ALTER TABLE ONLY event_logs
    ADD CONSTRAINT event_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY executor_heartbeats
    ADD CONSTRAINT executor_heartbeats_hostname_key UNIQUE (hostname);

ALTER TABLE ONLY executor_heartbeats
    ADD CONSTRAINT executor_heartbeats_pkey PRIMARY KEY (id);

ALTER TABLE ONLY external_service_repos
    ADD CONSTRAINT external_service_repos_repo_id_external_service_id_unique UNIQUE (repo_id, external_service_id);

ALTER TABLE ONLY external_services
    ADD CONSTRAINT external_services_pkey PRIMARY KEY (id);

ALTER TABLE ONLY feature_flag_overrides
    ADD CONSTRAINT feature_flag_overrides_unique_org_flag UNIQUE (namespace_org_id, flag_name);

ALTER TABLE ONLY feature_flag_overrides
    ADD CONSTRAINT feature_flag_overrides_unique_user_flag UNIQUE (namespace_user_id, flag_name);

ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (flag_name);

ALTER TABLE ONLY gitserver_localclone_jobs
    ADD CONSTRAINT gitserver_localclone_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY gitserver_repos
    ADD CONSTRAINT gitserver_repos_pkey PRIMARY KEY (repo_id);

ALTER TABLE ONLY global_state
    ADD CONSTRAINT global_state_pkey PRIMARY KEY (site_id);

ALTER TABLE ONLY insights_query_runner_jobs_dependencies
    ADD CONSTRAINT insights_query_runner_jobs_dependencies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY insights_query_runner_jobs
    ADD CONSTRAINT insights_query_runner_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_configuration_policies
    ADD CONSTRAINT lsif_configuration_policies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_configuration_policies_repository_pattern_lookup
    ADD CONSTRAINT lsif_configuration_policies_repository_pattern_lookup_pkey PRIMARY KEY (policy_id, repo_id);

ALTER TABLE ONLY lsif_dependency_syncing_jobs
    ADD CONSTRAINT lsif_dependency_indexing_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_dependency_indexing_jobs
    ADD CONSTRAINT lsif_dependency_indexing_jobs_pkey1 PRIMARY KEY (id);

ALTER TABLE ONLY lsif_dependency_repos
    ADD CONSTRAINT lsif_dependency_repos_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_dependency_repos
    ADD CONSTRAINT lsif_dependency_repos_unique_triplet UNIQUE (scheme, name, version);

ALTER TABLE ONLY lsif_dirty_repositories
    ADD CONSTRAINT lsif_dirty_repositories_pkey PRIMARY KEY (repository_id);

ALTER TABLE ONLY lsif_index_configuration
    ADD CONSTRAINT lsif_index_configuration_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_index_configuration
    ADD CONSTRAINT lsif_index_configuration_repository_id_key UNIQUE (repository_id);

ALTER TABLE ONLY lsif_indexes
    ADD CONSTRAINT lsif_indexes_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_last_index_scan
    ADD CONSTRAINT lsif_last_index_scan_pkey PRIMARY KEY (repository_id);

ALTER TABLE ONLY lsif_last_retention_scan
    ADD CONSTRAINT lsif_last_retention_scan_pkey PRIMARY KEY (repository_id);

ALTER TABLE ONLY lsif_packages
    ADD CONSTRAINT lsif_packages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_references
    ADD CONSTRAINT lsif_references_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_retention_configuration
    ADD CONSTRAINT lsif_retention_configuration_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_retention_configuration
    ADD CONSTRAINT lsif_retention_configuration_repository_id_key UNIQUE (repository_id);

ALTER TABLE ONLY lsif_uploads
    ADD CONSTRAINT lsif_uploads_pkey PRIMARY KEY (id);

ALTER TABLE ONLY names
    ADD CONSTRAINT names_pkey PRIMARY KEY (name);

ALTER TABLE ONLY notebook_stars
    ADD CONSTRAINT notebook_stars_pkey PRIMARY KEY (notebook_id, user_id);

ALTER TABLE ONLY notebooks
    ADD CONSTRAINT notebooks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY org_invitations
    ADD CONSTRAINT org_invitations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY org_members
    ADD CONSTRAINT org_members_org_id_user_id_key UNIQUE (org_id, user_id);

ALTER TABLE ONLY org_members
    ADD CONSTRAINT org_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY org_stats
    ADD CONSTRAINT org_stats_pkey PRIMARY KEY (org_id);

ALTER TABLE ONLY orgs_open_beta_stats
    ADD CONSTRAINT orgs_open_beta_stats_pkey PRIMARY KEY (id);

ALTER TABLE ONLY orgs
    ADD CONSTRAINT orgs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY out_of_band_migrations_errors
    ADD CONSTRAINT out_of_band_migrations_errors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY out_of_band_migrations
    ADD CONSTRAINT out_of_band_migrations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY phabricator_repos
    ADD CONSTRAINT phabricator_repos_pkey PRIMARY KEY (id);

ALTER TABLE ONLY phabricator_repos
    ADD CONSTRAINT phabricator_repos_repo_name_key UNIQUE (repo_name);

ALTER TABLE ONLY product_licenses
    ADD CONSTRAINT product_licenses_pkey PRIMARY KEY (id);

ALTER TABLE ONLY product_subscriptions
    ADD CONSTRAINT product_subscriptions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY registry_extension_releases
    ADD CONSTRAINT registry_extension_releases_pkey PRIMARY KEY (id);

ALTER TABLE ONLY registry_extensions
    ADD CONSTRAINT registry_extensions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY repo
    ADD CONSTRAINT repo_name_unique UNIQUE (name) DEFERRABLE;

ALTER TABLE ONLY repo_pending_permissions
    ADD CONSTRAINT repo_pending_permissions_perm_unique UNIQUE (repo_id, permission);

ALTER TABLE ONLY repo_permissions
    ADD CONSTRAINT repo_permissions_perm_unique UNIQUE (repo_id, permission);

ALTER TABLE ONLY repo
    ADD CONSTRAINT repo_pkey PRIMARY KEY (id);

ALTER TABLE ONLY saved_searches
    ADD CONSTRAINT saved_searches_pkey PRIMARY KEY (id);

ALTER TABLE ONLY search_context_repos
    ADD CONSTRAINT search_context_repos_unique UNIQUE (repo_id, search_context_id, revision);

ALTER TABLE ONLY search_contexts
    ADD CONSTRAINT search_contexts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY security_event_logs
    ADD CONSTRAINT security_event_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY settings
    ADD CONSTRAINT settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY survey_responses
    ADD CONSTRAINT survey_responses_pkey PRIMARY KEY (id);

ALTER TABLE ONLY temporary_settings
    ADD CONSTRAINT temporary_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY temporary_settings
    ADD CONSTRAINT temporary_settings_user_id_key UNIQUE (user_id);

ALTER TABLE ONLY user_credentials
    ADD CONSTRAINT user_credentials_domain_user_id_external_service_type_exter_key UNIQUE (domain, user_id, external_service_type, external_service_id);

ALTER TABLE ONLY user_credentials
    ADD CONSTRAINT user_credentials_pkey PRIMARY KEY (id);

ALTER TABLE ONLY user_emails
    ADD CONSTRAINT user_emails_no_duplicates_per_user UNIQUE (user_id, email);

ALTER TABLE ONLY user_emails
    ADD CONSTRAINT user_emails_unique_verified_email EXCLUDE USING btree (email WITH OPERATOR(=)) WHERE ((verified_at IS NOT NULL));

ALTER TABLE ONLY user_external_accounts
    ADD CONSTRAINT user_external_accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY user_pending_permissions
    ADD CONSTRAINT user_pending_permissions_service_perm_object_unique UNIQUE (service_type, service_id, permission, object_type, bind_id);

ALTER TABLE ONLY user_permissions
    ADD CONSTRAINT user_permissions_perm_object_unique UNIQUE (user_id, permission, object_type);

ALTER TABLE ONLY user_public_repos
    ADD CONSTRAINT user_public_repos_user_id_repo_id_key UNIQUE (user_id, repo_id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY versions
    ADD CONSTRAINT versions_pkey PRIMARY KEY (service);

ALTER TABLE ONLY webhook_logs
    ADD CONSTRAINT webhook_logs_pkey PRIMARY KEY (id);

CREATE INDEX access_tokens_lookup ON access_tokens USING hash (value_sha256) WHERE (deleted_at IS NULL);

CREATE INDEX batch_changes_namespace_org_id ON batch_changes USING btree (namespace_org_id);

CREATE INDEX batch_changes_namespace_user_id ON batch_changes USING btree (namespace_user_id);

CREATE INDEX batch_changes_site_credentials_credential_idx ON batch_changes_site_credentials USING btree (((encryption_key_id = ANY (ARRAY[''::text, 'previously-migrated'::text]))));

CREATE UNIQUE INDEX batch_changes_site_credentials_unique ON batch_changes_site_credentials USING btree (external_service_type, external_service_id);

CREATE UNIQUE INDEX batch_changes_unique_org_id ON batch_changes USING btree (name, namespace_org_id) WHERE (namespace_org_id IS NOT NULL);

CREATE UNIQUE INDEX batch_changes_unique_user_id ON batch_changes USING btree (name, namespace_user_id) WHERE (namespace_user_id IS NOT NULL);

CREATE INDEX batch_spec_workspace_execution_jobs_cancel ON batch_spec_workspace_execution_jobs USING btree (cancel);

CREATE INDEX batch_specs_rand_id ON batch_specs USING btree (rand_id);

CREATE INDEX changeset_jobs_bulk_group_idx ON changeset_jobs USING btree (bulk_group);

CREATE INDEX changeset_jobs_state_idx ON changeset_jobs USING btree (state);

CREATE INDEX changeset_specs_external_id ON changeset_specs USING btree (external_id);

CREATE INDEX changeset_specs_head_ref ON changeset_specs USING btree (head_ref);

CREATE INDEX changeset_specs_rand_id ON changeset_specs USING btree (rand_id);

CREATE INDEX changeset_specs_title ON changeset_specs USING btree (title);

CREATE INDEX changesets_batch_change_ids ON changesets USING gin (batch_change_ids);

CREATE INDEX changesets_external_state_idx ON changesets USING btree (external_state);

CREATE INDEX changesets_external_title_idx ON changesets USING btree (external_title);

CREATE INDEX changesets_publication_state_idx ON changesets USING btree (publication_state);

CREATE INDEX changesets_reconciler_state_idx ON changesets USING btree (reconciler_state);

CREATE INDEX cm_slack_webhooks_monitor ON cm_slack_webhooks USING btree (monitor);

CREATE INDEX cm_webhooks_monitor ON cm_webhooks USING btree (monitor);

CREATE UNIQUE INDEX critical_and_site_config_unique ON critical_and_site_config USING btree (id, type);

CREATE INDEX discussion_comments_author_user_id_idx ON discussion_comments USING btree (author_user_id);

CREATE INDEX discussion_comments_reports_array_length_idx ON discussion_comments USING btree (array_length(reports, 1));

CREATE INDEX discussion_comments_thread_id_idx ON discussion_comments USING btree (thread_id);

CREATE INDEX discussion_mail_reply_tokens_user_id_thread_id_idx ON discussion_mail_reply_tokens USING btree (user_id, thread_id);

CREATE INDEX discussion_threads_author_user_id_idx ON discussion_threads USING btree (author_user_id);

CREATE INDEX discussion_threads_target_repo_repo_id_path_idx ON discussion_threads_target_repo USING btree (repo_id, path);

CREATE INDEX event_logs_anonymous_user_id ON event_logs USING btree (anonymous_user_id);

CREATE INDEX event_logs_name ON event_logs USING btree (name);

CREATE INDEX event_logs_source ON event_logs USING btree (source);

CREATE INDEX event_logs_timestamp ON event_logs USING btree ("timestamp");

CREATE INDEX event_logs_timestamp_at_utc ON event_logs USING btree (date(timezone('UTC'::text, "timestamp")));

CREATE INDEX event_logs_user_id ON event_logs USING btree (user_id);

CREATE INDEX external_service_repos_clone_url_idx ON external_service_repos USING btree (clone_url);

CREATE INDEX external_service_repos_idx ON external_service_repos USING btree (external_service_id, repo_id);

CREATE INDEX external_service_repos_org_id_idx ON external_service_repos USING btree (org_id) WHERE (org_id IS NOT NULL);

CREATE INDEX external_service_sync_jobs_state_idx ON external_service_sync_jobs USING btree (state);

CREATE INDEX external_service_user_repos_idx ON external_service_repos USING btree (user_id, repo_id) WHERE (user_id IS NOT NULL);

CREATE INDEX external_services_has_webhooks_idx ON external_services USING btree (has_webhooks);

CREATE INDEX external_services_namespace_org_id_idx ON external_services USING btree (namespace_org_id);

CREATE INDEX external_services_namespace_user_id_idx ON external_services USING btree (namespace_user_id);

CREATE INDEX feature_flag_overrides_org_id ON feature_flag_overrides USING btree (namespace_org_id) WHERE (namespace_org_id IS NOT NULL);

CREATE INDEX feature_flag_overrides_user_id ON feature_flag_overrides USING btree (namespace_user_id) WHERE (namespace_user_id IS NOT NULL);

CREATE INDEX gitserver_repos_cloned_status_idx ON gitserver_repos USING btree (repo_id) WHERE (clone_status = 'cloned'::text);

CREATE INDEX gitserver_repos_cloning_status_idx ON gitserver_repos USING btree (repo_id) WHERE (clone_status = 'cloning'::text);

CREATE INDEX gitserver_repos_last_error_idx ON gitserver_repos USING btree (repo_id) WHERE (last_error IS NOT NULL);

CREATE INDEX gitserver_repos_not_cloned_status_idx ON gitserver_repos USING btree (repo_id) WHERE (clone_status = 'not_cloned'::text);

CREATE INDEX gitserver_repos_shard_id ON gitserver_repos USING btree (shard_id, repo_id);

CREATE INDEX insights_query_runner_jobs_cost_idx ON insights_query_runner_jobs USING btree (cost);

CREATE INDEX insights_query_runner_jobs_dependencies_job_id_fk_idx ON insights_query_runner_jobs_dependencies USING btree (job_id);

CREATE INDEX insights_query_runner_jobs_priority_idx ON insights_query_runner_jobs USING btree (priority);

CREATE INDEX insights_query_runner_jobs_processable_priority_id ON insights_query_runner_jobs USING btree (priority, id) WHERE ((state = 'queued'::text) OR (state = 'errored'::text));

CREATE INDEX insights_query_runner_jobs_state_btree ON insights_query_runner_jobs USING btree (state);

CREATE UNIQUE INDEX kind_cloud_default ON external_services USING btree (kind, cloud_default) WHERE ((cloud_default = true) AND (deleted_at IS NULL));

CREATE INDEX lsif_configuration_policies_repository_id ON lsif_configuration_policies USING btree (repository_id);

CREATE INDEX lsif_dependency_indexing_jobs_upload_id ON lsif_dependency_syncing_jobs USING btree (upload_id);

CREATE INDEX lsif_indexes_commit_last_checked_at ON lsif_indexes USING btree (commit_last_checked_at) WHERE (state <> 'deleted'::text);

CREATE INDEX lsif_indexes_repository_id_commit ON lsif_indexes USING btree (repository_id, commit);

CREATE INDEX lsif_indexes_state ON lsif_indexes USING btree (state);

CREATE INDEX lsif_nearest_uploads_links_repository_id_ancestor_commit_bytea ON lsif_nearest_uploads_links USING btree (repository_id, ancestor_commit_bytea);

CREATE INDEX lsif_nearest_uploads_links_repository_id_commit_bytea ON lsif_nearest_uploads_links USING btree (repository_id, commit_bytea);

CREATE INDEX lsif_nearest_uploads_repository_id_commit_bytea ON lsif_nearest_uploads USING btree (repository_id, commit_bytea);

CREATE INDEX lsif_nearest_uploads_uploads ON lsif_nearest_uploads USING gin (uploads);

CREATE INDEX lsif_packages_dump_id ON lsif_packages USING btree (dump_id);

CREATE INDEX lsif_packages_scheme_name_version_dump_id ON lsif_packages USING btree (scheme, name, version, dump_id);

CREATE INDEX lsif_references_dump_id ON lsif_references USING btree (dump_id);

CREATE INDEX lsif_references_scheme_name_version_dump_id ON lsif_references USING btree (scheme, name, version, dump_id);

CREATE INDEX lsif_uploads_associated_index_id ON lsif_uploads USING btree (associated_index_id);

CREATE INDEX lsif_uploads_commit_last_checked_at ON lsif_uploads USING btree (commit_last_checked_at) WHERE (state <> 'deleted'::text);

CREATE INDEX lsif_uploads_committed_at ON lsif_uploads USING btree (committed_at) WHERE (state = 'completed'::text);

CREATE INDEX lsif_uploads_repository_id_commit ON lsif_uploads USING btree (repository_id, commit);

CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root_indexer ON lsif_uploads USING btree (repository_id, commit, root, indexer) WHERE (state = 'completed'::text);

CREATE INDEX lsif_uploads_state ON lsif_uploads USING btree (state);

CREATE INDEX lsif_uploads_uploaded_at ON lsif_uploads USING btree (uploaded_at);

CREATE INDEX lsif_uploads_visible_at_tip_repository_id_upload_id ON lsif_uploads_visible_at_tip USING btree (repository_id, upload_id);

CREATE INDEX notebook_stars_user_id_idx ON notebook_stars USING btree (user_id);

CREATE INDEX notebooks_blocks_tsvector_idx ON notebooks USING gin (blocks_tsvector);

CREATE INDEX notebooks_namespace_org_id_idx ON notebooks USING btree (namespace_org_id);

CREATE INDEX notebooks_namespace_user_id_idx ON notebooks USING btree (namespace_user_id);

CREATE INDEX notebooks_title_trgm_idx ON notebooks USING gin (title gin_trgm_ops);

CREATE INDEX org_invitations_org_id ON org_invitations USING btree (org_id) WHERE (deleted_at IS NULL);

CREATE INDEX org_invitations_recipient_user_id ON org_invitations USING btree (recipient_user_id) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX orgs_name ON orgs USING btree (name) WHERE (deleted_at IS NULL);

CREATE INDEX registry_extension_releases_registry_extension_id ON registry_extension_releases USING btree (registry_extension_id, release_tag, created_at DESC) WHERE (deleted_at IS NULL);

CREATE INDEX registry_extension_releases_registry_extension_id_created_at ON registry_extension_releases USING btree (registry_extension_id, created_at) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX registry_extension_releases_version ON registry_extension_releases USING btree (registry_extension_id, release_version) WHERE (release_version IS NOT NULL);

CREATE UNIQUE INDEX registry_extensions_publisher_name ON registry_extensions USING btree (COALESCE(publisher_user_id, 0), COALESCE(publisher_org_id, 0), name) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX registry_extensions_uuid ON registry_extensions USING btree (uuid);

CREATE INDEX repo_archived ON repo USING btree (archived);

CREATE INDEX repo_blocked_idx ON repo USING btree (((blocked IS NOT NULL)));

CREATE INDEX repo_created_at ON repo USING btree (created_at);

CREATE UNIQUE INDEX repo_external_unique_idx ON repo USING btree (external_service_type, external_service_id, external_id);

CREATE INDEX repo_fork ON repo USING btree (fork);

CREATE INDEX repo_hashed_name_idx ON repo USING btree (sha256((lower((name)::text))::bytea)) WHERE (deleted_at IS NULL);

CREATE INDEX repo_is_not_blocked_idx ON repo USING btree (((blocked IS NULL)));

CREATE INDEX repo_metadata_gin_idx ON repo USING gin (metadata);

CREATE INDEX repo_name_case_sensitive_trgm_idx ON repo USING gin (((name)::text) gin_trgm_ops);

CREATE INDEX repo_name_idx ON repo USING btree (lower((name)::text) COLLATE "C");

CREATE INDEX repo_name_trgm ON repo USING gin (lower((name)::text) gin_trgm_ops);

CREATE INDEX repo_non_deleted_id_name_idx ON repo USING btree (id, name) WHERE (deleted_at IS NULL);

CREATE INDEX repo_private ON repo USING btree (private);

CREATE INDEX repo_stars_desc_id_desc_idx ON repo USING btree (stars DESC NULLS LAST, id DESC) WHERE ((deleted_at IS NULL) AND (blocked IS NULL));

CREATE INDEX repo_stars_idx ON repo USING btree (stars DESC NULLS LAST);

CREATE INDEX repo_uri_idx ON repo USING btree (uri);

CREATE UNIQUE INDEX search_contexts_name_namespace_org_id_unique ON search_contexts USING btree (name, namespace_org_id) WHERE (namespace_org_id IS NOT NULL);

CREATE UNIQUE INDEX search_contexts_name_namespace_user_id_unique ON search_contexts USING btree (name, namespace_user_id) WHERE (namespace_user_id IS NOT NULL);

CREATE UNIQUE INDEX search_contexts_name_without_namespace_unique ON search_contexts USING btree (name) WHERE ((namespace_user_id IS NULL) AND (namespace_org_id IS NULL));

CREATE INDEX search_contexts_query_idx ON search_contexts USING btree (query);

CREATE INDEX security_event_logs_timestamp ON security_event_logs USING btree ("timestamp");

CREATE INDEX settings_global_id ON settings USING btree (id DESC) WHERE ((user_id IS NULL) AND (org_id IS NULL));

CREATE INDEX settings_org_id_idx ON settings USING btree (org_id);

CREATE INDEX settings_user_id_idx ON settings USING btree (user_id);

CREATE UNIQUE INDEX sub_repo_permissions_repo_id_user_id_version_uindex ON sub_repo_permissions USING btree (repo_id, user_id, version);

CREATE INDEX sub_repo_perms_user_id ON sub_repo_permissions USING btree (user_id);

CREATE INDEX user_credentials_credential_idx ON user_credentials USING btree (((encryption_key_id = ANY (ARRAY[''::text, 'previously-migrated'::text]))));

CREATE UNIQUE INDEX user_emails_user_id_is_primary_idx ON user_emails USING btree (user_id, is_primary) WHERE (is_primary = true);

CREATE UNIQUE INDEX user_external_accounts_account ON user_external_accounts USING btree (service_type, service_id, client_id, account_id) WHERE (deleted_at IS NULL);

CREATE INDEX user_external_accounts_user_id ON user_external_accounts USING btree (user_id) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX users_billing_customer_id ON users USING btree (billing_customer_id) WHERE (deleted_at IS NULL);

CREATE INDEX users_created_at_idx ON users USING btree (created_at);

CREATE UNIQUE INDEX users_username ON users USING btree (username) WHERE (deleted_at IS NULL);

CREATE INDEX webhook_logs_external_service_id_idx ON webhook_logs USING btree (external_service_id);

CREATE INDEX webhook_logs_received_at_idx ON webhook_logs USING btree (received_at);

CREATE INDEX webhook_logs_status_code_idx ON webhook_logs USING btree (status_code);

CREATE TRIGGER trig_delete_batch_change_reference_on_changesets AFTER DELETE ON batch_changes FOR EACH ROW EXECUTE FUNCTION delete_batch_change_reference_on_changesets();

CREATE TRIGGER trig_delete_repo_ref_on_external_service_repos AFTER UPDATE OF deleted_at ON repo FOR EACH ROW EXECUTE FUNCTION delete_repo_ref_on_external_service_repos();

CREATE TRIGGER trig_invalidate_session_on_password_change BEFORE UPDATE OF passwd ON users FOR EACH ROW EXECUTE FUNCTION invalidate_session_for_userid_on_password_change();

CREATE TRIGGER trig_soft_delete_user_reference_on_external_service AFTER UPDATE OF deleted_at ON users FOR EACH ROW EXECUTE FUNCTION soft_delete_user_reference_on_external_service();

CREATE TRIGGER versions_insert BEFORE INSERT ON versions FOR EACH ROW EXECUTE FUNCTION versions_insert_row_trigger();

ALTER TABLE ONLY access_tokens
    ADD CONSTRAINT access_tokens_creator_user_id_fkey FOREIGN KEY (creator_user_id) REFERENCES users(id);

ALTER TABLE ONLY access_tokens
    ADD CONSTRAINT access_tokens_subject_user_id_fkey FOREIGN KEY (subject_user_id) REFERENCES users(id);

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_changes_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs(id) DEFERRABLE;

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_changes_initial_applier_id_fkey FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_changes_last_applier_id_fkey FOREIGN KEY (last_applier_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_changes_namespace_org_id_fkey FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_changes_namespace_user_id_fkey FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY batch_spec_execution_cache_entries
    ADD CONSTRAINT batch_spec_execution_cache_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY batch_spec_resolution_jobs
    ADD CONSTRAINT batch_spec_resolution_jobs_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY batch_spec_workspace_execution_jobs
    ADD CONSTRAINT batch_spec_workspace_execution_job_batch_spec_workspace_id_fkey FOREIGN KEY (batch_spec_workspace_id) REFERENCES batch_spec_workspaces(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY batch_spec_workspace_execution_jobs
    ADD CONSTRAINT batch_spec_workspace_execution_jobs_access_token_id_fkey FOREIGN KEY (access_token_id) REFERENCES access_tokens(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY batch_spec_workspaces
    ADD CONSTRAINT batch_spec_workspaces_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY batch_spec_workspaces
    ADD CONSTRAINT batch_spec_workspaces_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) DEFERRABLE;

ALTER TABLE ONLY batch_specs
    ADD CONSTRAINT batch_specs_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY changeset_events
    ADD CONSTRAINT changeset_events_changeset_id_fkey FOREIGN KEY (changeset_id) REFERENCES changesets(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY changeset_jobs
    ADD CONSTRAINT changeset_jobs_batch_change_id_fkey FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY changeset_jobs
    ADD CONSTRAINT changeset_jobs_changeset_id_fkey FOREIGN KEY (changeset_id) REFERENCES changesets(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY changeset_jobs
    ADD CONSTRAINT changeset_jobs_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY changeset_specs
    ADD CONSTRAINT changeset_specs_batch_spec_id_fkey FOREIGN KEY (batch_spec_id) REFERENCES batch_specs(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY changeset_specs
    ADD CONSTRAINT changeset_specs_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) DEFERRABLE;

ALTER TABLE ONLY changeset_specs
    ADD CONSTRAINT changeset_specs_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY changesets
    ADD CONSTRAINT changesets_changeset_spec_id_fkey FOREIGN KEY (current_spec_id) REFERENCES changeset_specs(id) DEFERRABLE;

ALTER TABLE ONLY changesets
    ADD CONSTRAINT changesets_owned_by_batch_spec_id_fkey FOREIGN KEY (owned_by_batch_change_id) REFERENCES batch_changes(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY changesets
    ADD CONSTRAINT changesets_previous_spec_id_fkey FOREIGN KEY (previous_spec_id) REFERENCES changeset_specs(id) DEFERRABLE;

ALTER TABLE ONLY changesets
    ADD CONSTRAINT changesets_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY cm_action_jobs
    ADD CONSTRAINT cm_action_jobs_email_fk FOREIGN KEY (email) REFERENCES cm_emails(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_action_jobs
    ADD CONSTRAINT cm_action_jobs_slack_webhook_fkey FOREIGN KEY (slack_webhook) REFERENCES cm_slack_webhooks(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_action_jobs
    ADD CONSTRAINT cm_action_jobs_trigger_event_fk FOREIGN KEY (trigger_event) REFERENCES cm_trigger_jobs(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_action_jobs
    ADD CONSTRAINT cm_action_jobs_webhook_fkey FOREIGN KEY (webhook) REFERENCES cm_webhooks(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_emails
    ADD CONSTRAINT cm_emails_changed_by_fk FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_emails
    ADD CONSTRAINT cm_emails_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_emails
    ADD CONSTRAINT cm_emails_monitor FOREIGN KEY (monitor) REFERENCES cm_monitors(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_last_searched
    ADD CONSTRAINT cm_last_searched_monitor_id_fkey FOREIGN KEY (monitor_id) REFERENCES cm_monitors(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_monitors
    ADD CONSTRAINT cm_monitors_changed_by_fk FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_monitors
    ADD CONSTRAINT cm_monitors_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_monitors
    ADD CONSTRAINT cm_monitors_org_id_fk FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_monitors
    ADD CONSTRAINT cm_monitors_user_id_fk FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_recipients
    ADD CONSTRAINT cm_recipients_emails FOREIGN KEY (email) REFERENCES cm_emails(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_recipients
    ADD CONSTRAINT cm_recipients_org_id_fk FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_recipients
    ADD CONSTRAINT cm_recipients_user_id_fk FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_slack_webhooks
    ADD CONSTRAINT cm_slack_webhooks_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_slack_webhooks
    ADD CONSTRAINT cm_slack_webhooks_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_slack_webhooks
    ADD CONSTRAINT cm_slack_webhooks_monitor_fkey FOREIGN KEY (monitor) REFERENCES cm_monitors(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_trigger_jobs
    ADD CONSTRAINT cm_trigger_jobs_query_fk FOREIGN KEY (query) REFERENCES cm_queries(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_queries
    ADD CONSTRAINT cm_triggers_changed_by_fk FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_queries
    ADD CONSTRAINT cm_triggers_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_queries
    ADD CONSTRAINT cm_triggers_monitor FOREIGN KEY (monitor) REFERENCES cm_monitors(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_webhooks
    ADD CONSTRAINT cm_webhooks_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_webhooks
    ADD CONSTRAINT cm_webhooks_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cm_webhooks
    ADD CONSTRAINT cm_webhooks_monitor_fkey FOREIGN KEY (monitor) REFERENCES cm_monitors(id) ON DELETE CASCADE;

ALTER TABLE ONLY discussion_comments
    ADD CONSTRAINT discussion_comments_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY discussion_comments
    ADD CONSTRAINT discussion_comments_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE;

ALTER TABLE ONLY discussion_mail_reply_tokens
    ADD CONSTRAINT discussion_mail_reply_tokens_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE;

ALTER TABLE ONLY discussion_mail_reply_tokens
    ADD CONSTRAINT discussion_mail_reply_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY discussion_threads
    ADD CONSTRAINT discussion_threads_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY discussion_threads
    ADD CONSTRAINT discussion_threads_target_repo_id_fk FOREIGN KEY (target_repo_id) REFERENCES discussion_threads_target_repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY discussion_threads_target_repo
    ADD CONSTRAINT discussion_threads_target_repo_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY discussion_threads_target_repo
    ADD CONSTRAINT discussion_threads_target_repo_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE;

ALTER TABLE ONLY external_service_repos
    ADD CONSTRAINT external_service_repos_external_service_id_fkey FOREIGN KEY (external_service_id) REFERENCES external_services(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY external_service_repos
    ADD CONSTRAINT external_service_repos_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE;

ALTER TABLE ONLY external_service_repos
    ADD CONSTRAINT external_service_repos_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY external_service_repos
    ADD CONSTRAINT external_service_repos_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY external_service_sync_jobs
    ADD CONSTRAINT external_services_id_fk FOREIGN KEY (external_service_id) REFERENCES external_services(id) ON DELETE CASCADE;

ALTER TABLE ONLY external_services
    ADD CONSTRAINT external_services_namepspace_user_id_fkey FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY external_services
    ADD CONSTRAINT external_services_namespace_org_id_fkey FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY feature_flag_overrides
    ADD CONSTRAINT feature_flag_overrides_flag_name_fkey FOREIGN KEY (flag_name) REFERENCES feature_flags(flag_name) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY feature_flag_overrides
    ADD CONSTRAINT feature_flag_overrides_namespace_org_id_fkey FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE CASCADE;

ALTER TABLE ONLY feature_flag_overrides
    ADD CONSTRAINT feature_flag_overrides_namespace_user_id_fkey FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY gitserver_repos
    ADD CONSTRAINT gitserver_repos_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY insights_query_runner_jobs_dependencies
    ADD CONSTRAINT insights_query_runner_jobs_dependencies_fk_job_id FOREIGN KEY (job_id) REFERENCES insights_query_runner_jobs(id) ON DELETE CASCADE;

ALTER TABLE ONLY lsif_dependency_syncing_jobs
    ADD CONSTRAINT lsif_dependency_indexing_jobs_upload_id_fkey FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;

ALTER TABLE ONLY lsif_dependency_indexing_jobs
    ADD CONSTRAINT lsif_dependency_indexing_jobs_upload_id_fkey1 FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;

ALTER TABLE ONLY lsif_index_configuration
    ADD CONSTRAINT lsif_index_configuration_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY lsif_packages
    ADD CONSTRAINT lsif_packages_dump_id_fkey FOREIGN KEY (dump_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;

ALTER TABLE ONLY lsif_references
    ADD CONSTRAINT lsif_references_dump_id_fkey FOREIGN KEY (dump_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;

ALTER TABLE ONLY lsif_retention_configuration
    ADD CONSTRAINT lsif_retention_configuration_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY names
    ADD CONSTRAINT names_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY names
    ADD CONSTRAINT names_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY notebook_stars
    ADD CONSTRAINT notebook_stars_notebook_id_fkey FOREIGN KEY (notebook_id) REFERENCES notebooks(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY notebook_stars
    ADD CONSTRAINT notebook_stars_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY notebooks
    ADD CONSTRAINT notebooks_creator_user_id_fkey FOREIGN KEY (creator_user_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY notebooks
    ADD CONSTRAINT notebooks_namespace_org_id_fkey FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY notebooks
    ADD CONSTRAINT notebooks_namespace_user_id_fkey FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY notebooks
    ADD CONSTRAINT notebooks_updater_user_id_fkey FOREIGN KEY (updater_user_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

ALTER TABLE ONLY org_invitations
    ADD CONSTRAINT org_invitations_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs(id);

ALTER TABLE ONLY org_invitations
    ADD CONSTRAINT org_invitations_recipient_user_id_fkey FOREIGN KEY (recipient_user_id) REFERENCES users(id);

ALTER TABLE ONLY org_invitations
    ADD CONSTRAINT org_invitations_sender_user_id_fkey FOREIGN KEY (sender_user_id) REFERENCES users(id);

ALTER TABLE ONLY org_members
    ADD CONSTRAINT org_members_references_orgs FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT;

ALTER TABLE ONLY org_members
    ADD CONSTRAINT org_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY org_stats
    ADD CONSTRAINT org_stats_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY out_of_band_migrations_errors
    ADD CONSTRAINT out_of_band_migrations_errors_migration_id_fkey FOREIGN KEY (migration_id) REFERENCES out_of_band_migrations(id) ON DELETE CASCADE;

ALTER TABLE ONLY product_licenses
    ADD CONSTRAINT product_licenses_product_subscription_id_fkey FOREIGN KEY (product_subscription_id) REFERENCES product_subscriptions(id);

ALTER TABLE ONLY product_subscriptions
    ADD CONSTRAINT product_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY registry_extension_releases
    ADD CONSTRAINT registry_extension_releases_creator_user_id_fkey FOREIGN KEY (creator_user_id) REFERENCES users(id);

ALTER TABLE ONLY registry_extension_releases
    ADD CONSTRAINT registry_extension_releases_registry_extension_id_fkey FOREIGN KEY (registry_extension_id) REFERENCES registry_extensions(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE ONLY registry_extensions
    ADD CONSTRAINT registry_extensions_publisher_org_id_fkey FOREIGN KEY (publisher_org_id) REFERENCES orgs(id);

ALTER TABLE ONLY registry_extensions
    ADD CONSTRAINT registry_extensions_publisher_user_id_fkey FOREIGN KEY (publisher_user_id) REFERENCES users(id);

ALTER TABLE ONLY saved_searches
    ADD CONSTRAINT saved_searches_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs(id);

ALTER TABLE ONLY saved_searches
    ADD CONSTRAINT saved_searches_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY search_context_repos
    ADD CONSTRAINT search_context_repos_repo_id_fk FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY search_context_repos
    ADD CONSTRAINT search_context_repos_search_context_id_fk FOREIGN KEY (search_context_id) REFERENCES search_contexts(id) ON DELETE CASCADE;

ALTER TABLE ONLY search_contexts
    ADD CONSTRAINT search_contexts_namespace_org_id_fk FOREIGN KEY (namespace_org_id) REFERENCES orgs(id) ON DELETE CASCADE;

ALTER TABLE ONLY search_contexts
    ADD CONSTRAINT search_contexts_namespace_user_id_fk FOREIGN KEY (namespace_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY settings
    ADD CONSTRAINT settings_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY settings
    ADD CONSTRAINT settings_references_orgs FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT;

ALTER TABLE ONLY settings
    ADD CONSTRAINT settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY sub_repo_permissions
    ADD CONSTRAINT sub_repo_permissions_repo_id_fk FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY sub_repo_permissions
    ADD CONSTRAINT sub_repo_permissions_users_id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY survey_responses
    ADD CONSTRAINT survey_responses_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY temporary_settings
    ADD CONSTRAINT temporary_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY user_credentials
    ADD CONSTRAINT user_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY user_emails
    ADD CONSTRAINT user_emails_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY user_external_accounts
    ADD CONSTRAINT user_external_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY user_public_repos
    ADD CONSTRAINT user_public_repos_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;

ALTER TABLE ONLY user_public_repos
    ADD CONSTRAINT user_public_repos_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY webhook_logs
    ADD CONSTRAINT webhook_logs_external_service_id_fkey FOREIGN KEY (external_service_id) REFERENCES external_services(id) ON UPDATE CASCADE ON DELETE CASCADE;

INSERT INTO lsif_configuration_policies VALUES (1, NULL, 'Default tip-of-branch retention policy', 'GIT_TREE', '*', true, 2016, false, false, 0, false, true, NULL, NULL);
INSERT INTO lsif_configuration_policies VALUES (2, NULL, 'Default tag retention policy', 'GIT_TAG', '*', true, 8064, false, false, 0, false, true, NULL, NULL);
INSERT INTO lsif_configuration_policies VALUES (3, NULL, 'Default commit retention policy', 'GIT_TREE', '*', true, 168, true, false, 0, false, true, NULL, NULL);

SELECT pg_catalog.setval('lsif_configuration_policies_id_seq', 3, true);
