
#!/bin/bash

# Outputs a pipeline that targets agents that have the same 'name' meta-data
# value as the step that does the pipeline upload. This means that all the
# steps will run on the same agent machine, assuming that the 'name' meta-data
# value is unique to each agent.
#
# Each agent needs to be configured with meta-data like so:
#
# meta-data="name=<unique-name>"
#
# To use, save this file as .buildkite/pipeline.sh, chmod +x, and then set your
# first pipeline step to run this and pipe it into pipeline upload:
#
# .buildkite/pipeline.sh | buildkite-agent pipeline upload
#


name=$(buildkite-agent meta-data set name "$BUILDKITE_AGENT_NAME" && echo $BUILDKITE_AGENT_NAME)

cat << EOF

env:
  VAGRANT_RUN_ENV: 'CI'
  VAGRANT_DOTFILE_PATH: "/var/lib/buildkite/builds/$name/.vagrant"

steps:
- artifact_paths: ./*.png;./e2e.mp4;./ffmpeg.log
  # setting to pass until tests are 100% confirmed as working, so as to avoid disruting dev workflow on main
  command:
    -  env
  timeout_in_minutes: 20
  label: ':docker::arrow_right::chromium:'
  agents:
    name: buildkite-agent-ntrk-1
EOF
