#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -ex

psql -d sourcegraph-test-db -c 'drop schema public cascade; create schema public;'
