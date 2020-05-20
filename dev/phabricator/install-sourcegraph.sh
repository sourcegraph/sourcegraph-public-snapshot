#!/usr/bin/env bash

until curl -s "http://127.0.0.1" | grep -q "Login to Phabricator"; do
  echo "Waiting for the Phabricator instance to be ready..."
  sleep 2s
done

sourcegraph_extension="https://github.com/sourcegraph/phabricator-extension.git"

phab_container=$(docker ps -aq -f name=phabricator$)
phab_directory="/opt/bitnami/phabricator"

# Install the Phabricator native extension
docker exec -it "$phab_container" \
  sh -c "cd ${phab_directory}/src/extensions && git clone -b release-v1.2 ${sourcegraph_extension} sourcegraph"

# Add the static CSS/JS assets
docker exec -it "$phab_container" \
  sh -c "cd ${phab_directory} && ./bin/celerity map"
