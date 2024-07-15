DROP VIEW IF EXISTS prompts_view;

DROP INDEX IF EXISTS prompts_name_is_unique_in_owner_user;
DROP INDEX IF EXISTS prompts_name_is_unique_in_owner_org;

DROP TABLE IF EXISTS prompts;
