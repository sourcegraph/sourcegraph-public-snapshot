#!/usr/bin/env bash

go generate github.com/sourcegraph/sourcegraph/migrations github.com/sourcegraph/sourcegraph/cmd/frontend/db
