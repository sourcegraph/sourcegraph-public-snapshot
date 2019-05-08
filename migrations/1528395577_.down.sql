BEGIN;

DROP TABLE IF EXISTS "query_runner_state";
ALTER TABLE "saved_searches" DROP CONSTRAINT user_or_org_id_not_null;
DROP TABLE IF EXISTS "saved_searches";

COMMIT;
