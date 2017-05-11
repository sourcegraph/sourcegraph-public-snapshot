MAKEFLAGS+=--no-print-directory

.PHONY: drop-entire-local-database generate serve-dev test test-app

serve-dev:
	@./dev/server.sh

generate:
	@# Ignore app/assets because its output is not checked into Git.
	go list ./... | grep -v /vendor/ | grep -v app/assets | xargs go generate
	cd ui && yarn run generate

drop-entire-local-database:
	psql -c "drop schema public cascade; create schema public;"

# convenience target for dev to run all checks that might be affected by changes to the main app ONLY
APPTESTPKGS ?= $(shell go list ./... | grep -v /vendor/ | grep -v '^sourcegraph.com/sourcegraph/sourcegraph/xlang$$' | sort)
test-app:
	./dev/check/all.sh
	cd ui && yarn test
	CDPATH= cd ui/scripts/tsmapimports && yarn test
	go test -race ${APPTESTPKGS}
