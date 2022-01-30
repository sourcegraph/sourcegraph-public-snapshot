BEGIN;

-- We encountered performance issues for our use cases when we deployed
-- RLS to production. We made the decision to back that approach out and
-- solve the security concerns in application-level code instead.
--
-- ref migrations/frontend/1528395860_remove_repo_table_policy.up.sql
-- ref migrations/frontend/1528395861_remove_sg_service_grants.up.sql
-- ref migrations/frontend/1528395862_remove_sg_service_role.up.sql

COMMIT;
