CREATE TYPE presentation_type_enum AS ENUM (
    'LINE',
    'PIE'
);

CREATE TYPE time_unit AS ENUM (
    'HOUR',
    'DAY',
    'WEEK',
    'MONTH',
    'YEAR'
);

CREATE TABLE commit_index (
    committed_at timestamp with time zone NOT NULL,
    repo_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    indexed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    debug_field text
);

CREATE TABLE commit_index_metadata (
    repo_id integer NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    last_indexed_at timestamp with time zone DEFAULT '1900-01-01 00:00:00+00'::timestamp with time zone NOT NULL
);

CREATE TABLE dashboard (
    id integer NOT NULL,
    title text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    created_by_user_id integer,
    last_updated_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone,
    save boolean DEFAULT false NOT NULL
);

COMMENT ON TABLE dashboard IS 'Metadata for dashboards of insights';

COMMENT ON COLUMN dashboard.title IS 'Title of the dashboard';

COMMENT ON COLUMN dashboard.created_at IS 'Timestamp the dashboard was initially created.';

COMMENT ON COLUMN dashboard.created_by_user_id IS 'User that created the dashboard, if available.';

COMMENT ON COLUMN dashboard.last_updated_at IS 'Time the dashboard was last updated, either metadata or insights.';

COMMENT ON COLUMN dashboard.deleted_at IS 'Set to the time the dashboard was soft deleted.';

COMMENT ON COLUMN dashboard.save IS 'TEMPORARY Do not delete this dashboard when migrating settings.';

CREATE TABLE dashboard_grants (
    id integer NOT NULL,
    dashboard_id integer NOT NULL,
    user_id integer,
    org_id integer,
    global boolean
);

COMMENT ON TABLE dashboard_grants IS 'Permission grants for dashboards. Each row should represent a unique principal (user, org, etc).';

COMMENT ON COLUMN dashboard_grants.user_id IS 'User ID that that receives this grant.';

COMMENT ON COLUMN dashboard_grants.org_id IS 'Org ID that that receives this grant.';

COMMENT ON COLUMN dashboard_grants.global IS 'Grant that does not belong to any specific principal and is granted to all users.';

CREATE SEQUENCE dashboard_grants_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE dashboard_grants_id_seq OWNED BY dashboard_grants.id;

CREATE SEQUENCE dashboard_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE dashboard_id_seq OWNED BY dashboard.id;

CREATE TABLE dashboard_insight_view (
    id integer NOT NULL,
    dashboard_id integer NOT NULL,
    insight_view_id integer NOT NULL
);

CREATE SEQUENCE dashboard_insight_view_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE dashboard_insight_view_id_seq OWNED BY dashboard_insight_view.id;

CREATE TABLE insight_dirty_queries (
    id integer NOT NULL,
    insight_series_id integer,
    query text NOT NULL,
    dirty_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    reason text NOT NULL,
    for_time timestamp without time zone NOT NULL
);

COMMENT ON TABLE insight_dirty_queries IS 'Stores queries that were unsuccessful or otherwise flagged as incomplete or incorrect.';

COMMENT ON COLUMN insight_dirty_queries.query IS 'Sourcegraph query string that was executed.';

COMMENT ON COLUMN insight_dirty_queries.dirty_at IS 'Timestamp when this query was marked dirty.';

COMMENT ON COLUMN insight_dirty_queries.reason IS 'Human readable string indicating the reason the query was marked dirty.';

COMMENT ON COLUMN insight_dirty_queries.for_time IS 'Timestamp for which the original data point was recorded or intended to be recorded.';

CREATE SEQUENCE insight_dirty_queries_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insight_dirty_queries_id_seq OWNED BY insight_dirty_queries.id;

CREATE TABLE insight_series (
    id integer NOT NULL,
    series_id text NOT NULL,
    query text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    oldest_historical_at timestamp without time zone DEFAULT (CURRENT_TIMESTAMP - '1 year'::interval) NOT NULL,
    last_recorded_at timestamp without time zone DEFAULT (CURRENT_TIMESTAMP - '10 years'::interval) NOT NULL,
    next_recording_after timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    backfill_queued_at timestamp without time zone,
    last_snapshot_at timestamp without time zone DEFAULT (CURRENT_TIMESTAMP - '10 years'::interval),
    next_snapshot_after timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    repositories text[],
    sample_interval_unit time_unit DEFAULT 'MONTH'::time_unit NOT NULL,
    sample_interval_value integer DEFAULT 1 NOT NULL,
    generated_from_capture_groups boolean DEFAULT false NOT NULL,
    generation_method text NOT NULL,
    just_in_time boolean DEFAULT false NOT NULL
);

