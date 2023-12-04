#!/usr/bin/env bash

# Prepares and creates a standalone docker image for demoing the prototype
pnpm install
pnpm run -w generate
pnpm run build
docker build . -t fkling/web-sveltekit
