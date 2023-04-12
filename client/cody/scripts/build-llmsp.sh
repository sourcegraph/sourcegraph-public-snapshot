#!/bin/bash

VERSION="v1.0.0"

run() {
        LLMSP_DIR="$(dirname "$(readlink -f "$0")")/../resources/bin"
        mkdir -p "${LLMSP_DIR}"
	git clone "git@github.com:pjlast/llmsp" && cd llmsp && \
    env GOOS=windows GOARCH=amd64 go build -o "${LLMSP_DIR}/llmsp-v1.0.0-amd64-windows.exe" && \
    env GOOS=darwin GOARCH=amd64 go build -o "${LLMSP_DIR}/llmsp-v1.0.0-amd64-darwin" && \
    env GOOS=darwin GOARCH=arm64 go build -o "${LLMSP_DIR}/llmsp-v1.0.0-arm64-darwin" && \
    env GOOS=linux GOARCH=amd64 go build -o "${LLMSP_DIR}/llmsp-v1.0.0-amd64-linux" && \
    env GOOS=linux GOARCH=arm64 go build -o "${LLMSP_DIR}/llmsp-v1.0.0-arm64-linux"

  pushd "${LLMSP_DIR}" || return
	trap 'popd' EXIT
}

run
