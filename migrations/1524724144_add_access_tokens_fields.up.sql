-- Rename column because we want to store both the subject and the creator separately.
ALTER TABLE access_tokens RENAME COLUMN user_id TO subject_user_id;
ALTER TABLE access_tokens RENAME CONSTRAINT access_tokens_user_id_fkey TO access_tokens_subject_user_id_fkey;

ALTER TABLE access_tokens ADD COLUMN creator_user_id integer REFERENCES users(id);
-- Assume that the subject is the creator. This is not always true; site admins can create an access
-- token for a user. However, it is a defensible assumption.
UPDATE access_tokens SET creator_user_id=subject_user_id;
ALTER TABLE access_tokens ALTER COLUMN creator_user_id SET NOT NULL;
