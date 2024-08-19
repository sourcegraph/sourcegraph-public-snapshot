#!/usr/bin/env sh

GOLLY_DOTCOM_ACCESS_TOKEN="$(gcloud secrets versions access latest --secret CODY_PRO_ACCESS_TOKEN --project cody-agent-tokens --quiet)"
export GOLLY_DOTCOM_ACCESS_TOKEN
GOLLY_S2_ACCESS_TOKEN="$(gcloud secrets versions access latest --secret CODY_S2_ACCESS_TOKEN --project cody-agent-tokens --quiet)"
export GOLLY_S2_ACCESS_TOKEN
