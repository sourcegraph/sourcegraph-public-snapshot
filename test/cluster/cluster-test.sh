#!/usr/bin/env bash
set -euxo pipefail

# setup DIR for easier pathing /Users/dax/work/sourcegraph/test/cluster
DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)""
# cd to repo root
cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit

function cluster_setup() {
  git clone --depth 1 --branch v3.20.1 \
    https://github.com/sourcegraph/deploy-sourcegraph.git \
    "$DIR/deploy-sourcegraph"

  # TODO(Dax): Bit concerning this works...
  gcloud container clusters get-credentials default-buildkite --zone=us-central1-c --project=sourcegraph-ci

  #NAMESPACE="cluster-ci-$BUILDKITE_BUILD_NUMBER"
  export NAMESPACE=cluster-ci-122
  kubectl create ns $NAMESPACE -oyaml --dry-run | kubectl apply -f -
  kubectl apply -f "$DIR/storageClass.yaml"
  kubectl config set-context --current --namespace="$NAMESPACE"
  kubectl config current-context
  sleep 15 #wait for namespace to come up
  kubectl get -n $NAMESPACE pods

  pushd "$DIR/deploy-sourcegraph/"
  pwd
  # script contains relative paths :(
  ./create-new-cluster.sh
  popd

  kubectl get pods
  time kubectl wait --for=condition=Ready -l app=sourcegraph-frontend pod --timeout=20m
  #shellcheck disable=SC2064
  trap "kubectl delete namespace $NAMESPACE" EXIT
}

function test_setup() {

  set +x +u
  # shellcheck disable=SC1091
  source /root/.profile

  test/setup-deps.sh

  sleep 15
  export SOURCEGRAPH_BASE_URL="http://sourcegraph-frontend.$NAMESPACE.svc.cluster.local:30080"
  curl $SOURCEGRAPH_BASE_URL

  # setup admin users, etc
  go run test/init-server.go -base-url=$SOURCEGRAPH_BASE_URL

  # Load variables set up by init-server, disabling `-x` to avoid printing variables, setting +u to avoid blowing up on ubound ones
  set +x +u
  # shellcheck disable=SC1091
  source /root/.profile
  set -x

  echo "TEST: Checking Sourcegraph instance is accessible"

  curl --fail $SOURCEGRAPH_BASE_URL
  curl --fail "$SOURCEGRAPH_BASE_URL/healthz"
}

function e2e() {
  echo "TEST: Running tests"
  pushd client/web
  echo $SOURCEGRAPH_BASE_URL
  # TODO: File issue for broken test
  #yarn run test:regression:core
  yarn run test:regression:config-settings
  popd
}

# main
cluster_setup
test_setup
e2e
