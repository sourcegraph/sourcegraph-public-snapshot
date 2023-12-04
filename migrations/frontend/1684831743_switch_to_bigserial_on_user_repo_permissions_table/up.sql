-- change the id column to bigint
ALTER TABLE IF EXISTS user_repo_permissions 
ALTER COLUMN id TYPE BIGINT;

-- update the sequence
ALTER SEQUENCE IF EXISTS user_repo_permissions_id_seq AS BIGINT OWNED BY user_repo_permissions.id;
