CREATE TABLE survey_responses (
       id bigserial NOT NULL PRIMARY KEY,
       user_id integer REFERENCES users (id),
       email text,
       score integer NOT NULL,
       reason text,
       better text,
       created_at timestamp with time zone NOT NULL DEFAULT now(),
       updated_at timestamp with time zone
);
