#!/usr/bin/env bash
set -ex -o nounset -o pipefail

export IGNITE_VERSION=v0.10.0
export CNI_VERSION=v0.9.1
export EXECUTOR_FIRECRACKER_IMAGE="sourcegraph/ignite-ubuntu:insiders"

## Install logging agent
## Reference: https://cloud.google.com/logging/docs/agent/installation
function install_logging_agent() {
  curl -sSO https://dl.google.com/cloudagents/add-logging-agent-repo.sh
  bash ./add-logging-agent-repo.sh
  rm add-logging-agent-repo.sh
  apt-get update -y
  apt-get install -y 'google-fluentd=1.*' google-fluentd-catch-all-config-structured
  systemctl start google-fluentd
}

## Install monitoring agent
## Reference: https://cloud.google.com/monitoring/agent/installation
function install_monitoring_agent() {
  curl -sSO https://dl.google.com/cloudagents/add-monitoring-agent-repo.sh
  bash ./add-monitoring-agent-repo.sh
  rm add-monitoring-agent-repo.sh
  apt-get update -y
  apt-get install -y 'stackdriver-agent=6.*'
  systemctl start stackdriver-agent
}

## Install Docker
function install_docker() {
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
  add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  apt-get update -y
  apt-cache policy docker-ce
  apt-get install -y binutils docker-ce docker-ce-cli containerd.io

  DOCKER_DAEMON_CONFIG_FILE='/etc/docker/daemon.json'

  if [ ! -f "${DOCKER_DAEMON_CONFIG_FILE}" ]; then
    mkdir -p "$(dirname "${DOCKER_DAEMON_CONFIG_FILE}")"
    echo '{"log-driver": "journald"}' >"${DOCKER_DAEMON_CONFIG_FILE}"
  fi

  # Restart Docker daemon to pick up our changes.
  systemctl restart --now docker
}

## Install git >=2.18 (to enable -c protocol.version=2)
function install_git() {
  add-apt-repository ppa:git-core/ppa
  apt-get update -y
  apt-get install -y git
}

## Install Weaveworks Ignite
## Reference: https://ignite.readthedocs.io/en/stable/installation/
function install_ignite() {
  # Install ignite
  curl -sfLo ignite https://github.com/weaveworks/ignite/releases/download/${IGNITE_VERSION}/ignite-amd64
  chmod +x ignite
  mv ignite /usr/local/bin

  # Install ignited
  curl -sfLo ignited https://github.com/weaveworks/ignite/releases/download/${IGNITE_VERSION}/ignited-amd64
  chmod +x ignited
  mv ignited /usr/local/bin

  # Install container network interface
  mkdir -p /opt/cni/bin
  curl -sSL https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-amd64-${CNI_VERSION}.tgz | tar -xz -C /opt/cni/bin
}

## Install and configure executor service
function install_executor() {
  # Move binary into PATH
  mv /tmp/executor /usr/local/bin

  # Create configuration file and stub environment file
  cat <<EOF >/etc/systemd/system/executor.service
[Unit]
Description=User code executor

[Service]
ExecStart=/usr/local/bin/executor
Restart=always
EnvironmentFile=/etc/systemd/system/executor.env
Environment=HOME="%h"
Environment=SRC_LOG_LEVEL=dbug
Environment=EXECUTOR_FIRECRACKER_IMAGE="${EXECUTOR_FIRECRACKER_IMAGE}"

[Install]
WantedBy=multi-user.target
EOF
  echo 'THIS_ENV_IS="unconfigured"' >>/etc/systemd/system/executor.env
}

## Build the ignite-ubuntu image for use in firecracker.
## Set SRC_CLI_VERSION to the minimum required version in internal/src-cli/consts.go
function generate_ignite_base_image() {
  docker build -t "${EXECUTOR_FIRECRACKER_IMAGE}" --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" /tmp/ignite-ubuntu
  ignite image import --runtime docker "${EXECUTOR_FIRECRACKER_IMAGE}"
  docker image rm "${EXECUTOR_FIRECRACKER_IMAGE}"
}

function cleanup() {
  apt-get -y autoremove
  apt-get clean
  rm -rf /var/cache/*
  rm -rf /var/lib/apt/lists/*
  history -c
}

# Prerequisites
install_logging_agent
install_monitoring_agent
install_docker
install_git
install_ignite

# Services
install_executor

# Service prep and cleanup
generate_ignite_base_image
cleanup
