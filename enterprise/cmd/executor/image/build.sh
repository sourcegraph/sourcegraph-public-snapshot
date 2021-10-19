#!/usr/bin/env bash
set -ex -o nounset -o pipefail

export IGNITE_VERSION=v0.10.0
export CNI_VERSION=v0.9.1
export KERNEL_IMAGE="weaveworks/ignite-kernel:5.10.51"
export EXECUTOR_FIRECRACKER_IMAGE="sourcegraph/ignite-ubuntu:insiders"
export NODE_EXPORTER_VERSION=1.2.2
export EXPORTER_EXPORTER_VERSION=0.4.5

## Install ops agent
## Reference: https://cloud.google.com/logging/docs/agent/ops-agent/installation
function install_ops_agent() {
  curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
  sudo bash add-google-cloud-ops-agent-repo.sh --also-install
}

## Install CloudWatch agent
## Reference: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Install-CloudWatch-Agent.html
function install_cloudwatch_agent() {
  wget -q https://s3.us-west-2.amazonaws.com/amazoncloudwatch-agent-us-west-2/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
  dpkg -i -E ./amazon-cloudwatch-agent.deb
  rm ./amazon-cloudwatch-agent.deb

  CLOUDWATCH_CONFIG_FILE_PATH=/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json
  cat <<EOF >"${CLOUDWATCH_CONFIG_FILE_PATH}"
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/syslog",
            "log_group_name": "executors",
            "timezone": "UTC"
          }
        ]
      }
    },
    "log_stream_name": "{instance_id}-syslog"
  }
}
EOF
  /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -s -c file:"${CLOUDWATCH_CONFIG_FILE_PATH}"
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
  curl -sfLo ignite https://github.com/weaveworks/ignite/releases/download/${IGNITE_VERSION}/ignite-amd64
  chmod +x ignite
  mv ignite /usr/local/bin

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
ExecStopPost=/shutdown_executor.sh
Restart=on-failure
EnvironmentFile=/etc/systemd/system/executor.env
Environment=HOME="%h"
Environment=SRC_LOG_LEVEL=dbug
Environment=SRC_PROF_HTTP=127.0.0.1:6060
Environment=EXECUTOR_FIRECRACKER_IMAGE="${EXECUTOR_FIRECRACKER_IMAGE}"

[Install]
WantedBy=multi-user.target
EOF

  # Create empty environment file (overwritten on VM startup)
  cat <<EOF >/etc/systemd/system/executor.env
THIS_ENV_IS="unconfigured"
EOF

  # Write a script to shutdown the host after clean exit from executor.
  # This is meant to support our scaling pattern, where each executor will
  # run for a pre-determined amount of time before exiting. We only need to
  # scale up in this situation, as executors will naturally exit and not be
  # replaced during periods of lighter loads.

  cat <<EOF >/shutdown_executor.sh
#!/usr/bin/env bash

if [ "\${EXIT_STATUS}" == "0" ]; then
  echo 'Executor has exited cleanly. Shutting down host.'
  systemctl poweroff
else
  echo 'Executor has exited with an error. Service will restart.'
fi
EOF

  # Ensure systemd can execute shutdown script
  chmod +x /shutdown_executor.sh
}

function install_node_exporter() {
  useradd --system --shell /bin/false node_exporter

  wget https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz
  tar xvfz node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz
  mv node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64/node_exporter /usr/local/bin/node_exporter
  rm -rf node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64 node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz

  chown node_exporter:node_exporter /usr/local/bin/node_exporter

  cat <<EOF >/etc/systemd/system/node_exporter.service
[Unit]
Description=Node Exporter
[Service]
User=node_exporter
ExecStart=/usr/local/bin/node_exporter \
  --web.listen-address="127.0.0.1:9100" \
  --collector.disable-defaults \
  --collector.cpu \
  --collector.diskstats \
  --collector.filesystem \
  --collector.meminfo \
  --collector.netclass \
  --collector.netdev \
  --collector.netstat \
  --collector.softnet \
  --collector.pressure \
  --collector.vmstat
[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable node_exporter
}

function install_exporter_exporter() {
  useradd --system --shell /bin/false exporter_exporter

  wget https://github.com/QubitProducts/exporter_exporter/releases/download/v${EXPORTER_EXPORTER_VERSION}/exporter_exporter-${EXPORTER_EXPORTER_VERSION}.linux-amd64.tar.gz
  tar xvfz exporter_exporter-${EXPORTER_EXPORTER_VERSION}.linux-amd64.tar.gz
  mv exporter_exporter-${EXPORTER_EXPORTER_VERSION}.linux-amd64/exporter_exporter /usr/local/bin/exporter_exporter
  rm -rf exporter_exporter-${EXPORTER_EXPORTER_VERSION}.linux-amd64 exporter_exporter-${EXPORTER_EXPORTER_VERSION}.linux-amd64.tar.gz

  chown exporter_exporter:exporter_exporter /usr/local/bin/exporter_exporter

  cat <<EOF >/usr/local/bin/exporter_exporter.yaml
modules:
  node:
    method: http
    http:
      port: 9100
  executor:
    method: http
    http:
      port: 6060
EOF

  cat <<EOF >/etc/systemd/system/exporter_exporter.service
[Unit]
Description=Exporter Exporter
[Service]
User=exporter_exporter
ExecStart=/usr/local/bin/exporter_exporter -config.file "/usr/local/bin/exporter_exporter.yaml"
[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable exporter_exporter
}

## Build the ignite-ubuntu image for use in firecracker.
## Set SRC_CLI_VERSION to the minimum required version in internal/src-cli/consts.go
function generate_ignite_base_image() {
  docker build -t "${EXECUTOR_FIRECRACKER_IMAGE}" --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" /tmp/ignite-ubuntu
  ignite image import --runtime docker "${EXECUTOR_FIRECRACKER_IMAGE}"
  docker image rm "${EXECUTOR_FIRECRACKER_IMAGE}"
  # Remove intermediate layers and base image used in ignite-ubuntu.
  docker system prune --force
}

## Loads the required kernel image so it doesn't have to happen on the first VM start.
function preheat_kernel_image() {
  ignite kernel import --runtime docker "${KERNEL_IMAGE}"
  docker pull "weaveworks/ignite:${IGNITE_VERSION}"
}

function cleanup() {
  apt-get -y autoremove
  apt-get clean
  rm -rf /var/cache/*
  rm -rf /var/lib/apt/lists/*
  history -c
}

# Prerequisites
if [ "${PLATFORM_TYPE}" == "gcp" ]; then
  install_ops_agent
else
  install_cloudwatch_agent
fi
install_docker
install_git
install_ignite

# Services
install_executor
install_node_exporter
install_exporter_exporter

# Service prep and cleanup
generate_ignite_base_image
preheat_kernel_image
cleanup
