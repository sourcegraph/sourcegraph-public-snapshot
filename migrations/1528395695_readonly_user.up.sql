begin;
  CREATE USER sgreader with PASSWORD 'sgreader';
  do
  $$
  begin
      execute format('grant connect on database %I to sgreader', current_database());
  end;
  $$;
  GRANT USAGE ON SCHEMA public to sgreader;
  GRANT SELECT ON ALL TABLES IN SCHEMA public to sgreader;
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO sgreader;
COMMIT;
