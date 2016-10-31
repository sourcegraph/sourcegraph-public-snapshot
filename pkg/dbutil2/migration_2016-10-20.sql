CREATE table global_ref_source (
	id serial primary key NOT NULL,
	source text NOT NULL,
	UNIQUE(source)
);
CREATE INDEX global_ref_source_source ON global_ref_source USING btree (source);

CREATE table global_ref_version (
	id serial primary key NOT NULL,
	version text NOT NULL,
	UNIQUE(version)
);
CREATE INDEX global_ref_version_version ON global_ref_version USING btree (version);

CREATE table global_ref_file (
	id serial primary key NOT NULL,
	file text NOT NULL,
	UNIQUE(file)
);
CREATE INDEX global_ref_file_file ON global_ref_file USING btree (file);

CREATE table global_ref_name (
	id serial primary key NOT NULL,
	name text NOT NULL,
	UNIQUE(name)
);
CREATE INDEX global_ref_name_name ON global_ref_name USING btree (name);

CREATE table global_ref_container (
	id serial primary key NOT NULL,
	container text NOT NULL,
	UNIQUE(container)
);
CREATE INDEX global_ref_container_container ON global_ref_container USING btree (container);

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
CREATE INDEX global_ref_by_source_def_name ON global_ref_by_source USING btree (def_name);
CREATE INDEX global_ref_by_source_def_container ON global_ref_by_source USING btree (def_container);
CREATE INDEX global_ref_by_source_def_scheme ON global_ref_by_source USING btree (def_scheme);
CREATE INDEX global_ref_by_source_def_source ON global_ref_by_source USING btree (def_source);
CREATE INDEX global_ref_by_source_def_version ON global_ref_by_source USING btree (def_version);
CREATE INDEX global_ref_by_source_scheme ON global_ref_by_source USING btree (scheme);
CREATE INDEX global_ref_by_source_source ON global_ref_by_source USING btree (source);
CREATE INDEX global_ref_by_source_version ON global_ref_by_source USING btree (version);
CREATE INDEX global_ref_by_source_refs ON global_ref_by_source USING btree (refs);
CREATE INDEX global_ref_by_source_score ON global_ref_by_source USING btree (score);

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
CREATE INDEX global_ref_by_file_def_name ON global_ref_by_file USING btree (def_name);
CREATE INDEX global_ref_by_file_def_container ON global_ref_by_file USING btree (def_container);
CREATE INDEX global_ref_by_file_def_scheme ON global_ref_by_file USING btree (def_scheme);
CREATE INDEX global_ref_by_file_def_source ON global_ref_by_file USING btree (def_source);
CREATE INDEX global_ref_by_file_def_version ON global_ref_by_file USING btree (def_version);
CREATE INDEX global_ref_by_file_scheme ON global_ref_by_file USING btree (scheme);
CREATE INDEX global_ref_by_file_source ON global_ref_by_file USING btree (source);
CREATE INDEX global_ref_by_file_version ON global_ref_by_file USING btree (version);
CREATE INDEX global_ref_by_file_score ON global_ref_by_file USING btree (score);
