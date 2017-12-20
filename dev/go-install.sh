#!/bin/bash

go install -tags="dev" -v sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,github-proxy,xlang-go,lsp-proxy,searcher,frontend}
