Migrations
==========

## 2015 Dec 03

```
alter table users add column write boolean default 'f';
update users set write = 't' where admin = 't';
```

## 2015 Aug 02

```
alter table repo_build_task drop column queue;
```

## 2015 July 20

Run

```
$ psql < migration_2015-07-20_01_rename.sql
$ src pgsql create
$ psql < migration_2015-07-20_02_migrate.sql
$ psql < migration_2015-07-20_03_drop.sql
```

Note that this takes a while to run. It migrates all data from the old master
schema to the new `nodb` schema. Please don't run this against `live2` rds
instance, but rather the new stack.

## 2015 Apr 19

```
ALTER TABLE repo_settings RENAME TO repo_config;
```

## 2015 Apr 8

```
DROP TABLE repo_queue_refresh_profile;
```

## 2015 Apr 08

```
ALTER TABLE repo_build DROP COLUMN pull_repo;
ALTER TABLE repo_build DROP COLUMN pull_number;
```

## 2015 March 27

```
ALTER TABLE repo_settings ADD COLUMN enabled character varying(255) default 'true';
```

## 2015 Feb 22

```
ALTER TABLE person_settings DROP COLUMN buildemails;
```

## 2015 Feb 20

```
ALTER TABLE repo_build ADD COLUMN pull_repo integer DEFAULT 0;
ALTER TABLE repo_build ADD COLUMN pull_number integer DEFAULT 0;

ALTER TABLE repo_settings ADD COLUMN admin_uid integer;

ALTER TABLE repo_settings DROP COLUMN srcbotenabled;
ALTER TABLE repo_settings ADD COLUMN external_commit_statuses boolean;
ALTER TABLE repo_settings ADD COLUMN unsuccessful_external_commit_statuses boolean;
```

## 2015 Jan 3

```
DROP INDEX repo_search;
CREATE INDEX CONCURRENTLY repo_name_lower ON repo(lower(name) text_pattern_ops);
DROP INDEX repo_name;
```

## 2015 Jan 1

```
CREATE INDEX CONCURRENTLY ref_def_xref ON ref(def_repo, def_path, def_unit_type, def_unit, repo) WHERE def_repo != repo AND canonical AND (NOT def);
CREATE INDEX CONCURRENTLY ref_def_rref ON ref(def_repo, def_path, def_unit_type, def_unit, commit_id) WHERE def_repo=repo AND (NOT def);
CREATE INDEX CONCURRENTLY def_name_lower2 ON def(repo, commit_id, lower(name) text_pattern_ops) WHERE local!=true;

-- Destructive stuff. Run this only after the new code is deployed.
ALTER TABLE def DROP COLUMN indexed_globally;
DROP INDEX def_name_lower;
DROP INDEX def_name;
ALTER INDEX def_name_lower2 RENAME TO def_name_lower;
DROP INDEX def_search_terms_idx CASCADE;
```

## 2014 Dec 30

```
CREATE INDEX CONCURRENTLY def_repo_file2 ON def(repo, commit_id, file text_pattern_ops) WHERE local!=true;
BEGIN;
DROP INDEX def_repo_file;
ALTER INDEX def_repo_file2 RENAME TO def_repo_file;
COMMIT;
```

## 2014 Dec 30

```
ALTER TABLE ref ALTER COLUMN canonical DROP default;
```

## 2014 Dec 30

```
CREATE UNIQUE INDEX CONCURRENTLY doc_pkey2 ON doc(repo, commit_id, path, unit, unit_type, format);

BEGIN;
ALTER TABLE doc DROP CONSTRAINT doc_pkey;
ALTER INDEX doc_pkey2 RENAME TO doc_pkey_idx;
ALTER TABLE doc ADD CONSTRAINT doc_pkey PRIMARY KEY USING INDEX doc_pkey_idx;
COMMIT;

```

## 2014 Dec 29

