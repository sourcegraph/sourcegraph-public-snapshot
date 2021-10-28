#!/usr/bin/env bash

# Readable name for Lighthouse audit
NAME=$1
# URL to run Lighthouse against
URL=$2
# File to output results to
OUTPUT_FILE=$3

yarn lhci collect --url=$URL --no-lighthouserc --settings.preset="desktop" --numberOfRuns=5
REPORT_URL=$(yarn lhci upload --target=temporary-public-storage | grep -o "https:\/\/storage.googleapis.*.html\+")
yarn lhci upload --target=filesystem

# Lighthouse runs multiple times and takes the median to account for varying network latency
# We remove the PWA category because it is not currently relevant to our application.
REPRESENTATIVE_RUN=$(jq -r '.[] | select(.isRepresentativeRun==true) | del(.summary.pwa)' manifest.json)

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

echo -e $SUMMARY >>$OUTPUT_FILE
