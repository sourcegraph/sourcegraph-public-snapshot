CREATE table global_ref_source (
	id serial primary key NOT NULL,
	source text NOT NULL,
	UNIQUE(source)
);

CREATE table global_ref_version (
	id serial primary key NOT NULL,
	version text NOT NULL,
	UNIQUE(version)
);

CREATE table global_ref_file (
	id serial primary key NOT NULL,
	file text NOT NULL,
	UNIQUE(file)
);

CREATE table global_ref_name (
	id serial primary key NOT NULL,
	name text NOT NULL,
	UNIQUE(name)
);

CREATE table global_ref_container (
	id serial primary key NOT NULL,
	container text NOT NULL,
	UNIQUE(container)
);

CREATE table global_ref_by_source (
	def_name integer references global_ref_name(id) NOT NULL,
	def_container integer references global_ref_container(id) NOT NULL,
	def_scheme smallint NOT NULL,
	def_source integer references global_ref_source(id) NOT NULL,
	def_version integer references global_ref_version(id) NOT NULL,
	def_file integer references global_ref_file(id) NOT NULL,
	scheme smallint NOT NULL,
	source integer references global_ref_source(id) NOT NULL,
	version integer references global_ref_version(id) NOT NULL,
	files smallint NOT NULL,
	refs smallint NOT NULL,
	score smallint NOT NULL,
	UNIQUE(def_name, def_container, def_scheme, def_source, def_version, def_file, scheme, source, version)
);

CREATE table global_ref_by_file (
	def_name integer references global_ref_name(id) NOT NULL,
	def_container integer references global_ref_container(id) NOT NULL,
	def_scheme smallint NOT NULL,
	def_source integer references global_ref_source(id) NOT NULL,
	def_version integer references global_ref_version(id) NOT NULL,
	def_file integer references global_ref_file(id) NOT NULL,
	scheme smallint NOT NULL,
	source integer references global_ref_source(id) NOT NULL,
	version integer references global_ref_version(id) NOT NULL,
	file integer references global_ref_file(id) NOT NULL,
	positions integer[] NOT NULL,
	score smallint NOT NULL,
	UNIQUE(def_name, def_container, def_scheme, def_source, def_version, def_file, scheme, source, version, file)
);
