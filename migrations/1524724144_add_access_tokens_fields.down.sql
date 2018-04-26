ALTER TABLE access_tokens RENAME COLUMN subject_user_id TO user_id;
ALTER TABLE access_tokens RENAME CONSTRAINT access_tokens_subject_user_id_fkey TO access_tokens_user_id_fkey;

ALTER TABLE access_tokens DROP COLUMN creator_user_id;
