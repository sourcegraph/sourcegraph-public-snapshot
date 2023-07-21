#!/usr/bin/env bash

set -e

# See https://github.com/sourcegraph/infrastructure/blob/bd8fec102059e25a0ad56591e0eb836db1afdf4b/docker-images/buildkite-agent-stateless/Dockerfile#L348
SONAR_HOME="/opt/sonar-scanner"

export SONAR_SCANNER_OPTS="-server"

export SONAR_TOKEN="${SONAR_TOKEN}"

if [ "$SONAR_TOKEN" = "" ];
then
  echo "Please set the SONAR_TOKEN environment variable"
  exit 1
fi

set -x

echo -e "--- :arrow_down: verifying Sonarcloud binary\n\n"
"${SONAR_HOME}/bin/sonar-scanner" --version

echo -e "--- :lock: running Sonarcloud scan\n"

# if pull request build scan for diff or perform branch analysis
if [ "$BUILDKITE_PULL_REQUEST" = "false" ]; then
  "${SONAR_HOME}/bin/sonar-scanner" \
    -Dsonar.organization=sourcegraph \
    -Dsonar.projectKey=sourcegraph_sourcegraph \
    -Dsonar.sources=. \
    -Dsonar.host.url=https://sonarcloud.io \
    -Dsonar.sourceEncoding=UTF-8
else
  "${SONAR_HOME}/bin/sonar-scanner" \
    -Dsonar.organization=sourcegraph \
    -Dsonar.projectKey=sourcegraph_sourcegraph \
    -Dsonar.sources=. \
    -Dsonar.host.url=https://sonarcloud.io \
    -Dsonar.sourceEncoding=UTF-8 \
    -Dsonar.pullrequest.key="$BUILDKITE_PULL_REQUEST" \
    -Dsonar.pullrequest.branch="$BUILDKITE_BRANCH" \
    -Dsonar.pullrequest.base="$BUILDKITE_PULL_REQUEST_BASE_BRANCH"
fi
