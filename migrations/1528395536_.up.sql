CREATE TABLE registry_extensions(
       id serial NOT NULL PRIMARY KEY,
       uuid uuid NOT NULL,
       publisher_user_id integer REFERENCES users(id),
       publisher_org_id integer REFERENCES orgs(id),
       name citext NOT NULL,
       manifest text,
       created_at timestamp with time zone NOT NULL DEFAULT now(),
       updated_at timestamp with time zone NOT NULL DEFAULT now(),
       deleted_at timestamp with time zone,
       CONSTRAINT registry_extensions_single_publisher CHECK((publisher_user_id IS NULL) != (publisher_org_id IS NULL)),
       CONSTRAINT registry_extensions_name_valid_chars CHECK (name ~ E'^[a-zA-Z0-9](?:[a-zA-Z0-9]|[_.-](?=[a-zA-Z0-9]))*$'::citext),
       CONSTRAINT registry_extensions_name_length CHECK (char_length(name) > 0 AND char_length(name) <= 128)
);

CREATE UNIQUE INDEX registry_extensions_uuid ON registry_extensions(uuid);
CREATE UNIQUE INDEX registry_extensions_publisher_name ON registry_extensions(publisher_user_id, publisher_org_id, name) WHERE deleted_at IS NULL;
