package database

// $PGHOST, $PGUSER, $PGPORT etc. must be set to run this generate script.
//go:generate env GO111MODULE=on go run ../../dev/schemadoc/main.go
//go:generate env GO111MODULE=on go run ../../dev/sg migration squash-all -db frontend -f ../../migrations/frontend/squashed.sql
//go:generate env GO111MODULE=on go run ../../dev/sg migration squash-all -db codeintel -f ../../migrations/codeintel/squashed.sql
//go:generate env GO111MODULE=on go run ../../dev/sg migration squash-all -db codeinsights -f ../../migrations/codeinsights/squashed.sql