```
CREATE UNIQUE INDEX CONCURRENTLY def_stat_pkey2 ON def_stat(repo, commit_id, type, unit_type, unit, path);

-- I have run these on sgg00 and sgg02, still need to run them on sgg01
BEGIN;
ALTER TABLE def_stat DROP CONSTRAINT def_stat_pkey;
ALTER INDEX def_stat_pkey2 RENAME TO def_stat_pkey_idx;
ALTER TABLE def_stat ADD CONSTRAINT def_stat_pkey PRIMARY KEY USING INDEX def_stat_pkey_idx;
COMMIT;
```

## 2014 Dec 19

```
CREATE TABLE feature_flag_user(
  uid integer NOT NULL,
  features text[],
  CONSTRAINT feature_flag_user_pkey PRIMARY KEY (uid)
);
```

## 2014 Dec 18

```
ALTER TABLE repo ADD COLUMN created_at timestamp with time zone default '0001-01-01 00:00:00';
ALTER TABLE repo ADD COLUMN updated_at timestamp with time zone default '0001-01-01 00:00:00';
ALTER TABLE repo ADD COLUMN pushed_at timestamp with time zone default '0001-01-01 00:00:00';
```

## 2014 Dec 16

```
ALTER TABLE user_email ADD COLUMN public boolean DEFAULT false;
```

## 2014 Dec 14

```
CREATE TABLE ann
(
  repo citext NOT NULL,
  commit_id character varying(255) NOT NULL,
  unit_type character varying(255) NOT NULL,
  unit character varying(255) NOT NULL,
  type character varying(255) NOT NULL,
  file character varying(255) NOT NULL,
  start integer NOT NULL,
  "end" integer NOT NULL,
  data bytea,
  CONSTRAINT ann_pkey PRIMARY KEY (repo, commit_id, unit, unit_type, file, start, "end", type)
);
ALTER TABLE unit ADD COLUMN private boolean;
CREATE INDEX ann_repo_commit_id_file ON ann(repo, commit_id, file);
CREATE INDEX unit_dockerfile_search ON unit USING gin(to_tsvector('simple', encode(data, 'escape'))) WHERE unit_type='Dockerfile';
```

## 2014 Dec 10

```
-- already migrated github.com/golang/go
-- begin; delete from repo where rid=2958402;update repo set github_id=23096959,owner_github_user_id=4314092,github_stars=2598,language='Go',uri_alias='github.com/golang/go' where uri='code.google.com/p/go';commit;

WITH delrepo AS (
  DELETE FROM repo WHERE uri IN('github.com/golang/codereview','github.com/golang/tools','github.com/golang/text','github.com/golang/talks','github.com/golang/net','github.com/golang/crypto','github.com/golang/image','github.com/golang/blog','github.com/golang/exp','github.com/golang/benchmarks','github.com/golang/mobile','github.com/golang/sys','github.com/golang/wiki','github.com/golang/example') RETURNING uri, github_id
) UPDATE repo r0 SET uri=replace(uri, 'code.google.com/p/go.', 'github.com/golang/'), uri_alias=uri, vcs='git', default_branch='master', http_clone_url=concat(replace(uri, 'code.google.com/p/go.', 'https://github.com/golang/'), '.git'), ssh_clone_url=concat(replace(uri, 'code.google.com/p/go.', 'ssh://git@github.com/golang/'), '.git'), owner_github_user_id=4314092, language='Go',owner_user_id=3441, github_id=(SELECT github_id FROM delrepo WHERE uri=replace(r0.uri, 'code.google.com/p/go.', 'github.com/golang/')) WHERE uri IN('code.google.com/p/go.codereview','code.google.com/p/go.tools','code.google.com/p/go.text','code.google.com/p/go.talks','code.google.com/p/go.net','code.google.com/p/go.crypto','code.google.com/p/go.image','code.google.com/p/go.blog','code.google.com/p/go.exp','code.google.com/p/go.benchmarks','code.google.com/p/go.mobile','code.google.com/p/go.sys','code.google.com/p/go.wiki','code.google.com/p/go.example');

UPDATE repo SET uri='github.com/golang/go', uri_alias='code.google.com/p/go', http_clone_url='https://go.googlesource.com/go', ssh_clone_url='git://git@github.com/golang/go.git', vcs='git', default_branch='master' WHERE uri='code.google.com/p/go';
```

