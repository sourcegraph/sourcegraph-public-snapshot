.PHONY: all lsp-proxy xlang-go deploy

all: lsp-proxy xlang-go

lsp-proxy:
	cmd/lsp-proxy/build.sh

xlang-go:
	golang/cmd/xlang-go/build.sh

# Update the Docker images used for lsp-proxy or xlang-go on all
# deployment environments (including prod).
deploy: all
	for ns in staging staging2 staging3 staging4 prod; do echo $$ns:; kubectl --namespace=$$ns get pods -o name | grep -E '(lsp-proxy|xlang-go)' | xargs kubectl --namespace=$$ns delete; echo; done
