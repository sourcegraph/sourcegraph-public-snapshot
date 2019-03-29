begin;

create table migrations (
  id          bigserial primary key,
  name        text not null,
  description text not null,
  started_at  timestamptz not null default now(),
  finished_at timestamptz not null default now(),
  metadata    jsonb not null default '{}' check (jsonb_typeof(metadata) = 'object')
);

commit;
