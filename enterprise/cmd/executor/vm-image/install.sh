#!/usr/bin/env bash
set -ex -o nounset -o pipefail

export EXECUTOR_FIRECRACKER_IMAGE="sourcegraph/executor-vm:$VERSION"
export NODE_EXPORTER_VERSION=1.2.2
export NODE_EXPORTER_ADDR="127.0.0.1:9100"

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
  apt-get update
  apt-cache policy docker-ce
  apt-get install -y docker-ce docker-ce-cli containerd.io

  DOCKER_DAEMON_CONFIG_FILE='/etc/docker/daemon.json'

  if [ ! -f "${DOCKER_DAEMON_CONFIG_FILE}" ]; then
    mkdir -p "$(dirname "${DOCKER_DAEMON_CONFIG_FILE}")"
    echo '{"log-driver": "journald"}' >"${DOCKER_DAEMON_CONFIG_FILE}"
  fi

  # Restart Docker daemon to pick up our changes.
  systemctl restart --now docker
}

## Install git >=2.26 (to enable -c protocol.version=2 and sparse checkouts)
function install_git() {
  add-apt-repository ppa:git-core/ppa
  apt-get update -y
  apt-get install -y git
}

## Install and configure executor service
function install_executor() {
  # Move binary into PATH
  mv /tmp/executor /usr/local/bin

  # Run all the installers:
  # TODO: Replace this by executor install all. For that install images executor-vm
  # has to work in this VM box though.

  # Loads the required kernel image so it doesn't have to happen on the first VM start.
  /usr/local/bin/executor install image kernel
  # Loads the required sandbox docker image so it doesn't have to happen on the first VM start.
  /usr/local/bin/executor install image sandbox
  # Install Weaveworks Ignite
  # Reference: https://ignite.readthedocs.io/en/stable/installation/
  # Install dependencies. Most of these are actually bundled by default, but
  # listing them out here explicitly makes it so that upstream image changes never
  # negatively impact us.
  apt-get update
  apt-get install -y mount tar binutils e2fsprogs openssh-client dmsetup
  # Download and install ignite binary.
  /usr/local/bin/executor install ignite
  # Install the CNI plus plugins, used by ignite.
  /usr/local/bin/executor install cni
  # Install src-cli to the host system. It's needed for src steps outside of firecracker.
  /usr/local/bin/executor install src-cli
  # Configures iptables rules for our ignite VMs. We don't want to allow any local
  # traffic except the traffic to nameservers. This is to prevent any internal attack
  # vector and talking to link-local services like the google metadata server.
  # Make sure the below install doesn't block.
  echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
  # Ensure iptables-persistent is installed.
  apt-get install -y iptables-persistent
  # Install all required rules.
  /usr/local/bin/executor install iptables-rules
  # Store the iptables config.
  mkdir -p /etc/iptables
  iptables-save >/etc/iptables/rules.v4

  # Create configuration file and stub environment file.
  # We also wait for docker to be ready, otherwise
  # jobs can fail to start while docker is still starting.
  cat <<EOF >/etc/systemd/system/executor.service
[Unit]
Description=User code executor

[Service]
ExecStart=/usr/local/bin/executor
ExecStopPost=/shutdown_executor.sh
Requires=docker
Restart=on-failure
EnvironmentFile=/etc/systemd/system/executor.env
Environment=HOME="%h"
Environment=SRC_LOG_LEVEL=dbug
Environment=EXECUTOR_FIRECRACKER_IMAGE="${EXECUTOR_FIRECRACKER_IMAGE}"
Environment=NODE_EXPORTER_URL="http://${NODE_EXPORTER_ADDR}"

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

## Verify executor is working properly.
function verify_executor() {
  # TODO: Later we might also want to use executor validate here, but it depends
  # on the env vars set in the terraform modules.
  # Start a VM to see if that succeeds. Then, clean up the VM so we don't leave it
  # behind in the image.
  VM="$(/usr/local/bin/executor test-vm --name-only)"
  ignite rm --force "$VM"
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
  --web.listen-address="${NODE_EXPORTER_ADDR}" \
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

## Build the sourcegraph/executor-vm image for use in firecracker.
## Set SRC_CLI_VERSION to the minimum required version in internal/src-cli/consts.go
function generate_ignite_base_image() {
  # TODO: Find a way to use executor install image executor-vm here.
  docker build -t "${EXECUTOR_FIRECRACKER_IMAGE}" --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" /tmp/executor-vm
  ignite image import --runtime docker "${EXECUTOR_FIRECRACKER_IMAGE}"
  docker image rm "${EXECUTOR_FIRECRACKER_IMAGE}"
  # Remove intermediate layers and base image used in executor-vm.
  docker system prune --force
}

function cleanup() {
  apt-get -y autoremove
  apt-get clean
  rm -rf /var/cache/*
  rm -rf /var/lib/apt/lists/*
  history -c
}

# Install cloud specific helpers.
if [ "${PLATFORM_TYPE}" == "gcp" ]; then
  install_ops_agent
elif [ "${PLATFORM_TYPE}" == "aws" ]; then
  install_cloudwatch_agent
fi

# Install dependencies.
install_docker
install_git

# Install the optional node exporter dependency.
install_node_exporter

# Install and setup executor.
install_executor
generate_ignite_base_image
verify_executor

# Final cleanup.
cleanup
