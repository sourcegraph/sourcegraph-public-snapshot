#!/usr/bin/env bash

VERSION=$1
fd=$(mktemp)

cat <<EOF | buildkite-agent annotate --style=info --context=cloud-ephemeral
  <div class="flex">
    <div>
    Images in this build will be pused to the Cloud Ephemeral registry with the following tag/version
    <pre>$VERSION</pre>

    Using this version you create a Cloud Ephemeral deployment by running
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
    <img class="emoji" alt="cloud-emoji" src="https://buildkiteassets.com/emojis/img-apple-64/2601-fe0f.png"/>
    Cloud Ephemeral
  </div>
EOF
