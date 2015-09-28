install: ${GOBIN}/makex

${GOBIN}/makex: $(shell find -type f -and -name '*.go')
	go install ./cmd/makex
