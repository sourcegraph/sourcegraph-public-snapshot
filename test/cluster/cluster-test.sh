#!/usr/bin/env bash
set -euxo pipefail

# setup DIR for easier pathing /Users/dax/work/sourcegraph/test/cluster
DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)""
# cd to repo root
cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit

function cluster_setup() {
git clone --depth 1 \
  https://github.com/sourcegraph/deploy-sourcegraph.git \
  "$DIR/deploy-sourcegraph"

#NAMESPACE="cluster-ci-$BUILDKITE_BUILD_NUMBER"
# TODO(Dax): Buildkite cannot create namespaces at cluster level
NAMESPACE=cluster-ci-122
#kubectl create namespace "$NAMESPACE"

# TODO(Dax): Bit concerning this works...
gcloud container clusters get-credentials default-buildkite --zone=us-central1-c --project=sourcegraph-ci
kubectl config current-context

kubectl apply -f "$DIR/storageClass.yaml"
kubectl config set-context --current --namespace="$NAMESPACE"
kubectl get -n $NAMESPACE pods

pushd "$DIR/deploy-sourcegraph/"
pwd
# script contains relative paths :(
./create-new-cluster.sh
popd

kubectl get pods
time kubectl wait --for=condition=Ready -l app=sourcegraph-frontend pod --timeout=20m
}

function test_setup() {
  LOGFILE=frontend-logs
  # kubectl logs
  kubectl_logs() {
    echo "Appending frontend logs"
    kubectl logs -l "app=sourcegraph-frontend" -c frontend >>$LOGFILE.log
    chmod 744 $LOGFILE.log
    #kubectl delete namespace $NAMESPACE
  }
  trap kubectl_logs EXIT

  set -x

  test/setup-deps.sh

  sleep 15
  SOURCEGRAPH_URL="http://sourcegraph-frontend.$NAMESPACE.svc.cluster.local:30080"
  curl $SOURCEGRAPH_URL

  # setup admin users, etc
  go run test/init-server.go -base-url=$SOURCEGRAPH_URL

  # Load variables set up by init-server, disabling `-x` to avoid printing variables
  set +x
  # shellcheck disable=SC1091
  source /root/.profile
  set -x

  echo "TEST: Checking Sourcegraph instance is accessible"

  curl --fail $SOURCEGRAPH_URL
  curl --fail "$SOURCEGRAPH_URL/healthz"
}

function e2e() {
  echo "TEST: Running tests"
  pushd client/web || exit
  yarn run test:regression:core

  popd || exit
}

# main
cluster_setup
test_setup
e2e

# ==========================
test/cleanup-display.sh
