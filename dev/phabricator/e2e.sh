set -e

echo "--- Phabricator"
source ./dev/phabricator/start.sh
PHABRICATOR_CONTAINER="$(docker ps -aq -f name=phabricator$)"

# Install the Sourcegraph native integration
source ./dev/phabricator/install-sourcegraph.sh

pushd web
yarn run test-phabricator-e2e
popd
