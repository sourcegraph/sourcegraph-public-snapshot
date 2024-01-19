#! /bin/bash

# Run bazel configure and if the error code is 110, exit with error code 0
# This is because 110 means that configuration files were successfully
# Can be used by processes which want to run configuration as an auto-fix
# and expect a 0 exit code
bazel configure
exit_code=$?

if [ $exit_code -eq 0 ]; then
  echo "No configuration changes made"
  exit 0
elif [ $exit_code -eq 110 ]; then
  echo "Bazel configuration completed"
  exit 0
else
  echo "Unknown error"
  exit $exit_code
fi
