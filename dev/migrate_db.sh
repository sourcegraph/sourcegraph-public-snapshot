#!/bin/sh

type migrate >/dev/null 2>&1 || { echo >&2 "error: \"migrate\" not installed (see https://github.com/mattes/migrate/tree/master/cli#installation)"; exit 1; }

migrate -path ./migrations -database postgres://$PGHOST:$PGPORT/$PGDATABASE?user=$PGUSER\&password=$PGPASSWORD\&sslmode=$PGSSLMODE up
