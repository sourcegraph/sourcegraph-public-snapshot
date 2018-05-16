#!/bin/bash

set -e
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

eval $(grep 'export OVERRIDE_AUTH_SECRET=' dev/start.sh)

username_to_delete="$1"

user_id_to_delete=$(
    curl -sS \
        -H "X-Override-Auth-Secret: $OVERRIDE_AUTH_SECRET" \
        -H 'X-Override-Auth-Username: dev' \
        -H 'Content-Type: application/json; charset=utf-8' \
        -XPOST \
        -d '{"query":"query($username: String!) { user(username: $username) { id } }","variables":{"username":"'"$username_to_delete"'"}}' \
        http://localhost:3080/.api/graphql \
        | jq -c -r '.data.user.id'
                 )

if [ "$user_id_to_delete" = "null" ]; then
    echo User not found: $username_to_delete
    exit 1
fi

curl -sS \
    -H "X-Override-Auth-Secret: $OVERRIDE_AUTH_SECRET" \
    -H 'X-Override-Auth-Username: dev' \
    -H 'Content-Type: application/json; charset=utf-8' \
    -XPOST \
    -d '{"query":"mutation($user: ID!) { deleteUser(user: $user) { alwaysNil } }","variables":{"user":"'"$user_id_to_delete"'"}}' \
    http://localhost:3080/.api/graphql
echo
