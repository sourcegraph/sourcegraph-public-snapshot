#!/bin/bash

set -e
cd $(dirname "${BASH_SOURCE[0]}")/..

if [ ! -d ../sourcegraph ]; then
    echo "OSS repo not found at ../sourcegraph"
    exit 1
fi

echo "Installing web dependencies..."
yarn

echo "Linking OSS webapp to node_modules"
pushd ../sourcegraph
yarn link
popd
yarn link @sourcegraph/webapp

# Generated with:
#
#   go run ./pkg/license/generate-license.go -private-key key.pem -plan='For development use only' -max-user-count=50
#
# See the docs in generate-license.go to obtain key.pem.
export SOURCEGRAPH_LICENSE_KEY=eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJRcW5DdVd5d0hhbnlRM0hxTFk2UEZaS2dTa0hwdlUvK3hGa3RRcldiREFRVERnTDZiaU45SEh5QVNseGpabmdKdTB5NXNGbTRKR3MxRk1qVVVxYjVRUTNwYy9wNW5MNm5ILzFMZFFGTUdUNVhrVzFNODB4bFJEYlVZNVUzclkyblV3aU5maitFeFl4bzZhc0FVUFlHaDdNc3JMZFdFZUJpRXpJb1ZZMytaWS9tL1hSYTQ2SkpiWURLaE5JVnZHMnQ2aEJXek5ha2R3dzAwdW1WWEd1QndNWEd6WUl2cUpOaW1SZ2xvZWVBaFhaSzZUVGJ6OGZDUkFMT0NPQ3NYWGVyQXY4N2xYVllKaGhIaGRrRElsU25BaEZOWXcxZ0FmdUdYTk43MHZqbkxLU3h3Z05QOGVad2o5ZDRnTklKMGU3MloweStDU29IK2Z2WVNqTzFNTUlLUkE9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk5EWXNNalUwTERNMExEWXpMREkwT0N3eE1qZ3NPVGtzTkRWZExDSndJam9pUm05eUlHUmxkbVZzYjNCdFpXNTBJSFZ6WlNCdmJteDVJaXdpZFdNaU9qVXdMQ0psZUhBaU9tNTFiR3g5In0

export SAML_ONELOGIN_CERT=$(cat dev/auth-provider/config/external/client-onelogin-saml-dev-736334.cert.pem)
export SAML_ONELOGIN_KEY=$(cat dev/auth-provider/config/external/client-onelogin-saml-dev-736334.key.pem)

SOURCEGRAPH_CONFIG_FILE=$PWD/dev/config.json GOMOD_ROOT=$PWD ENTERPRISE_COMMANDS="frontend" PROCFILE=./dev/Procfile ./dev/launch.sh
