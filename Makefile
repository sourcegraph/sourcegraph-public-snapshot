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

.PHONY: upgrade-zap
upgrade-zap:
	cd ui && yarn add libzap vscode-zap
	git config url."git@github.com:".insteadOf "https://github.com/"
	govendor fetch github.com/sourcegraph/zap/...

.PHONY: develop-zap
develop-zap:
	cd ui && yarn link libzap vscode-zap
	rm -rf vendor/github.com/sourcegraph/zap
	ln -s $(shell go list -f '{{.Dir}}' github.com/sourcegraph/zap) vendor/github.com/sourcegraph/zap

.PHONY: undevelop-zap
undevelop-zap:
	cd ui && yarn unlink libzap vscode-zap
	git checkout -- vendor/github.com/sourcegraph/zap

.PHONY: deploy-zap
deploy-zap:
	git fetch && git push origin origin/master:docker-images/zap
