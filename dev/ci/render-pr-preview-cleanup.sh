#!/usr/bin/env bash

set -e

# Env variables:
# - RENDER_COM_API_KEY (required)
# - RENDER_COM_OWNER_ID (required)

print_usage() {
  echo "Usage: [-e EXPIRE_AFTER_DAYS]" 1>&2
  echo "-e: Number of days since the last time PR previews were updated (defaults: 5 days)" 1>&2
}

urlencode() {
  echo "$1" | curl -Gso /dev/null -w "%{url_effective}" --data-urlencode @- "" | cut -c 3- | sed -e 's/%0A//'
}

get_days_ago_in_ISO() {
  unameOut="$(uname -s)"
  case "${unameOut}" in
  Linux*) machine=Linux ;;
  Darwin*) machine=Mac ;;
  *) ;;
  esac

  # `date` in macos is different from Linux
  if [[ $machine = "Mac" ]]; then
    date -u "-v-$1d" +"%Y-%m-%dT%H:%M:%SZ"
  else
    date -d "$1 days ago" +"%Y-%m-%dT%H:%M:%SZ"
  fi
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

expiration_date_ISO=$(get_days_ago_in_ISO "$expire_after_days")

if [[ -z "${render_api_key}" || -z "${render_owner_id}" ]]; then
  echo "RENDER_COM_API_KEY or RENDER_COM_OWNER_ID is not set"
  exit 1
fi

# Get id list of services which are created before expiration date
# We should take care about render.com rate limit (https://api-docs.render.com/reference/rate-limiting)
# GET 100/minute
# DELETE 30/minute

cursor=""
ids=()

# Collect ids of all services which are updated before expiration date and `not_suspended`
# We use `updatedBefore` since the services are updated on deployments
for (( ; ; )); do
  # render.com API > List services: https://api-docs.render.com/reference/get-services
  service_list=$(curl -sSf --request GET \
    --url "https://api.render.com/v1/services?type=web_service&updatedBefore=$(urlencode "$expiration_date_ISO")&suspended=not_suspended&ownerId=${render_owner_id}&limit=100&cursor=${cursor}" \
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
  echo "Deleting service: $service_id"

  curl -sSf -o /dev/null --request DELETE \
    --url "https://api.render.com/v1/services/${service_id}" \
    --header 'Accept: application/json' \
    --header "Authorization: Bearer ${render_api_key}"

  # To make sure we don't reach the limitation of 30/minute DELETE requests
  sleep 2
done