## 2014 Dec 05

```
CREATE UNIQUE INDEX CONCURRENTLY repo_uri_alias ON repo(uri_alias);
CREATE UNIQUE INDEX CONCURRENTLY repo_uri_and_uri_alias_uniq1 ON repo((greatest(uri, uri_alias)));
CREATE UNIQUE INDEX CONCURRENTLY repo_uri_and_uri_alias_uniq2 ON repo((least(uri, uri_alias)));
ALTER TABLE repo ADD CHECK (uri != uri_alias);
```

## 2014 Dec 3
```
CREATE INDEX CONCURRENTLY repo_owner ON repo(owner_user_id);
```

## 2014 Dec 3

```
BEGIN;
ALTER TABLE repo DROP COLUMN actual_clone_url;
ALTER TABLE repo ADD COLUMN uri_alias citext;
UPDATE repo SET uri_alias=replace(uri, 'sourcegraph.com/', 'github.com/') WHERE github_id IS NOT NULL AND uri LIKE 'sourcegraph.com/%';
UPDATE repo SET uri=replace(uri, 'github.com/', 'sourcegraph.com/'), uri_alias=uri WHERE github_id IS NOT NULL AND replace(uri, 'github.com/sourcegraph/', '') IN ('sourcegraph','go-sourcegraph','vcsstore','go-vcs','s3cache','go-selenium','multicache','rwvfs','srclib-php','srclib-go','srclib','srclib-haskell','srclib-javascript','srclib-java','srclib-ruby','srclib-python','srclib-sample') and owner_user_id=36;
ALTER TABLE repo ADD COLUMN ssh_clone_url TEXT;
UPDATE repo SET ssh_clone_url=concat('ssh://git@', uri, '.git') WHERE github_id IS NOT NULL;
ALTER TABLE repo RENAME COLUMN clone_url TO http_clone_url;
COMMIT;
```

## 2014 Dec 02

```
BEGIN;
ALTER TABLE repo_build_task RENAME COLUMN unittype TO unit_type;
ALTER TABLE repo_build_task ADD COLUMN "created_at" timestamp with time zone;
UPDATE repo_build_task SET created_at=started_at WHERE created_at IS NULL;
ALTER TABLE repo_build_task ADD COLUMN "queue" boolean default false;
COMMIT;
CREATE INDEX repo_build_task_queue ON repo_build_task(created_at ASC) WHERE started_at IS NULL AND queue;
```

## 2014 Dec 1 (xref stats)
```
CREATE TABLE ref_by_unit (
    repo citext,
    unit_type character varying(255),
    unit character varying(255),
    commit_id character varying(255),
    def_repo citext,
    def_unit_type character varying(255),
    def_unit character varying(255),
    def_path text,
    count integer,
    canonical boolean,
    private boolean
    );
CREATE INDEX ref_by_unit_def ON ref_by_unit USING btree (def_repo, def_unit_type, def_unit, def_path);
CREATE INDEX ref_by_unit_origin ON ref_by_unit USING btree (repo, commit_id, unit_type, unit);
```

## 2014 Nov 13 (sharding)

```sql
-- Remove def.sid and use the concrete key as the new primary key
ALTER TABLE def DROP CONSTRAINT def_pkey;
ALTER TABLE def DROP COLUMN sid;
ALTER TABLE def ADD PRIMARY KEY USING INDEX def_concrete_key;
ANALYZE VERBOSE def;

DROP TABLE unit_author CASCADE;
DROP TABLE unit_client_unit CASCADE;
DROP TABLE refd_def_by_author_unit CASCADE;
DROP TABLE ref_author CASCADE;
DROP TABLE def_author CASCADE;
```

