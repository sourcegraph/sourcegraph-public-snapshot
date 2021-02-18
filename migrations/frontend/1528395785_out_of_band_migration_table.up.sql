BEGIN;

CREATE TABLE out_of_band_migrations (
    id               serial primary key,
    team             text not null,
    component        text not null,
    description      text not null,
    introduced       text not null,
    deprecated       text,
    progress         float default 0 not null,
    created          timestamp with time zone not null default now(),
    last_updated     timestamp with time zone,
    non_destructive  boolean not null,
    apply_reverse    boolean default false not null
);

COMMENT ON TABLE out_of_band_migrations IS 'Stores metadata and progress about an out-of-band migration routine.';
COMMENT ON COLUMN out_of_band_migrations.id IS 'A globally unique primary key for this migration. The same key is used consistently across all Sourcegraph instances for the same migration.';
COMMENT ON COLUMN out_of_band_migrations.team IS 'The name of the engineering team responsible for the migration.';
COMMENT ON COLUMN out_of_band_migrations.component IS 'The name of the component undergoing a migration.';
COMMENT ON COLUMN out_of_band_migrations.description IS 'A brief description about the migration.';
COMMENT ON COLUMN out_of_band_migrations.introduced IS 'The Sourcegraph version in which this migration was first introduced.';
COMMENT ON COLUMN out_of_band_migrations.deprecated IS 'The lowest Sourcegraph version that assumes the migration has completed.';
COMMENT ON COLUMN out_of_band_migrations.progress IS 'The percentage progress in the up direction (0=0%, 1=100%).';
COMMENT ON COLUMN out_of_band_migrations.created IS 'The date and time the migration was inserted into the database (via an upgrade).';
COMMENT ON COLUMN out_of_band_migrations.last_updated IS 'The date and time the migration was last updated.';
COMMENT ON COLUMN out_of_band_migrations.non_destructive IS 'Whether or not this migration alters data so it can no longer be read by the previous Sourcegraph instance.';
COMMENT ON COLUMN out_of_band_migrations.apply_reverse IS 'Whether this migration should run in the opposite direction (to support an upcoming downgrade).';

ALTER TABLE out_of_band_migrations ADD CONSTRAINT out_of_band_migrations_team_nonempty CHECK (team <> '');
ALTER TABLE out_of_band_migrations ADD CONSTRAINT out_of_band_migrations_component_nonempty CHECK (component <> '');
ALTER TABLE out_of_band_migrations ADD CONSTRAINT out_of_band_migrations_description_nonempty CHECK (description <> '');
ALTER TABLE out_of_band_migrations ADD CONSTRAINT out_of_band_migrations_introduced_valid_version CHECK (introduced ~ '^(\d+)\.(\d+)\.(\d+)$'::text);
ALTER TABLE out_of_band_migrations ADD CONSTRAINT out_of_band_migrations_deprecated_valid_version CHECK (deprecated ~ '^(\d+)\.(\d+)\.(\d+)$'::text);
ALTER TABLE out_of_band_migrations ADD CONSTRAINT out_of_band_migrations_progress_range CHECK (progress >= 0 AND progress <= 1);

CREATE TABLE out_of_band_migrations_errors (
    id            serial primary key,
    migration_id  int not null,
    message       text not null,
    created       timestamp with time zone not null default now()
);

ALTER TABLE out_of_band_migrations_errors ADD FOREIGN KEY (migration_id) REFERENCES out_of_band_migrations(id) ON DELETE CASCADE;
ALTER TABLE out_of_band_migrations_errors ADD CONSTRAINT out_of_band_migrations_errors_message_nonempty CHECK (message <> '');

COMMENT ON TABLE out_of_band_migrations_errors IS 'Stores errors that occurred while performing an out-of-band migration.';
COMMENT ON COLUMN out_of_band_migrations_errors.id IS 'A unique identifer.';
COMMENT ON COLUMN out_of_band_migrations_errors.migration_id IS 'The identifier of the migration.';
COMMENT ON COLUMN out_of_band_migrations_errors.message IS 'The error message.';
COMMENT ON COLUMN out_of_band_migrations_errors.created IS 'The date and time the error occurred.';

COMMIT;
