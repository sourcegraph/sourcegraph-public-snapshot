#!/usr/bin/env bash

set -eu
# Ensure we're at the root of the repository.
if [ ! -d ".git" ]; then
  echo "This command must run at the root of the sourcegraph repository."
  echo "Please run it again with:"
  echo "  bazel run :release_patch --run_under=\"cd $PWD &&\""
  exit 1
fi

echo "Checking out a new branch named {{new_version}}"
git checkout -b {{new_version}}

# Update the buildfile for schema_descriptions so it has our new schema.
buildozer 'add outs schema-descriptions/{{new_version}}-internal_database_schema.codeinsights.json schema-descriptions/{{new_version}}-internal_database_schema.codeintel.json schema-descriptions/{{new_version}}-internal_database_schema.json' //cmd/migrator:schema_descriptions

# Update the shell script powering that target
echo "{{new_version}}" >> cmd/migrator/wip_git_versions.txt

# Ensure the result is correct
bazel test //cmd/migrator:schema_descriptions_test

git add cmd/migrator/BUILD.bazel
git add cmd/migrator/wip_git_versions.text

git commit -m "release_patch: build {{new_version}}"
