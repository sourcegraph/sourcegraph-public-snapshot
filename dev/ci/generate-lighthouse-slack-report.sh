#!/usr/bin/env bash

# This script will run Lighthouse audits against a specified URL
# and generate a text report suitable for uploading to Slack.

# Friendly name to associate with the Lighthouse audit
NAME=$1
# URL to run Lighthouse against
URL=$2
# File to output results to
OUTPUT_FILE=$3

yarn lhci collect --url="$URL" --no-lighthouserc --settings.preset="desktop" --numberOfRuns=5

# LHCI doesn't an provide a way to easily expose the temporary storage URL, we have to extract it ourselves
REPORT_URL=$(yarn lhci upload --target=temporary-public-storage | grep -o "https:\/\/storage.googleapis.*.html\+")
# Primary result source, we'll use this to extract the raw audit data.
yarn lhci upload --target=filesystem

# Lighthouse runs multiple times and takes the median to account for varying network latency
REPRESENTATIVE_RUN=$(jq -r '.[] | select(.isRepresentativeRun==true)' manifest.json)

# Extract the Lighthouse score for each relevant category
PERFORMANCE=$(jq -r '.summary.performance' <<<"$REPRESENTATIVE_RUN")
ACCESSIBILITY=$(jq -r '.summary.accessibility' <<<"$REPRESENTATIVE_RUN")
BEST_PRACTICES=$(jq -r '.summary."best-practices"' <<<"$REPRESENTATIVE_RUN")
SEO=$(jq -r '.summary.seo' <<<"$REPRESENTATIVE_RUN")

SUMMARY="
$NAME: <$REPORT_URL|Report>\n
Performance: $PERFORMANCE\n
Accessibility: $ACCESSIBILITY\n
Best practices: $BEST_PRACTICES\n
SEO: $SEO\n\n
"

echo -e "$SUMMARY" >>"$OUTPUT_FILE"
