#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates
