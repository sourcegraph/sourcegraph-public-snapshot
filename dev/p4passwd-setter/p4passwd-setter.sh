#!/bin/bash

# Print each command before executing it; exit in case of an error
set -ex

# Name arguments
code_host_id=$1
password=$2

# Fetch and decode data
# shellcheck disable=SC2016
query='query GetCodeHostInfo($id:ID!){node(id:$id){...on ExternalService{config}}}'
response=$(echo "$query" | src api "id=$code_host_id")
# If response contains "error", show error message
if [[ $response == *"error"* ]]; then
    error_message=$(echo "$response" | jq .errors[0].message)
    echo "Error: $error_message"
    exit 1
fi

# Get config
json=$(echo "$response" | jq .data.node.config)

# Update data
updatedJson="${json/REDACTED/$password}"
mutation=("mutation UpdateCodeHostPassword(\$input:UpdateExternalServiceInput!){updateExternalService(input:\$input){id}}")
mutation_vars=("{\"input\":{\"id\":\"$code_host_id\",\"config\":$updatedJson}}")
response=$(src api -query "${mutation[@]}" -vars "${mutation_vars[@]}")

# If response contains "error", show error message
if [[ $response == *"error"* ]]; then
    error_message=$(echo "$response" | jq .errors[0].message)
    echo "Error: $error_message"
    exit 1
fi

echo "Password for code host updated."
