#!/usr/bin/env bash

set -e
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

eval "$(grep 'export OVERRIDE_AUTH_SECRET=' dev/start.sh)"

cd ./dev/api

username_to_delete="$1"

user_id_to_delete_data=$(jq --arg username "$username_to_delete" '.variables={"username":$username}' <user_id_query.json)

user_id_to_delete=$(
  curl -sS \
    -H "X-Override-Auth-Secret: $OVERRIDE_AUTH_SECRET" \
    -H 'X-Override-Auth-Username: dev' \
    -H 'Content-Type: application/json; charset=utf-8' \
    -XPOST \
    -d "$user_id_to_delete_data" \
    http://localhost:3080/.api/graphql |
    jq -c -r '.data.user.id'
)

if [ "$user_id_to_delete" = "null" ]; then
  echo User not found: "$username_to_delete"
  exit 1
fi

delete_user_data=$(jq --arg user "$user_id_to_delete" '.variables={"user":$user}' <delete_user_mutation.json)

curl -sS \
  -H "X-Override-Auth-Secret: $OVERRIDE_AUTH_SECRET" \
  -H 'X-Override-Auth-Username: dev' \
  -H 'Content-Type: application/json; charset=utf-8' \
  -XPOST \
  -d "$delete_user_data" \
  http://localhost:3080/.api/graphql
echo
