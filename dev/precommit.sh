#!/usr/bin/env bash

# Run linters depending on what changed.

unset CDPATH

if git diff --cached --name-only | grep --quiet '.scss$'
then
	npm --prefix web run stylelint || exit 1
fi

if ! git diff --quiet --cached --name-only cmd/frontend/internal/graphqlbackend/schema.graphql
then
	npm --prefix web graphql || exit 1
	# TODO(sqs): ensure we regenerated schema.go
fi

if git diff --cached --name-only | grep --quiet '.tsx\?$'
then
	# Run tslint only on changed files, for faster execution.
	files=$(git diff --cached --name-only | grep '.tsx\?$')
	web/node_modules/.bin/tslint -c web/tslint.json -p web/tsconfig.json -e 'web/node_modules/**' $files || exit 1

	(cd web && node_modules/.bin/tsc --noEmit -p .) || exit 1
fi

if git diff --cached --name-only | grep --quiet '.go$'
then
	go test -run='^$' ./... || exit 1
fi
