MAKEFLAGS+=--no-print-directory

.PHONY: drop-entire-local-database generate serve-dev test test-app

serve-dev:
	@./dev/server.sh

generate:
	@# Ignore app/assets because its output is not checked into Git.
	go list ./... | grep -v /vendor/ | grep -v app/assets | xargs go generate

drop-entire-local-database:
	psql -c "drop schema public cascade; create schema public;"
