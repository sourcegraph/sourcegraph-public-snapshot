-- Drop these tables, which have been unused for 2+ major releases. (The users.tags column has been
-- used instead.)
DROP TABLE user_tags;
DROP TABLE org_tags;
