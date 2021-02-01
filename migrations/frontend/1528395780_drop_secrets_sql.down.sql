BEGIN;

-- Copied as-is from the squashed migration file.
CREATE TABLE secrets (
    id bigint NOT NULL,
    source_type character varying(50),
    source_id bigint,
    key_name character varying(100),
    value text NOT NULL
);

CREATE SEQUENCE secrets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE secrets_id_seq OWNED BY secrets.id;

COMMIT;
