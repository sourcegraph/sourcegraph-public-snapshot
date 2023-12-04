-- Undo allow underscores in usernames.

ALTER TABLE users
  DROP CONSTRAINT users_username_valid_chars,
    ADD CONSTRAINT users_username_valid_chars CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'::citext);
