#!/usr/bin/env bash

set -e
export SONAR_SCANNER_VERSION=4.7.0.2747
export SONAR_SCANNER_HOME=/buildkite/sonar/sonar-scanner-$SONAR_SCANNER_VERSION-linux
export SONAR_SCANNER_OPTS="-server"

echo "--- :arrow_down: downloading Sonarcloud binary"
echo ""
curl --create-dirs -sSLo /buildkite/sonar/sonar-scanner.zip https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-$SONAR_SCANNER_VERSION-linux.zip
unzip -o /buildkite/sonar/sonar-scanner.zip -d /buildkite/sonar/

echo "--- :lock: running Sonarcloud scan"
echo ""
cd /buildkite-git-references/sourcegraph.reference
$SONAR_SCANNER_HOME/bin/sonar-scanner \
  -Dsonar.organization=sourcegraph \
  -Dsonar.projectKey=sourcegraph_sourcegraph \
  -Dsonar.sources=. \
  -Dsonar.host.url=https://sonarcloud.io
