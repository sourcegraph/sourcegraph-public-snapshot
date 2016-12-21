.PHONY: all lsp-proxy xlang-go xlang-javascript-typescript xlang-ctags deploy deploy-staging

all: lsp-proxy xlang-go xlang-javascript-typescript xlang-ctags xlang-php

lsp-proxy:
	cmd/lsp-proxy/build.sh

xlang-go:
	golang/cmd/xlang-go/build.sh

xlang-javascript-typescript:
	javascript-typescript/build.sh

xlang-ctags:
	ctags/cmd/langserver-ctags/build.sh

xlang-php:
	php/build.sh

deploy:
	@echo "This is deprecated. To update staging environments use the deploy-staging target."
	@echo "Otherwise, just use:"
	@echo
	@echo "  git push origin -f master:docker-images/xlang"
	@echo

# Update the Docker images used for lsp-proxy or xlang-* on all staging
# environments.
deploy-staging: all
	for ns in staging staging2 staging3 staging4; do echo $$ns:; kubectl --namespace=$$ns get pods -o name | grep -E '(lsp-proxy|xlang-go|xlang-typescript|xlang-ctags)' | xargs kubectl --namespace=$$ns delete; echo; done
