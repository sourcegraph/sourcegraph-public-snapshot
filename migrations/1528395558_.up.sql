CREATE TABLE site_configuration_files (
	"id" SERIAL NOT NULL PRIMARY KEY,
    "contents" text,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX "site_configuration_files_unique" ON site_configuration_files(id);

CREATE TABLE core_configuration_files (
	"id" SERIAL NOT NULL PRIMARY KEY,
    "contents" text,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX "core_configuration_files_unique" ON core_configuration_files(id);
