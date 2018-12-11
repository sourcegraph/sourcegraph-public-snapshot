CREATE TYPE critical_or_site AS ENUM ('critical', 'site');
CREATE TABLE critical_and_site_config (
	"id" SERIAL NOT NULL PRIMARY KEY,
    "type" critical_or_site NOT NULL,
	"contents" text NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX "critical_and_site_config_unique" ON critical_and_site_config(id, type);
