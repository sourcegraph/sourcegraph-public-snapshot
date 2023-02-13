#!/usr/bin/env bash

# This script builds the svelte docker image.

pnpm install
pnpm -w generate
pnpm build

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

IMAGE="us-central1-docker.pkg.dev/sourcegraph-dogfood/svelte/web"

echo "--- docker build web-svelte $(pwd)"

docker build -f client/web-sveltekit/Dockerfile --build-arg PROJECT_ROOT=./client/web-sveltekit -t "$IMAGE" "$(pwd)" \

docker push $IMAGE