COMMENT ON TABLE insight_series IS 'Data series that comprise code insights.';

COMMENT ON COLUMN insight_series.id IS 'Primary key ID of this series';

COMMENT ON COLUMN insight_series.series_id IS 'Timestamp that this series completed a full repository iteration for backfill. This flag has limited semantic value, and only means it tried to queue up queries for each repository. It does not guarantee success on those queries.';

COMMENT ON COLUMN insight_series.query IS 'Query string that generates this series';

COMMENT ON COLUMN insight_series.created_at IS 'Timestamp when this series was created';

COMMENT ON COLUMN insight_series.oldest_historical_at IS 'Timestamp representing the oldest point of which this series is backfilled.';

COMMENT ON COLUMN insight_series.last_recorded_at IS 'Timestamp when this series was last recorded (non-historical).';

COMMENT ON COLUMN insight_series.next_recording_after IS 'Timestamp when this series should next record (non-historical).';

COMMENT ON COLUMN insight_series.deleted_at IS 'Timestamp of a soft-delete of this row.';

COMMENT ON COLUMN insight_series.generation_method IS 'Specifies the execution method for how this series is generated. This helps the system understand how to generate the time series data.';

COMMENT ON COLUMN insight_series.just_in_time IS 'Specifies if the series should be resolved just in time at query time, or recorded in background processing.';

CREATE SEQUENCE insight_series_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insight_series_id_seq OWNED BY insight_series.id;

CREATE TABLE insight_view (
    id integer NOT NULL,
    title text,
    description text,
    unique_id text NOT NULL,
    default_filter_include_repo_regex text,
    default_filter_exclude_repo_regex text,
    other_threshold double precision,
    presentation_type presentation_type_enum DEFAULT 'LINE'::presentation_type_enum NOT NULL,
    is_frozen boolean DEFAULT false NOT NULL
);

COMMENT ON TABLE insight_view IS 'Views for insight data series. An insight view is an abstraction on top of an insight data series that allows for lightweight modifications to filters or metadata without regenerating the underlying series.';

COMMENT ON COLUMN insight_view.id IS 'Primary key ID for this view';

COMMENT ON COLUMN insight_view.title IS 'Title of the view. This may render in a chart depending on the view type.';

COMMENT ON COLUMN insight_view.description IS 'Description of the view. This may render in a chart depending on the view type.';

COMMENT ON COLUMN insight_view.unique_id IS 'Globally unique identifier for this view that is externally referencable.';

COMMENT ON COLUMN insight_view.other_threshold IS 'Percent threshold for grouping series under "other"';

COMMENT ON COLUMN insight_view.presentation_type IS 'The basic presentation type for the insight view. (e.g Line, Pie, etc.)';

CREATE TABLE insight_view_grants (
    id integer NOT NULL,
    insight_view_id integer NOT NULL,
    user_id integer,
    org_id integer,
    global boolean
);

COMMENT ON TABLE insight_view_grants IS 'Permission grants for insight views. Each row should represent a unique principal (user, org, etc).';

COMMENT ON COLUMN insight_view_grants.user_id IS 'User ID that that receives this grant.';

COMMENT ON COLUMN insight_view_grants.org_id IS 'Org ID that that receives this grant.';

COMMENT ON COLUMN insight_view_grants.global IS 'Grant that does not belong to any specific principal and is granted to all users.';

CREATE SEQUENCE insight_view_grants_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insight_view_grants_id_seq OWNED BY insight_view_grants.id;

CREATE SEQUENCE insight_view_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE insight_view_id_seq OWNED BY insight_view.id;

CREATE TABLE insight_view_series (
    insight_view_id integer NOT NULL,
    insight_series_id integer NOT NULL,
    label text,
    stroke text
);

COMMENT ON TABLE insight_view_series IS 'Join table to correlate data series with insight views';

COMMENT ON COLUMN insight_view_series.insight_view_id IS 'Foreign key to insight view.';