## 2014 Nov 10
```sql
ALTER TABLE def ADD COLUMN private boolean;
ALTER TABLE ref ADD COLUMN private boolean;
UPDATE def SET private=true WHERE repo IN (SELECT uri FROM repo WHERE private);
UPDATE ref SET private=true WHERE repo IN (SELECT uri FROM repo WHERE private);
```

## 2014 Nov 06

```sql
ALTER TABLE repo_build ADD COLUMN killed boolean DEFAULT false;
ALTER TABLE repo_build ADD COLUMN heartbeat_at timestamp with time zone;
```

## 2014 Nov 05

```sql
ALTER TABLE repo_build ADD COLUMN purged boolean DEFAULT false;
```

## 2014 Oct 30

```sql
ALTER TABLE repo_build DROP COLUMN tries;
```

## 2014 Oct 28

```sql
CREATE TABLE repo_key(
  rid integer NOT NULL,
  private_key_pem text,
  CONSTRAINT repo_key_pkey PRIMARY KEY (rid)
);
```

## 2014 Oct 27

```sql
ALTER TABLE repo_settings ADD COLUMN use_ssh_private_key boolean DEFAULT false;
```

## 2014 Oct 27

```sql
ALTER TABLE repo_settings RENAME COLUMN enabled TO build_pushes;
```

## 2014 Oct 27

Add actor field to all of the simple queue tables and their primary keys:

```sql
BEGIN;
ALTER TABLE person_queue_refresh_profile ADD COLUMN actor integer NOT NULL DEFAULT 0;
ALTER TABLE person_queue_refresh_profile DROP CONSTRAINT person_queue_refresh_profile_pkey;
ALTER TABLE person_queue_refresh_profile ADD PRIMARY KEY (id, actor, task);
COMMIT;
BEGIN;
ALTER TABLE person_queue_compute_stats ADD COLUMN actor integer NOT NULL DEFAULT 0;
ALTER TABLE person_queue_compute_stats DROP CONSTRAINT person_queue_compute_stats_pkey;
ALTER TABLE person_queue_compute_stats ADD PRIMARY KEY (id, actor, task);
COMMIT;
BEGIN;
ALTER TABLE repo_queue_compute_stats ADD COLUMN actor integer NOT NULL DEFAULT 0;
ALTER TABLE repo_queue_compute_stats DROP CONSTRAINT repo_queue_compute_stats_pkey;
ALTER TABLE repo_queue_compute_stats ADD PRIMARY KEY (id, actor, task);
COMMIT;
BEGIN;
ALTER TABLE repo_queue_refresh_profile ADD COLUMN actor integer NOT NULL DEFAULT 0;
ALTER TABLE repo_queue_refresh_profile DROP CONSTRAINT repo_queue_refresh_profile_pkey;
ALTER TABLE repo_queue_refresh_profile ADD PRIMARY KEY (id, actor, task);
COMMIT;
BEGIN;
ALTER TABLE repo_queue_refresh_vcs_data ADD COLUMN actor integer NOT NULL DEFAULT 0;
ALTER TABLE repo_queue_refresh_vcs_data DROP CONSTRAINT repo_queue_refresh_vcs_data_pkey;
ALTER TABLE repo_queue_refresh_vcs_data ADD PRIMARY KEY (id, actor, task);
COMMIT;
BEGIN;
ALTER TABLE repo_index ADD COLUMN actor integer NOT NULL DEFAULT 0;
ALTER TABLE repo_index DROP CONSTRAINT repo_index_pkey;
ALTER TABLE repo_index ADD PRIMARY KEY (id, actor, task);
COMMIT;
```

## 2014 Oct 15

```sql
ALTER TABLE person_settings DROP COLUMN srcbot;
ALTER TABLE repo_settings ADD COLUMN srcbotenabled boolean default false;
```

## 2014 Oct 09 (bug fix)
```sql
DROP INDEX CONCURRENTLY user_email_email_primary;
CREATE UNIQUE INDEX user_email_email_primary ON user_email(email) WHERE (NOT blacklisted) AND "primary";
```

