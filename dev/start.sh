#!/bin/bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

GOBIN="$PWD"/vendor/.bin go get sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/sqs/rego sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/mattn/goreman

export AUTH0_CLIENT_ID=onW9hT0c7biVUqqNNuggQtMLvxUWHWRC
export AUTH0_CLIENT_SECRET=cpse5jYzcduFkQY79eDYXSwI6xVUO0bIvc4BP6WpojdSiEEG6MwGrt8hj_uX3p5a
export AUTH0_DOMAIN=sourcegraph-dev.auth0.com
export AUTH0_MANAGEMENT_API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJSYW1KekRwRmN6SFZZNTBpcmFSb0JMdTNRVmFHTE1VRiIsInNjb3BlcyI6eyJ1c2VycyI6eyJhY3Rpb25zIjpbInJlYWQiLCJ1cGRhdGUiXX0sInVzZXJfaWRwX3Rva2VucyI6eyJhY3Rpb25zIjpbInJlYWQiXX0sInVzZXJzX2FwcF9tZXRhZGF0YSI6eyJhY3Rpb25zIjpbInVwZGF0ZSJdfX0sImlhdCI6MTQ3NzA5NDQxOSwianRpIjoiMTA3YzYyMTZjNWZjYzVjNGNkYjYzZTgxNjRjYjg3ODgifQ.ANOcIGeFPH7X_ppl-AXcv2m0zI7hWwqDlRwJ6h_rMdI
export GITHUB_CLIENT_ID=6f2a43bd8877ff5fd1d5
export GITHUB_CLIENT_SECRET=c5ff37d80e3736924cbbdf2922a50cac31963e43
export LIGHTSTEP_PROJECT=sourcegraph-dev
export LIGHTSTEP_ACCESS_TOKEN=d60b0b2477a7ccb05d7783917f648816
export LIGHTSTEP_INCLUDE_SENSITIVE=true
export PGSSLMODE=disable
export SRC_GITHUB_APP_ID=2534
export SRC_GITHUB_APP_URL=https://github.com/apps/sourcegraph-dev
export SRC_GITHUB_APP_PRIVATE_KEY="$(cat $PWD/dev/github/sourcegraph-dev.private-key.pem)"
export PUBLIC_REPO_REDIRECTS=false
export AUTO_REPO_ADD=true

export SRC_APP_SECRET_KEY=OVSHB1Yru3rlsQ0eKNi2GXCZ47zU7DCK
export GITHUB_BASE_URL=http://127.0.0.1:3180
export SRC_REPOS_DIR=$HOME/.sourcegraph/repos
export DEBUG=true
export SRC_APP_DISABLE_SUPPORT_SERVICES=true
export SRC_GIT_SERVERS=127.0.0.1:3178
export SEARCHER_URL=http://127.0.0.1:3181
export LSP_PROXY=127.0.0.1:4388
export REDIS_MASTER_ENDPOINT=127.0.0.1:6379
export SRC_SESSION_STORE_REDIS=127.0.0.1:6379
export SRC_INDEXER=127.0.0.1:3179
export SRC_SYNTECT_SERVER=http://localhost:3700
export SRC_FRONTEND_INTERNAL=localhost:3090

export SG_FEATURE_SEP20AUTH=true # TODO: Deprecated. Remove.

