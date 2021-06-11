#!/usr/bin/env bash

readonly files=$(grep -rn --include \*.go "github.com/pkg/errors")

readonly count=$(echo files | wc -l)

if [[ "${count}" -gt 0 ]]; then
  echo 'ERROR: go-error-package check failed. Do not use "github.com/pkg/errors". Use "github.com/cockroachdb/errors".'
  echo 'Files using "github.com/pkg/errors" are:'
  echo "${files}"
  exit 1
fi

echo "Success: go-error-package check passed."
