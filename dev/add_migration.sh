#!/bin/bash

if [ -z "$1" ]; then echo "error: name arg required (e.g. ./add_migration.sh \"add new table\")"; exit 1; fi

NAME="${1// /_}"

migrate create -ext sql -dir ./migrations/ $NAME
