BEGIN;

DROP TABLE IF EXISTS "query_runner_state";
DROP TABLE IF EXISTS "saved_searches";
DROP TYPE IF EXISTS "user_or_org";
ALTER TABLE "saved_searches" DROP CONSTRAINT user_or_org_id_not_null;

COMMIT;
