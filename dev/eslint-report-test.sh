#!/usr/bin/env bash

# The relative path to the eslint report
ESLINT_REPORT="$1"

# Ensure that the eslint report exists
if [ ! -f "$ESLINT_REPORT" ]; then
  echo "${ESLINT_REPORT} does not exist."
  exit 1
fi

# Check if the eslint report is empty
if [ -s "$ESLINT_REPORT" ]; then
  # Get the absolute path to the eslint report.
  absolute_report_path="$(realpath "$ESLINT_REPORT")"

  # Remove everything before "__main__/" from the absolute path to the report.
  workspace_report_path="${absolute_report_path#*__main__/}"

  # Print the relative report path.
  echo "ESLint report: $workspace_report_path"

  cat "$ESLINT_REPORT"
  exit 1
else
  echo "No ESLint issues found."
  exit 0
fi
