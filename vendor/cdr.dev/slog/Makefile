all: fmt lint test

.SILENT:

.PHONY: *

.ONESHELL:
SHELL = bash
.SHELLFLAGS = -ceuo pipefail

include ci/fmt.mk
include ci/lint.mk
include ci/test.mk
