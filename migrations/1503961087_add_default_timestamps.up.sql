ALTER TABLE threads
    ALTER COLUMN "created_at" SET DEFAULT now(),
    ALTER COLUMN "updated_at" SET DEFAULT now();

ALTER TABLE comments
    ALTER COLUMN "created_at" SET DEFAULT now(),
    ALTER COLUMN "updated_at" SET DEFAULT now();

ALTER TABLE local_repos
    ALTER COLUMN "created_at" SET DEFAULT now(),
    ALTER COLUMN "updated_at" SET DEFAULT now();
