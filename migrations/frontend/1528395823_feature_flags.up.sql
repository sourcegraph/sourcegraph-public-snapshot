BEGIN;

CREATE TYPE feature_flag_type AS ENUM ('bool', 'bool_var');

CREATE TABLE IF NOT EXISTS feature_flags (
	flag_name text NOT NULL PRIMARY KEY,
	flag_type feature_flag_type NOT NULL,

	-- Bool value only defined when flag_type is 'bool'
	bool_value boolean,
	
	-- Rollout only defined when flag_type is 'bool_var'. Increments of 0.01%
	rollout integer CHECK (rollout >= 0 AND rollout <= 10000),

	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone DEFAULT now() NOT NULL,
	deleted_at timestamp with time zone,

	CONSTRAINT required_bool_fields	CHECK ( 1 = CASE
		WHEN flag_type = 'bool' AND bool_value IS NULL THEN 0
		WHEN flag_type <> 'bool' AND bool_value IS NOT NULL THEN 0
		ELSE 1
	END),

	CONSTRAINT required_bool_var_fields CHECK (1 = CASE
		WHEN flag_type = 'bool_var' AND rollout IS NULL THEN 0
		WHEN flag_type <> 'bool_var' AND rollout IS NOT NULL THEN 0
		ELSE 1
	END)
);


CREATE TABLE IF NOT EXISTS feature_flag_overrides (
	namespace_user_id integer NOT NULL,
	flag_name text NOT NULL,
	flag_value boolean NOT NULL,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone DEFAULT now() NOT NULL,
	deleted_at timestamp with time zone,

	UNIQUE (namespace_user_id, flag_name),

	FOREIGN KEY (flag_name) REFERENCES feature_flags (flag_name) ON DELETE CASCADE,
	FOREIGN KEY (namespace_org_id) REFERENCES orgs (id) ON DELETE CASCADE,
	FOREIGN KEY (namespace_user_id) REFERENCES users (id) ON DELETE CASCADE
);

COMMIT;
