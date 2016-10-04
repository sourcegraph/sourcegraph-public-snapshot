.PHONY: all cmd/lsp-proxy golang deploy

all: cmd/lsp-proxy golang

cmd/lsp-proxy:
	$(MAKE) -C cmd/lsp-proxy

golang:
	$(MAKE) -C golang/cmd/xlang-go

# Update the Docker images used for lsp-proxy or xlang-go on all
# deployment environments (including prod).
deploy: all
	for ns in staging staging2 staging3 staging4 prod; do echo $$ns:; kubectl --namespace=$$ns get pods -o name | grep -E '(lsp-proxy|xlang-go)' | xargs kubectl --namespace=$$ns delete; echo; done
