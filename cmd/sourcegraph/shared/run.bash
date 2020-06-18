#!/bin/bash

set -eu

(cd web && NODE_ENV=production DISABLE_TYPECHECKING=true yarn run build)

go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
go install -buildmode exe -tags dist ./cmd/sourcegraph
PGDATABASE=cmd_sourcegraph_shared_test sourcegraph
