#!/usr/bin/env bash

VERSION=$1
fd=$(mktemp)

cat <<EOF | buildkite-agent annotate --style=info --context=cloud-ephemeral
  <div class="flex">
    <div>
    Images in this build will be pushed to the Cloud Ephemeral registry with the following tag/version
    <pre class="term">$VERSION</pre>
    <div>
    Using this version you create a Cloud Ephemeral deployment by running
    <pre class="term">sg cloud deploy --version "$VERSION"</pre>
    Or you can upgrade an existing Cloud Ephemeral deployment by running
    <pre class="term">sg cloud upgrade --version "$VERSION"</pre>
    </div>
  </div>
  <div class="ml-auto">
  :gcp: Cloud Ephemeral
  </div>
EOF