## 2014 Oct 09

```sql
ALTER TABLE person_settings ADD COLUMN srcbot boolean default false;
```

## 2014 Oct 06
```sql
CREATE INDEX user_email_email ON user_email(email) WHERE (NOT blacklisted);
CREATE UNIQUE INDEX user_email_email_primary ON user_email(email, "primary") WHERE (NOT blacklisted);
DROP INDEX CONCURRENTLY user_email2_email;
```

## 2014 Oct 4
```sql
alter table repo add column actual_clone_url character varying(255);
```

## 2014 Sep 12

```sql
alter table person_settings add column buildemails boolean;
alter table person_settings alter column planid drop not null;
alter table person_settings alter column planid set default null;
```

## 2014 Sep 09

All the migrations to upgrade to the **x** branch's code.

```sql
\set ON_ERROR_STOP on
alter table def add column indexed_globally bool default true;
alter table ref add column canonical bool default true;
alter table ref_by_unit add column canonical bool default true;


create table def_x as (select *, true as indexed_globally from def);
create table ref_x as (select *, true as indexed_globally from ref);

CREATE OR REPLACE FUNCTION def_x_search_terms(s def_x) RETURNS tsvector AS $$
SELECT to_tsvector(
  array_to_string(regexp_split_to_array(s.treepath || ' ' || s.unit, '[/]+'), ' ')
)
$$ LANGUAGE SQL IMMUTABLE RETURNS NULL ON NULL INPUT;
CREATE INDEX def_x_search_terms_idx
  ON def_x
  USING gin
  (def_x_search_terms(def_x.*))
  WHERE indexed_globally AND exported AND NOT test;

create sequence def_x_sid_seq minvalue 64859626;
ALTER TABLE def_x ALTER COLUMN sid SET DEFAULT nextval('def_x_sid_seq');
ALTER SEQUENCE def_x_sid_seq OWNED BY def_x.sid;
CREATE UNIQUE INDEX def_x_concrete_key ON def_x(repo, commit_id, unit, unit_type, path);
CREATE INDEX def_x_name ON def_x(name text_pattern_ops) WHERE exported AND NOT test;
CREATE INDEX def_x_name_lower ON def_x(lower(name) text_pattern_ops) WHERE exported AND NOT test;
CREATE INDEX def_x_repo_file ON def_x(repo, file text_pattern_ops) WHERE exported AND NOT test;


CREATE INDEX ref_x_origin ON ref_x(repo,commit_id,unit_type,unit); -- in progress
create index ref_x_pkey on ref_x(def_repo, def_unit_type, def_unit, def_path, def, repo, unit_type, unit, file, commit_id, start, "end");
alter table ref_x add primary key using index ref_x_pkey;
CREATE INDEX ref_x_location ON ref_x(repo,commit_id,file,start,"end");


--------------

CREATE TABLE ext_auth_token
(
  "user" integer NOT NULL,
  host character varying(255) NOT NULL,
  token character varying(255),
  scope character varying(255),
  refreshed_at timestamp with time zone,
  client_id character varying(255) NOT NULL,
  disabled boolean,
  auth_failure_count integer,
  first_auth_failure_at timestamp with time zone,
  first_auth_failure_message character varying(1000),
  CONSTRAINT ext_auth_token_pkey PRIMARY KEY ("user", host, client_id)
);
CREATE TABLE old_users_github_oauth_tokens AS (SELECT uid, login, github_oauth2_access_token FROM users WHERE registered_at IS NOT NULL);
ALTER TABLE users DROP COLUMN github_oauth2_access_token;
CREATE UNIQUE INDEX ext_auth_token_token_host
  ON ext_auth_token
  USING btree
  (token COLLATE pg_catalog."default", host COLLATE pg_catalog."default", client_id COLLATE pg_catalog."default");


CREATE TABLE person_settings
(
  uid integer NOT NULL,
  planid character varying(64) NOT NULL,
  requestedupgradeat timestamp with time zone,
  CONSTRAINT person_settings_pkey PRIMARY KEY (uid)
);

CREATE TABLE org_settings
(
  uid integer NOT NULL,
  planid character varying(64) NOT NULL,
  CONSTRAINT org_settings_pkey PRIMARY KEY (uid)
);

CREATE TABLE repo_index
(
  id integer NOT NULL,
  task character varying(255) NOT NULL,
  enqueued_at timestamp with time zone,
  started_at timestamp with time zone,
  ended_at timestamp with time zone,
  CONSTRAINT repo_index_pkey PRIMARY KEY (id, task)
);

CREATE TABLE repo_settings
(
  rid integer NOT NULL,
  enabled character varying(255),
  CONSTRAINT repo_settings_pkey PRIMARY KEY (rid)
);

------------------

ALTER TABLE user_email RENAME TO user_email_old0;
CREATE TABLE user_email
(
  uid integer NOT NULL,
  email citext NOT NULL,
  verified boolean,
  "primary" boolean,
  guessed boolean,
  blacklisted boolean,
  CONSTRAINT user_email_pkey2 PRIMARY KEY (uid, email)
);
CREATE UNIQUE INDEX user_email2_email
  ON user_email
  USING btree
  (email COLLATE pg_catalog."default")
  WHERE NOT blacklisted;
INSERT INTO user_email(uid,email,verified,"primary",guessed,blacklisted) SELECT min(uid),email,false,false,true,false FROM user_email_old0 group by email;
```

