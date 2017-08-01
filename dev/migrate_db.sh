if [ -z $PGHOST ]; then echo "error: \$PGHOST not set"; exit 1; fi
if [ -z $PGPORT ]; then echo "error: \$PGPORT not set"; exit 1; fi
if [ -z $PGDATABASE ]; then echo "error: \$PGDATABASE not set"; exit 1; fi
if [ -z $PGUSER ]; then echo "error: \$PGUSER not set"; exit 1; fi

type migrate >/dev/null 2>&1 || { echo >&2 "error: \"migrate\" not installed (see https://github.com/mattes/migrate/tree/master/cli#installation)"; exit 1; }

migrate -path ./migrations -database postgres://$PGHOST:$PGPORT/$PGDATABASE?user=$PGUSER\&password=$PGPASSWORD\&sslmode=$PGSSLMODE up