COMMENT ON COLUMN insight_view_series.insight_series_id IS 'Foreign key to insight data series.';

COMMENT ON COLUMN insight_view_series.label IS 'Label text for this data series. This may render in a chart depending on the view type.';

COMMENT ON COLUMN insight_view_series.stroke IS 'Stroke color metadata for this data series. This may render in a chart depending on the view type.';

CREATE TABLE metadata (
    id bigint NOT NULL,
    metadata jsonb NOT NULL
);

COMMENT ON TABLE metadata IS 'Records arbitrary metadata about events. Stored in a separate table as it is often repeated for multiple events.';

COMMENT ON COLUMN metadata.id IS 'The metadata ID.';

COMMENT ON COLUMN metadata.metadata IS 'Metadata about some event, this can be any arbitrary JSON emtadata which will be returned when querying events, and can be filtered on and grouped using jsonb operators ?, ?&, ?|, and @>. This should be small data only.';

CREATE SEQUENCE metadata_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE metadata_id_seq OWNED BY metadata.id;

CREATE TABLE repo_names (
    id bigint NOT NULL,
    name citext NOT NULL,
    CONSTRAINT check_name_nonempty CHECK ((name OPERATOR(<>) ''::citext))
);

COMMENT ON TABLE repo_names IS 'Records repository names, both historical and present, using a unique repository _name_ ID (unrelated to the repository ID.)';

COMMENT ON COLUMN repo_names.id IS 'The repository _name_ ID.';

COMMENT ON COLUMN repo_names.name IS 'The repository name string, with unique constraint for table entry deduplication and trigram index for e.g. regex filtering.';

CREATE SEQUENCE repo_names_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE repo_names_id_seq OWNED BY repo_names.id;

CREATE TABLE series_points (
    series_id text NOT NULL,
    "time" timestamp with time zone NOT NULL,
    value double precision NOT NULL,
    metadata_id integer,
    repo_id integer,
    repo_name_id integer,
    original_repo_name_id integer,
    capture text,
    CONSTRAINT check_repo_fields_specifity CHECK ((((repo_id IS NULL) AND (repo_name_id IS NULL) AND (original_repo_name_id IS NULL)) OR ((repo_id IS NOT NULL) AND (repo_name_id IS NOT NULL) AND (original_repo_name_id IS NOT NULL))))
);

COMMENT ON TABLE series_points IS 'Records events over time associated with a repository (or none, i.e. globally) where a single numerical value is going arbitrarily up and down.  Repository association is based on both repository ID and name. The ID can be used to refer toa specific repository, or lookup the current name of a repository after it has been e.g. renamed. The name can be used to refer to the name of the repository at the time of the events creation, for example to trace the change in a gauge back to a repository being renamed.';

COMMENT ON COLUMN series_points.series_id IS 'A unique identifier for the series of data being recorded. This is not an ID from another table, but rather just a unique identifier.';

COMMENT ON COLUMN series_points."time" IS 'The timestamp of the recorded event.';

COMMENT ON COLUMN series_points.value IS 'The floating point value at the time of the event.';

COMMENT ON COLUMN series_points.metadata_id IS 'Associated metadata for this event, if any.';

COMMENT ON COLUMN series_points.repo_id IS 'The repository ID (from the main application DB) at the time the event was created. Note that the repository may no longer exist / be valid at query time, however.';

COMMENT ON COLUMN series_points.repo_name_id IS 'The most recently known name for the repository, updated periodically to account for e.g. repository renames. If the repository was deleted, this is still the most recently known name.  null if the event was not for a single repository (i.e. a global gauge).';

COMMENT ON COLUMN series_points.original_repo_name_id IS 'The repository name as it was known at the time the event was created. It may have been renamed since.';

CREATE TABLE series_points_snapshots (
    series_id text NOT NULL,
    "time" timestamp with time zone NOT NULL,
    value double precision NOT NULL,
    metadata_id integer,
    repo_id integer,
    repo_name_id integer,
    original_repo_name_id integer,
    capture text,
    CONSTRAINT check_repo_fields_specifity CHECK ((((repo_id IS NULL) AND (repo_name_id IS NULL) AND (original_repo_name_id IS NULL)) OR ((repo_id IS NOT NULL) AND (repo_name_id IS NOT NULL) AND (original_repo_name_id IS NOT NULL))))
);

