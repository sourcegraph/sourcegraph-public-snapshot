#!/bin/bash

# According to the handbook [https://handbook.sourcegraph.com/departments/engineering/dev/process/releases/#when-we-release], we
# release on the 22nd of each month, If the 22nd falls on a non-working day,
# the release captain will shift the release earlier to the last working day before the 22nd.

# ALL dates must be in the format YYYY-MM-DD

# This returns true or false depending on if the date passed is a working day.
# We use this to adjust the release date / freeze date from the usual date defined in
# the handbook.
function is_weekend() {
  # Parse the date string
  local date_string="$1"
  # Get the day of the week (0-6, where Sunday is 0)
  local day_of_week
  day_of_week=$(date -d "$date_string" +%u)
  # Check if the day of the week is 6 (Saturday) or 7 (Sunday)
  if [ "$day_of_week" -eq 6 ] || [ "$day_of_week" -eq 7 ]; then
    echo "true"
  else
    echo "false"
  fi
}

# Use the this to get the last working day before the date passed in as an argument
function find_last_working_day() {
  local date_string=$1
  while $(is_weekend "$date_string"); do
    date_string=$(date -d "$date_string 1 day ago" +%F)
  done
  echo "$date_string"
}

function get_closest_working_day() {
  if [ -z "$1" ]; then
    echo "Error: No date string argument passed. Please pass date in format YYYY-MM-DD" >&2
    return 1
  fi

  local date_to_check=$1
  if $(is_weekend "$date_to_check"); then
    date_to_check=$(find_last_working_day "$date_to_check")
  fi
  echo "$date_to_check"
}

function get_epoch() {
  if [ -z "$1" ]; then
    echo "Error: No date string argument passed. Please pass date in format YYYY-MM-DD" >&2
    return 1
  fi

  echo "$(date -d "$1" +%s)"
}

release_day=22
current_month=$(date +'%m')
current_year=$(date +'%Y')

release_date=$(get_closest_working_day "${current_year}-${current_month}-${release_day}")
cut_date=$(get_closest_working_day "$(date -d "$release_date - 3 days" +%F)")
freeze_date=$(get_closest_working_day "$(date -d "$cut_date - 2 days" +%F)")
todays_date=$(date +'%Y-%m-%d %H:%M')

has_label=$1

todays_date_epoch=$(get_epoch "$todays_date")
freeze_date_epoch=$(get_epoch "$freeze_date 00:00")
release_date_epoch=$(get_epoch "$release_date 23:59")

if [ "$todays_date_epoch" -ge "$freeze_date_epoch" ] && [ "$todays_date_epoch" -lt "$release_date_epoch" ]; then
  if [ "${has_label}" = "true" ]; then
    echo "✅ Label 'i-acknowledge-this-goes-into-the-release' is present"
    exit 0
  else
    echo "❌ Label 'i-acknowledge-this-goes-into-the-release' is absent"
    echo "👉 We're in the next Sourcegraph release code freeze period. If you are 100% sure your changes should get released or provide no risk to the release, add the label your PR with 'i-acknowledge-this-goes-into-the-release'"
    exit 1
  fi
else
  echo "📅 Not enabled, we're not yet on ${freeze_date} and release code freeze has not started yet."
  exit 0
fi
