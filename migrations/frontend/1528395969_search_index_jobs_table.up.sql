-- +++
-- parent: 1528395968
-- +++

BEGIN;

CREATE TABLE serach_index_jobs (
  id                SERIAL PRIMARY KEY,
  state             text DEFAULT 'queued',
  failure_message   text,
  started_at        timestamp with time zone,
  finished_at       timestamp with time zone,
  process_after     timestamp with time zone,
  num_resets        integer not null default 0,
  num_failures      integer not null default 0,
  last_heartbeat_at timestamp with time zone,
  execution_logs    json[],
  worker_hostname   text not null default '',

  repository_id integer not null,
  revision      text not null
);

COMMIT;
