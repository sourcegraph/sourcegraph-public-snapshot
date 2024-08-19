#!/usr/bin/env bash

# Against a dev environment will collect and merge code paths that are missing
# tenant

exec go tool pprof -http :10810 http://localhost:{6063,6069,6074,6089,3551,3552}/debug/pprof/missing_tenant