export PHABRICATOR_URL="http://phabricator.sgdev.org"
export GITOLITE_HOSTS="gitolite.sgdev.org/!git@gitolite.sgdev.org"
export CORS_ORIGIN="https://github.com http://phabricator.sgdev.org"
export PHABRICATOR_CONFIG='[{"url":"http://phabricator.sgdev.org","token":"api-agswx2nwodkweitoo3t5l4dcc5xu"}]'
export GITHUB_CONFIG='[{"url": "https://35.197.32.22", "token":"23993bbf8e0fee068b8f70db05fc445d5a7a83da", "certificate": "-----BEGIN CERTIFICATE-----\nMIIFTjCCAzagAwIBAgIJAOnZ9BB0S2aeMA0GCSqGSIb3DQEBCwUAMBcxFTATBgNV\nBAMMDDM1LjE5Ny4zMi4yMjAeFw0xNzEyMDgwNjExNDlaFw0yNzEyMDYwNjExNDla\nMBcxFTATBgNVBAMMDDM1LjE5Ny4zMi4yMjCCAiIwDQYJKoZIhvcNAQEBBQADggIP\nADCCAgoCggIBAMsVpIFFcRydUcwI0Hetxe0MGYI/e6PgzyB+Wt2WNOJuRKnk3ahm\nwQHWKZmvjTVCaUmdOD/jIQn1OtJCFero+l200Ks/SZGtLoxuaWB6j5TapN1WGzu2\nFZdWkJgcDQTHY+URcs3RSMKWPkZ36FzLbXq3pg6AG6f/5leIkplXGWXwJu0kobOz\nkeC6SlxLf7JvItl9yjo1CLRHuFed+GasyNfno51DrilsGDm2E4C3R7vT6M2/PnUy\n+Fr4i42l5mWqfKGNzxN4xEpGXyOYU3SHdyvcPmyzWWHn9r+MhK4SuvAlqpdAj/5q\nP6mOMhsSMXt/npP8KNYjf2tKD7FV/DkM5I1HgzG3d9ZNeLK1UUFXtwe+gFKckeu4\nRSbCHxhyn6YREA63P7ZSiOJG41DsQUOr+lAOsQqX3wC9NMahX2RhPdpw8LfxW6Kw\nyPyBKSQbKEdIFp/NC4elpoRhBn3hlqVyQJU6t0bHXl2ni7Biur22leHTkZbyguxe\nzM5z0ysKcFZ7eiJZ4aNlrk9t0hy1d91PQWn97AkpSiYpBRUJiptxQR5dXUNqWKcY\nFG95laqG2CegR6L9qSmo226DgvkO+E6f1AhxHUCJXdMmAb7824kF+Q6H6kqF97U9\n+CQxbHLqzDlaZdyoMitnS6aRABouuzM6AB46At3T0mLECd2Y88cy3U0vAgMBAAGj\ngZwwgZkwHQYDVR0OBBYEFKTny+mi+pz85rlaNOJf6UixMWgHMB8GA1UdIwQYMBaA\nFKTny+mi+pz85rlaNOJf6UixMWgHMAwGA1UdEwQFMAMBAf8wSQYDVR0RBEIwQIcE\nI8UgFoIMMzUuMTk3LjMyLjIyghMzNS4xOTcuMzIuMjIueGlwLmlvghUqLjM1LjE5\nNy4zMi4yMi54aXAuaW8wDQYJKoZIhvcNAQELBQADggIBAAWCcbnr0u76zBlIrvei\nqqzbLavrkjDw9pKh+4Po01dUPH7BzekgYulOyEaG6rmXk2vdaGLXVIViNKRLVDr5\nUJcwraawELb2SBDRCSj8XnEek6500USZ5jfQZ/neYZ7+yQvKj1aPJ+TWOJ2lqkDU\nQXIt4kF/eag5qqW/eLdO7Z0SulawowXV0AFyysuHf1iFzqtb4p8v5t8NFlkH/Wtz\nEHeXnBioMQAs/xzTVOKlfBLaz7Y+/fAzDmmZpF730Ot5ay3LFUlGF/lHqNByCIsI\n8mLwZaP0sCUsk7kejJEKEa7CdtkoAkzIWS0NeeUUgyipWkg0orB1EyE1RrxUbong\netAKq1lO71tK1XD5jkCs1IAW1ZLUu7NHczNLJGqIMSMlPvcdtS8MRhiQB3sP9ybz\nG3HmZrCEqT9dVR1oqOgxSfayP9r74+5WUfmWX7zItEc3/6MUNov2u3BDNFNQiLK1\nt+vEgwhf5AEIbO2TnE6rSRfvB3cc3zPxzugxKJuvNBV7jCW1ImKGmEqTEEFjb7fK\nJVVyFAgGihepSmjqS2mj0YKtBYhygH5xcfbbb/ukLbbf9imwu88LqUMfJyYELS3u\ni2NSpa3qDQjdzY09gtdMnKNr6WZahEoeBAYTY4D81apWNZnnSQBh5vXNA9S80b7J\n46mtbylQW07gfvR4NXMnSU+B\n-----END CERTIFICATE-----"}]'

export LANGSERVER_GO=${LANGSERVER_GO-"tcp://localhost:4389"}
export LANGSERVER_GO_BG=${LANGSERVER_GO_BG-"tcp://localhost:4389"}

export LICENSE_KEY=${LICENSE_KEY:-24348deeb9916a070914b5617a9a4e2c7bec0d313ca6ae11545ef034c7138d4d8710cddac80980b00426fb44830263268f028c9735}

# WebApp
export NODE_ENV=development

mkdir -p .bin
env GOBIN=$PWD/.bin go install -tags="dev" -v sourcegraph.com/sourcegraph/sourcegraph/cmd/{gitserver,indexer,github-proxy,xlang-go,lsp-proxy,searcher}

# Increase ulimit (not needed on Windows/WSL)
type ulimit > /dev/null && ulimit -n 10000 || true

exec "$PWD"/vendor/.bin/goreman -f dev/Procfile start
