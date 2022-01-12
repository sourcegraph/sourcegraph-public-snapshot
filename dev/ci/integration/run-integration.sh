#!/usr/bin/env bash

# This script sets up a Sourcegraph instance for integration testing. This script expects to be
# passed a path to a bash script that runs the actual tests against a running instance. The passed
# script will be passed a single parameter: the target URL from which the instance is accessible.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
root_dir=$(pwd)
set -ex

echo "--- set up deploy-sourcegraph-docker"
test_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)""
clone_dir="$test_dir/deploy-sourcegraph-docker"
rm -rf "$clone_dir"
git clone --depth 1 \
  https://github.com/sourcegraph/deploy-sourcegraph-docker.git \
  "$clone_dir"
compose_dir="$test_dir/deploy-sourcegraph-docker/docker-compose"

function docker_cleanup() {
  echo "--- docker cleanup"
  if [[ $(docker ps -aq | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rm -f $(docker ps -aq)
  fi
  if [[ $(docker images -q | wc -l) -gt 0 ]]; then
    # shellcheck disable=SC2046
    docker rmi -f $(docker images -q)
  fi
  docker volume prune -f
  docker network prune -f
}

# Do a pre-run cleanup
docker_cleanup

function cleanup() {
  exit_status=$?
  if [ $exit_status -ne 0 ]; then
    # Expand the output if our run failed.
    echo "^^^ +++"
  fi

  pushd "$compose_dir"
  echo "--- dump server logs"
  docker-compose logs >"$root_dir/server.log" 2>&1

  echo "--- stop project"
  docker-compose down
  popd

  docker_cleanup

  if [ $exit_status -ne 0 ]; then
    # This command will fail, so our last step will be expanded. We don't want
    # to expand "docker cleanup" so we add in a dummy section.
    echo "--- integration test failed"
    echo "See integration test section for test runner logs, and uploaded artefacts for server logs."
  fi
}
trap cleanup EXIT

pushd "$compose_dir"

echo "--- Generating compose config"
# Overwrite image versions
if [ -z "$DOCKER_COMPOSE_IMAGES" ]; then
  # Expects newline-delimited list of image names to update, see pipeline generator for
  # how this variable is generated.
  while IFS= read -r line; do
    echo "$line"
    grep -lr './' -e "index.docker.io/sourcegraph/$line" --include \*.yaml | xargs sed -i -E "s#index.docker.io/sourcegraph/$line:.*#us.gcr.io/sourcegraph-dev/$line:$VERSION#g"
  done < <(printf '%s\n' "$DOCKER_COMPOSE_IMAGES")
fi

# Customize deployment
yq eval 'del(.services.caddy)' --inplace docker-compose.yaml
yq eval 'del(.services.prometheus)' --inplace docker-compose.yaml
yq eval 'del(.services.grafana)' --inplace docker-compose.yaml
yq eval 'del(.services.jaeger)' --inplace docker-compose.yaml
yq eval '.services.*.cpus = 1' --inplace docker-compose.yaml
yq eval '.services.sourcegraph-frontend-0.ports = [ "0.0.0.0:7080:3080" ]' --inplace docker-compose.yaml
docker-compose config

echo "--- Running Sourcegraph"
docker-compose up --detach --force-recreate --renew-anon-volumes --quiet-pull
popd

URL="http://localhost:7080"
echo "--- Waiting for $URL to be up"
set +e
timeout 120s bash -c "until curl --output /dev/null --silent --head --fail $URL; do
    echo Waiting 5s for $URL...
    sleep 5
done"
# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
  echo "^^^ +++"
  echo "$URL was not accessible within 120s."
  docker top
  exit 1
fi
set -e
echo "Waiting for $URL... done"

# Run tests against instance
echo "--- Running ${1} against $URL"
"${1}" "${URL}"