## 2014 Sep 08

```sql
alter table def add column indexed_globally bool default true;
alter table ref add column indexed_globally bool default true;
alter table ref_by_unit add column indexed_globally bool default false;
```

## 2014 Sep 5
```sql
alter table person_settings add column requestedUpgradeAt timestamp with time zone default null;
```

## 2014 Aug 24

```sql
alter table def drop column callable;
```

## 2014 Aug 22

```sql
alter table repo add column private bool default false;
```

## 2014 Aug 20 (change BuildTask fields)

```sql
alter table repo_build_task rename tid to taskid;
alter table repo_build_task add column "order" int not null default 0;
alter table repo_build_task rename title to op;
```

## 2014 Aug 13 (rename symbol to def)

see sym_to_def.sql

## 2014 Jan 12 (js-improvements)

```sql
UPDATE sym SET specific_kind = 'commonjs_module' WHERE specific_kind = 'nodejs_module_file';
UPDATE sym SET specific_kind = 'npm_package' WHERE specific_kind = 'nodejs_package';
```


## 2013 Dec 17 (branch Improve-Grapher-interface)
```sql
alter table sym drop column pkg;
```

## 2013 Dec 15 (branch save-README-filename)
```sql
alter table repo add column readme_filename character varying;
```


## 2013 Dec 13 (branch Check-ref-uniqueness-in-Go-not-PostgreSQL)

```sql
DROP INDEX ref_uniq;
```

## 2013 Dec 08

```sql
drop view repo_job_vw;
alter table repo_job drop column type;
alter table repo_job drop column prereq_job_succeeded;
alter table repo drop column graph_succeeded;
alter table repo drop column graph_failed;

-- drop and recreate check restraint as:
ALTER TABLE repo_job ADD CONSTRAINT repo_job_outstanding_requests EXCLUDE (repo WITH =) WHERE (ended_at IS NULL);
```

## 2013 Dec 06: add dep table, remove ref.repo_repo_count table, copy from
   ref.repo_repo_count to dep

## 2013 Dec 03: change some ref cache tables
```
drop table ref.client_author_count;
create table ref.author_client; -- TODO(sqs): fill in full schema from db/db.go
```


## 2013 Nov 24: add author_date columns to ref_author and sym_author

```
alter table sym_author add column author_date timestamp without time zone default null;
alter table ref_author add column author_date timestamp without time zone default null;
```