COMMENT ON TABLE series_points_snapshots IS 'Stores ephemeral snapshot data of insight recordings.';

ALTER TABLE ONLY dashboard ALTER COLUMN id SET DEFAULT nextval('dashboard_id_seq'::regclass);

ALTER TABLE ONLY dashboard_grants ALTER COLUMN id SET DEFAULT nextval('dashboard_grants_id_seq'::regclass);

ALTER TABLE ONLY dashboard_insight_view ALTER COLUMN id SET DEFAULT nextval('dashboard_insight_view_id_seq'::regclass);

ALTER TABLE ONLY insight_dirty_queries ALTER COLUMN id SET DEFAULT nextval('insight_dirty_queries_id_seq'::regclass);

ALTER TABLE ONLY insight_series ALTER COLUMN id SET DEFAULT nextval('insight_series_id_seq'::regclass);

ALTER TABLE ONLY insight_view ALTER COLUMN id SET DEFAULT nextval('insight_view_id_seq'::regclass);

ALTER TABLE ONLY insight_view_grants ALTER COLUMN id SET DEFAULT nextval('insight_view_grants_id_seq'::regclass);

ALTER TABLE ONLY metadata ALTER COLUMN id SET DEFAULT nextval('metadata_id_seq'::regclass);

ALTER TABLE ONLY repo_names ALTER COLUMN id SET DEFAULT nextval('repo_names_id_seq'::regclass);

ALTER TABLE ONLY commit_index_metadata
    ADD CONSTRAINT commit_index_metadata_pkey PRIMARY KEY (repo_id);

ALTER TABLE ONLY commit_index
    ADD CONSTRAINT commit_index_pkey PRIMARY KEY (committed_at, repo_id, commit_bytea);

ALTER TABLE ONLY dashboard_grants
    ADD CONSTRAINT dashboard_grants_pk PRIMARY KEY (id);

ALTER TABLE ONLY dashboard_insight_view
    ADD CONSTRAINT dashboard_insight_view_pk PRIMARY KEY (id);

ALTER TABLE ONLY dashboard
    ADD CONSTRAINT dashboard_pk PRIMARY KEY (id);

ALTER TABLE ONLY insight_dirty_queries
    ADD CONSTRAINT insight_dirty_queries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY insight_series
    ADD CONSTRAINT insight_series_pkey PRIMARY KEY (id);

ALTER TABLE ONLY insight_view_grants
    ADD CONSTRAINT insight_view_grants_pk PRIMARY KEY (id);

ALTER TABLE ONLY insight_view
    ADD CONSTRAINT insight_view_pkey PRIMARY KEY (id);

ALTER TABLE ONLY insight_view_series
    ADD CONSTRAINT insight_view_series_pkey PRIMARY KEY (insight_view_id, insight_series_id);

ALTER TABLE ONLY metadata
    ADD CONSTRAINT metadata_pkey PRIMARY KEY (id);

ALTER TABLE ONLY repo_names
    ADD CONSTRAINT repo_names_pkey PRIMARY KEY (id);

ALTER TABLE ONLY dashboard_insight_view
    ADD CONSTRAINT unique_dashboard_id_insight_view_id UNIQUE (dashboard_id, insight_view_id);

CREATE INDEX commit_index_repo_id_idx ON commit_index USING btree (repo_id, committed_at);

CREATE INDEX dashboard_grants_dashboard_id_index ON dashboard_grants USING btree (dashboard_id);

CREATE INDEX dashboard_grants_global_idx ON dashboard_grants USING btree (global) WHERE (global IS TRUE);

CREATE INDEX dashboard_grants_org_id_idx ON dashboard_grants USING btree (org_id);

CREATE INDEX dashboard_grants_user_id_idx ON dashboard_grants USING btree (user_id);

CREATE INDEX dashboard_insight_view_dashboard_id_fk_idx ON dashboard_insight_view USING btree (dashboard_id);

CREATE INDEX dashboard_insight_view_insight_view_id_fk_idx ON dashboard_insight_view USING btree (insight_view_id);

CREATE INDEX insight_dirty_queries_insight_series_id_fk_idx ON insight_dirty_queries USING btree (insight_series_id);

CREATE INDEX insight_series_deleted_at_idx ON insight_series USING btree (deleted_at);

CREATE INDEX insight_series_next_recording_after_idx ON insight_series USING btree (next_recording_after);

