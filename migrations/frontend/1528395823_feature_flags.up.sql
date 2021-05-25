BEGIN;

CREATE TYPE feature_flag_type AS ENUM ('bool', 'bool_var');

CREATE TABLE IF NOT EXISTS feature_flags (
	flag_name text NOT NULL PRIMARY KEY,
	flag_type feature_flag_type NOT NULL,
	bool_value boolean,
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

COMMENT ON COLUMN feature_flags.bool_value IS 'Bool value only defined when flag_type is bool';
COMMENT ON COLUMN feature_flags.rollout IS 'Rollout only defined when flag_type is bool_var. Increments of 0.01%';
COMMENT ON CONSTRAINT required_bool_fields ON feature_flags IS 'Checks that bool_value is set IFF flag_type = bool';
COMMENT ON CONSTRAINT required_bool_var_fields ON feature_flags IS 'Checks that rollout is set IFF flag_type = bool_var';

CREATE TABLE IF NOT EXISTS feature_flag_overrides (
	namespace_org_id integer,
	namespace_user_id integer,
	flag_name text NOT NULL,
	flag_value boolean NOT NULL,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone DEFAULT now() NOT NULL,
	deleted_at timestamp with time zone,

	CONSTRAINT feature_flag_overrides_unique_user_flag 
		UNIQUE (namespace_user_id, flag_name),

	CONSTRAINT feature_flag_overrides_unique_org_flag 
		UNIQUE (namespace_org_id, flag_name),

	CONSTRAINT feature_flag_overrides_has_org_or_user_id CHECK(
		(namespace_org_id IS NOT NULL) OR (namespace_user_id IS NOT NULL)),

	FOREIGN KEY (flag_name) REFERENCES feature_flags (flag_name) ON DELETE CASCADE,
	FOREIGN KEY (namespace_org_id) REFERENCES orgs (id) ON DELETE CASCADE,
	FOREIGN KEY (namespace_user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX feature_flag_overrides_org_id
	ON feature_flag_overrides (namespace_org_id)
	WHERE namespace_org_id IS NOT NULL;

CREATE INDEX feature_flag_overrides_user_id
	ON feature_flag_overrides (namespace_user_id)
	WHERE namespace_user_id IS NOT NULL;

COMMIT;
