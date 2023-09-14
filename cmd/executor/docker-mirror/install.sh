#!/usr/bin/env bash
set -ex -o nounset -o pipefail

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
            "log_group_name": "executors_docker_mirror",
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

## Install NVMe cli
function install_nvme() {
  apt-get update -y
  apt-get install -y nvme-cli
}

## Run docker registry as a pull-through cache
## Reference: https://docs.docker.com/registry/recipes/mirror/
function setup_pull_through_docker_cache() {
  local DOCKER_REGISTRY_CONFIG_FILE='/etc/docker/registry_config.json'
  mkdir -p "$(dirname "${DOCKER_REGISTRY_CONFIG_FILE}")"
  cat <<EOF >"${DOCKER_REGISTRY_CONFIG_FILE}"
version: 0.1
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /var/lib/registry
http:
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
  debug:
    addr: :5001
    prometheus:
      enabled: true
      path: /metrics
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
proxy:
  remoteurl: https://registry-1.docker.io
EOF
  # TODO: This only monitors the docker client, not the container itself.
  cat <<EOF >/etc/systemd/system/docker_registry.service
[Unit]
Description=Docker Registry
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStart=/usr/bin/docker run \
  --rm \
  -p 5000:5000 \
  -p 5001:5001 \
  -v ${DOCKER_REGISTRY_CONFIG_FILE}:/etc/docker/registry/config.yml \
  -v /mnt/registry:/var/lib/registry \
  --name %n \
  registry:2

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
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
  --collector.loadavg \
  --collector.diskstats \
  --collector.filesystem \
  --collector.meminfo \
  --collector.netclass \
  --collector.netdev \
  --collector.netstat \
  --collector.softnet \
  --collector.pressure \
  --collector.vmstat \
  --collector.vmstat.fields '^(oom_kill|pgpg|pswp|pg.*fault|pgscan|pgsteal).*'
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
  registry:
    method: http
    http:
      port: 5001
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

function preload_registry_image() {
  docker pull registry:2
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
install_nvme
preload_registry_image

# Services
setup_pull_through_docker_cache
install_node_exporter
install_exporter_exporter

# Cleanup
cleanup
