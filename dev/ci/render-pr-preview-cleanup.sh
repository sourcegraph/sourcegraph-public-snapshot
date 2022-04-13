#!/usr/bin/env bash

set -e

# Env variables:
# - RENDER_COM_API_KEY (required)
# - RENDER_COM_OWNER_ID (required)
# - BUILDKITE_PULL_REQUEST_REPO (optional)

print_usage() {
  echo "Usage: [-e EXPIRE_AFTER_DAYS]" 1>&2
  echo "-e: Number of days since the last time PR previews were updated (defaults: 5 days)" 1>&2
}

urlencode() {
  echo "$1" | curl -Gso /dev/null -w "%{url_effective}" --data-urlencode @- "" | cut -c 3- | sed -e 's/%0A//'
}

expire_after_days="5"

while getopts 'e:' flag; do
  case "${flag}" in
  e) expire_after_days="${OPTARG}" ;;
  *)
    print_usage
    exit 1
    ;;
  esac
done

render_api_key="${RENDER_COM_API_KEY}"
render_owner_id="${RENDER_COM_OWNER_ID}"

expiration_date_UNIX=$(date -u "-v-${expire_after_days}d" +%s)
# expiration_date ISO format
expiration_date_ISO=$(date -u "-v-${expire_after_days}d" +"%Y-%m-%dT%H:%M:%SZ")

if [[ -z "${render_api_key}" || -z "${render_owner_id}" ]]; then
  echo "RENDER_COM_API_KEY or RENDER_COM_OWNER_ID is not set"
  exit 1
fi

# Get id list of services which are created before $expiration_date_ISO
# We should take care about render.com rate limit (https://api-docs.render.com/reference/rate-limiting)
# GET 100/minute
# DELETE 30/minute

cursor=""
ids=()

# Collect ids of all services which are created before `expiration_date_ISO` and `not_suspended`
for (( ; ; )); do
  service_list=$(curl -sSf --request GET \
    --url "https://api.render.com/v1/services?type=web_service&createdBefore=$(urlencode "$expiration_date_ISO")&suspended=not_suspended&ownerId=${render_owner_id}&limit=100&cursor=${cursor}" \
    --header 'Accept: application/json' \
    --header "Authorization: Bearer ${render_api_key}")

  num_of_records=$(echo "$service_list" | jq -r '. | length')

  if [ "${num_of_records}" == "0" ]; then
    break
  fi

  while IFS='' read -r line; do
    ids+=("$line")
  done < <(echo "$service_list" | jq -r '.[].service.id')

  cursor=$(echo "$service_list" | jq -r '.[-1].cursor')
done

for service_id in "${ids[@]}"; do
  echo "Checking service: $service_id"

  # Get the last deploy time
  last_deploy_created_at=$(curl -sSf --request GET \
    --url "https://api.render.com/v1/services/${service_id}/deploys?limit=1" \
    --header 'Accept: application/json' \
    --header "Authorization: Bearer ${render_api_key}" | jq -r '.[].deploy.createdAt')

  # Remove nanosecond part out of ISO Datetime format of `createdAt` then get the UNIX timestamp
  last_deploy_created_at_UNIX=$(date -jf "%Y-%m-%dT%H:%M:%SZ" "${last_deploy_created_at//.[[:digit:]]*Z/Z}" +%s)

  if [[ $last_deploy_created_at_UNIX -lt $expiration_date_UNIX ]]; then
    # there are no deploy times since `expiration_date_ISO` => remove it
    echo "-- Removing service ${service_id} (last deployed at ${last_deploy_created_at})"

    # curl -sSf -o /dev/null --request DELETE \
    #   --url "https://api.render.com/v1/services/${service_id}" \
    #   --header 'Accept: application/json' \
    #   --header "Authorization: Bearer ${render_api_key}"

    # To make sure we don't reach the limitation of 30/minute DELETE requests
    sleep 2
  fi
done
