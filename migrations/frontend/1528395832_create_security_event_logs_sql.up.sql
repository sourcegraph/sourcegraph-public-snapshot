BEGIN;

CREATE TABLE IF NOT EXISTS security_event_logs (
    id                BIGSERIAL PRIMARY KEY,
    name              TEXT      NOT NULL,
    url               TEXT      NOT NULL,
    user_id           INTEGER   NOT NULL,
    anonymous_user_id TEXT      NOT NULL,
    source            TEXT      NOT NULL,
    argument          JSONB     NOT NULL,
    version           TEXT      NOT NULL,
    "timestamp"       TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT security_event_logs_check_has_user          CHECK ((((user_id = 0) AND (anonymous_user_id <> ''::text)) OR ((user_id <> 0) AND (anonymous_user_id = ''::text)) OR ((user_id <> 0) AND (anonymous_user_id <> ''::text)))),
    CONSTRAINT security_event_logs_check_name_not_empty    CHECK ((name <> ''::text)),
    CONSTRAINT security_event_logs_check_source_not_empty  CHECK ((source <> ''::text)),
    CONSTRAINT security_event_logs_check_version_not_empty CHECK ((version <> ''::text))
);

CREATE INDEX security_event_logs_user_id           ON security_event_logs USING btree (user_id);
CREATE INDEX security_event_logs_anonymous_user_id ON security_event_logs USING btree (anonymous_user_id);
CREATE INDEX security_event_logs_name              ON security_event_logs USING btree (name);
CREATE INDEX security_event_logs_source            ON security_event_logs USING btree (source);
CREATE INDEX security_event_logs_timestamp         ON security_event_logs USING btree ("timestamp");
CREATE INDEX security_event_logs_timestamp_at_utc  ON security_event_logs USING btree (date(timezone('UTC'::text, "timestamp")));

COMMENT ON TABLE  security_event_logs                   IS 'Contains security-relevant events with a long time horizon for storage.';
COMMENT ON COLUMN security_event_logs.name              IS 'The event name as a CAPITALIZED_SNAKE_CASE string.';
COMMENT ON COLUMN security_event_logs.url               IS 'The URL within the Sourcegraph app which generated the event.';
COMMENT ON COLUMN security_event_logs.user_id           IS 'The ID of the actor associated with the event.';
COMMENT ON COLUMN security_event_logs.anonymous_user_id IS 'The UUID of the actor associated with the event.';
COMMENT ON COLUMN security_event_logs.source            IS 'The site section (WEB, BACKEND, etc.) that generated the event.';
COMMENT ON COLUMN security_event_logs.argument          IS 'An arbitrary JSON blob containing event data.';
COMMENT ON COLUMN security_event_logs.version           IS 'The version of Sourcegraph which generated the event.';

COMMIT;
