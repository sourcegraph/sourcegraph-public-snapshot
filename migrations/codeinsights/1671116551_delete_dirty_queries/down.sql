CREATE TABLE insight_dirty_queries (
    id integer PRIMARY KEY,
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

ALTER TABLE ONLY insight_dirty_queries ALTER COLUMN id SET DEFAULT nextval('insight_dirty_queries_id_seq'::regclass);

ALTER TABLE ONLY insight_dirty_queries
    ADD CONSTRAINT insight_dirty_queries_insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series(id) ON DELETE CASCADE;

CREATE INDEX insight_dirty_queries_insight_series_id_fk_idx ON insight_dirty_queries USING btree (insight_series_id);