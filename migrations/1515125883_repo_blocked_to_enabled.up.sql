-- All repositories are enabled prior to this migration running, because there was no
-- API or UI for disabling (blocking) them.
ALTER TABLE repo ADD COLUMN enabled boolean NOT NULL default true;

-- In case the user has rolled back and rolls forward again, repopulate enabled from blocked.
UPDATE repo SET enabled=false WHERE blocked;

-- Because this migration was changed retroactively (from c6f0adfb43111f8ffb1222d7625cbff9dda067e0),
-- there are 2 possible states as of schema_migrations.version=1515125883: (1) the repo.blocked column
-- exists, and (2) it does not exist. In a future migration we will handle making these consistent.
-- In the meantime, nothing reads from the repo.blocked column anymore, so it's harmless.
