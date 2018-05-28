
CREATE TABLE deployment_configuration (
	id int primary key,
	app_id uuid not NULL,
	enable_telemetry boolean DEFAULT true,
	email text,
	last_updated text
);

