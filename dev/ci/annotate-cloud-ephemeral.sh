#!/usr/bin/env bash

VERSION=$1
fd=$(mktemp)

cat <<EOF | buildkite-agent annotate --style=info --context=cloud-ephemeral
  <div class="flex">
    <div>
    Images in this build will be pushed to the Cloud Ephemeral registry with the following tag/version
    <pre>$VERSION</pre>

    <p>
    Using this version you create a Cloud Ephemeral deployment by running
    </p>
      <pre class="term">
      <code>
        sg cloud deploy --version "$VERSION"
      </code>
    </pre>
    Or you can upgrade an existing Cloud Ephemeral deployment by running
      <pre class="term">
      <code>
        sg cloud upgrade --version "$VERSION"
      </code>
    </pre>
  </div>
  <div class="ml-auto">
  :cloud: Cloud Ephemeral
  </div>
EOF
