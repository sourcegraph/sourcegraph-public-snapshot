CREATE EXTENSION IF NOT EXISTS citext;

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';

CREATE TABLE console_users (
    id integer NOT NULL,
    email citext NOT NULL
);

CREATE SEQUENCE console_users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE console_users_id_seq OWNED BY console_users.id;

ALTER TABLE ONLY console_users ALTER COLUMN id SET DEFAULT nextval('console_users_id_seq'::regclass);

ALTER TABLE ONLY console_users
    ADD CONSTRAINT console_users_pkey PRIMARY KEY (id);