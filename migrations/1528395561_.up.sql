CREATE TABLE external_services (
  id bigserial NOT NULL PRIMARY KEY,
  kind text NOT NULL,
  display_name text NOT NULL,
  config text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at timestamp with time zone
);
