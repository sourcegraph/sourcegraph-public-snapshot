-- Allow underscores in usernames.

ALTER TABLE users
  DROP CONSTRAINT users_username_valid_chars,
    ADD CONSTRAINT users_username_valid_chars CHECK (username ~ '^\w(?:\w|[-.](?=\w))*-?$'::citext);
