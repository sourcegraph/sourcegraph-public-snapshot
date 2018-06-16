#!/bin/bash

# The name is intentionally empty ('') so that it forces a merge conflict if two branches attempt to
# create a migration at the same sequence number (because they will both add a file with the same
# name, like `migrations/1528277032_.up.sql`).
unset CDPATH
cd migrations && migrate create -ext sql -dir . -digits 10 -seq ''
echo Empty migration files added in migrations/
