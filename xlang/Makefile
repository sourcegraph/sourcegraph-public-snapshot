.PHONY: all lsp-proxy xlang-go xlang-typescript deploy

all: lsp-proxy xlang-go xlang-typescript

lsp-proxy:
	cmd/lsp-proxy/build.sh

xlang-go:
	golang/cmd/xlang-go/build.sh

xlang-typescript:
	typescript/build.sh

# Update the Docker images used for lsp-proxy or xlang-* on all
# deployment environments (including prod).
deploy: all
	for ns in staging; do echo $$ns:; kubectl --namespace=$$ns get pods -o name | grep -E '(lsp-proxy|xlang-go|xlang-typescript)' | xargs kubectl --namespace=$$ns delete; echo; done
