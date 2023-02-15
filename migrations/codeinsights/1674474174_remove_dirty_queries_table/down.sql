CREATE TABLE IF NOT EXISTS insight_dirty_queries (
    id integer PRIMARY KEY,
    insight_series_id integer,
    query text NOT NULL,
    dirty_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    reason text NOT NULL,
    for_time timestamp without time zone NOT NULL
);


CREATE SEQUENCE IF NOT EXISTS insight_dirty_queries_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE IF EXISTS insight_dirty_queries_id_seq OWNED BY insight_dirty_queries.id;

ALTER TABLE IF EXISTS insight_dirty_queries ALTER COLUMN id SET DEFAULT nextval('insight_dirty_queries_id_seq'::regclass);

ALTER TABLE IF EXISTS insight_dirty_queries DROP CONSTRAINT IF EXISTS insight_dirty_queries_insight_series_id_fkey;
ALTER TABLE IF EXISTS insight_dirty_queries
    ADD CONSTRAINT insight_dirty_queries_insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS insight_dirty_queries_insight_series_id_fk_idx ON insight_dirty_queries USING btree (insight_series_id);