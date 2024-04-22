#!/usr/bin/env bash

cleanup() {
  src orgs members remove -org-id="$(src org get -f '{{.ID}}' -name=abc-org)" -user-id="$(src users get -f '{{.ID}}' -username=alice)"
  src orgs delete -id="$(src orgs get -f '{{.ID}}' -name=abc-org)"
  src users delete -id="$(src users get -f='{{.ID}}' -username=alice)"
}

cleanup >/dev/null 2>&1

set -euf -o pipefail

unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/../.." # cd to repo root dir

go install ./cmd/src

src users create -username=alice -email=alice@example.com
src orgs create -name=abc-org
src orgs members add -org-id="$(src org get -f '{{.ID}}' -name=abc-org)" -username=alice
cleanup
