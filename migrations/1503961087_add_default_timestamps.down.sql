ALTER TABLE threads
    ALTER COLUMN "created_at" DROP DEFAULT,
    ALTER COLUMN "updated_at" DROP DEFAULT;

ALTER TABLE comments
    ALTER COLUMN "created_at" DROP DEFAULT,
    ALTER COLUMN "updated_at" DROP DEFAULT;

ALTER TABLE local_repos
    ALTER COLUMN "created_at" DROP DEFAULT,
    ALTER COLUMN "updated_at" DROP DEFAULT;