CREATE UNIQUE INDEX insight_series_series_id_unique_idx ON insight_series USING btree (series_id);

CREATE INDEX insight_view_grants_global_idx ON insight_view_grants USING btree (global) WHERE (global IS TRUE);

CREATE INDEX insight_view_grants_insight_view_id_index ON insight_view_grants USING btree (insight_view_id);

CREATE INDEX insight_view_grants_org_id_idx ON insight_view_grants USING btree (org_id);

CREATE INDEX insight_view_grants_user_id_idx ON insight_view_grants USING btree (user_id);

CREATE UNIQUE INDEX insight_view_unique_id_unique_idx ON insight_view USING btree (unique_id);

CREATE INDEX metadata_metadata_gin ON metadata USING gin (metadata);

CREATE UNIQUE INDEX metadata_metadata_unique_idx ON metadata USING btree (metadata);

CREATE INDEX repo_names_name_trgm ON repo_names USING gin (lower((name)::text) gin_trgm_ops);

CREATE UNIQUE INDEX repo_names_name_unique_idx ON repo_names USING btree (name);

CREATE INDEX series_points_original_repo_name_id_btree ON series_points USING btree (original_repo_name_id);

CREATE INDEX series_points_repo_id_btree ON series_points USING btree (repo_id);

CREATE INDEX series_points_repo_name_id_btree ON series_points USING btree (repo_name_id);

CREATE INDEX series_points_series_id_btree ON series_points USING btree (series_id);

CREATE INDEX series_points_series_id_repo_id_time_idx ON series_points USING btree (series_id, repo_id, "time");

CREATE INDEX series_points_snapshots_original_repo_name_id_idx ON series_points_snapshots USING btree (original_repo_name_id);

CREATE INDEX series_points_snapshots_repo_id_idx ON series_points_snapshots USING btree (repo_id);

CREATE INDEX series_points_snapshots_repo_name_id_idx ON series_points_snapshots USING btree (repo_name_id);

CREATE INDEX series_points_snapshots_series_id_idx ON series_points_snapshots USING btree (series_id);

CREATE INDEX series_points_snapshots_series_id_repo_id_time_idx ON series_points_snapshots USING btree (series_id, repo_id, "time");

ALTER TABLE ONLY dashboard_grants
    ADD CONSTRAINT dashboard_grants_dashboard_id_fk FOREIGN KEY (dashboard_id) REFERENCES dashboard(id) ON DELETE CASCADE;

ALTER TABLE ONLY dashboard_insight_view
    ADD CONSTRAINT dashboard_insight_view_dashboard_id_fk FOREIGN KEY (dashboard_id) REFERENCES dashboard(id) ON DELETE CASCADE;

ALTER TABLE ONLY dashboard_insight_view
    ADD CONSTRAINT dashboard_insight_view_insight_view_id_fk FOREIGN KEY (insight_view_id) REFERENCES insight_view(id) ON DELETE CASCADE;

ALTER TABLE ONLY insight_dirty_queries
    ADD CONSTRAINT insight_dirty_queries_insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series(id) ON DELETE CASCADE;

ALTER TABLE ONLY insight_view_grants
    ADD CONSTRAINT insight_view_grants_insight_view_id_fk FOREIGN KEY (insight_view_id) REFERENCES insight_view(id) ON DELETE CASCADE;

ALTER TABLE ONLY insight_view_series
    ADD CONSTRAINT insight_view_series_insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series(id);

ALTER TABLE ONLY insight_view_series
    ADD CONSTRAINT insight_view_series_insight_view_id_fkey FOREIGN KEY (insight_view_id) REFERENCES insight_view(id) ON DELETE CASCADE;

ALTER TABLE ONLY series_points
    ADD CONSTRAINT series_points_metadata_id_fkey FOREIGN KEY (metadata_id) REFERENCES metadata(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY series_points
    ADD CONSTRAINT series_points_original_repo_name_id_fkey FOREIGN KEY (original_repo_name_id) REFERENCES repo_names(id) ON DELETE CASCADE DEFERRABLE;

ALTER TABLE ONLY series_points
    ADD CONSTRAINT series_points_repo_name_id_fkey FOREIGN KEY (repo_name_id) REFERENCES repo_names(id) ON DELETE CASCADE DEFERRABLE;