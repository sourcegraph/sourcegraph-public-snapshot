#!/usr/bin/env bash
# Set this to fail on the install
set -euxo pipefail

# Install and run the plugin for checkov
# Use the full path to run pip3.10
pip3 install checkov

# List of checks we do not want to run here
# This is a living list and will see additions and mostly removals over time.
SKIP_CHECKS="CKV_GCP_22,CKV_GCP_66,CKV_GCP_13,CKV_GCP_71,CKV_GCP_61,CKV_GCP_21,CKV_GCP_65,CKV_GCP_67,CKV_GCP_20,CKV_GCP_69,CKV_GCP_12,CKV_GCP_24,CKV_GCP_25,CKV_GCP_64,CKV_GCP_68,CKV2_AWS_5,CKV2_GCP_3,CKV2_GCP_5,CKV_AWS_23,CKV_GCP_70,CKV_GCP_62,CKV_GCP_62,CKV_GCP_62,CKV_GCP_62,CKV_GCP_29,CKV_GCP_39"

set +x
# In case no terraform code is present
echo "--- Starting Checkov..."
echo "Note: If there is no output below here then no terraform code was found to scan.  All good!"
echo "==========================================================================================="

# Set not to fail on non-zero exit code
set +e
# Run checkov
python3 -m checkov.main --skip-check $SKIP_CHECKS --quiet --framework terraform --compact -d .

# Options
# --quiet: Only show failing tests
# --compact: Do not show code snippets
# --framework: Only scan terraform code

# Capture the error code
CHECKOV_EXIT_CODE="$?"

# We check the exit code and display a warning if anything was found
if [[ "$CHECKOV_EXIT_CODE" != 0 ]]; then
  echo "^^^ +++"
  echo "Possible Terraform security issues found. "
  echo "Please refer to the Sourcegraph handbook for guidance: https://handbook.sourcegraph.com/product-engineering/engineering/cloud/security/checkov"
  exit 222
fi
