#!/usr/bin/env bash
set -ex -o nounset -o pipefail

export IGNITE_VERSION=v0.10.4
export CNI_VERSION=v0.9.1
export RUNTIME_IMAGE="sourcegraph/ignite:${IGNITE_VERSION}"
export KERNEL_IMAGE="sourcegraph/ignite-kernel:5.10.135-amd64"
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

## Install Weaveworks Ignite
## Reference: https://ignite.readthedocs.io/en/stable/installation/
function install_ignite() {
  # Install dependencies. Most of these are actually bundled by default, but
  # listing them out here explicitly makes it so that upstream image changes never
  # negatively impact us.
  apt-get update
  apt-get install -y mount tar binutils e2fsprogs openssh-client dmsetup

  # Download and install ignite binary.
  curl -sfLo ignite https://github.com/sourcegraph/ignite/releases/download/${IGNITE_VERSION}/ignite-amd64
  chmod +x ignite
  mv ignite /usr/local/bin
}

## Install the CNI plus plugins, used by ignite.
function install_cni() {
  mkdir -p /opt/cni/bin
  curl -sSL https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-amd64-${CNI_VERSION}.tgz | tar -xz -C /opt/cni/bin
  # Also install the isolation plugin.
  curl -sSL https://github.com/AkihiroSuda/cni-isolation/releases/download/v0.0.4/cni-isolation-amd64.tgz | tar -xz -C /opt/cni/bin
}

## Install and configure executor service
function install_executor() {
  # Move binary into PATH
  mv /tmp/executor /usr/local/bin

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
Environment=SRC_PROF_HTTP=127.0.0.1:6060
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

# Install src-cli to the host system. It's needed for src steps outside of firecracker.
function install_src_cli() {
  curl -f -L -o src-cli.tar.gz "https://github.com/sourcegraph/src-cli/releases/download/${SRC_CLI_VERSION}/src-cli_${SRC_CLI_VERSION}_linux_amd64.tar.gz"
  tar -xvzf src-cli.tar.gz src
  mv src /usr/local/bin/src
  chmod +x /usr/local/bin/src
  rm -rf src-cli.tar.gz
}

## Build the sourcegraph/executor-vm image for use in firecracker.
## Set SRC_CLI_VERSION to the minimum required version in internal/src-cli/consts.go
function generate_ignite_base_image() {
  docker build -t "${EXECUTOR_FIRECRACKER_IMAGE}" --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" /tmp/executor-vm
  ignite image import --runtime docker "${EXECUTOR_FIRECRACKER_IMAGE}"
  docker image rm "${EXECUTOR_FIRECRACKER_IMAGE}"
  # Remove intermediate layers and base image used in executor-vm.
  docker system prune --force
}

## Loads the required kernel image so it doesn't have to happen on the first VM start.
function preheat_kernel_image() {
  ignite kernel import --runtime docker "${KERNEL_IMAGE}"
  # Also preload the runtime image.
  docker pull "${RUNTIME_IMAGE}"
}

## Configures the CNI explicitly and adds the isolation plugin to the chain.
## This is to prevent cross-network communication (which currently doesn't happen
## as we only have 1 bridge).
## We also set the maximum bandwidth usable per VM to 500 MBit to avoid abuse and
## to make sure multiple VMs on the same host won't starve others.
function configure_cni() {
  mkdir -p /etc/cni/net.d
  cat <<EOF >/etc/cni/net.d/10-ignite.conflist
{
  "cniVersion": "0.4.0",
  "name": "ignite-cni-bridge",
  "plugins": [
    {
      "type": "bridge",
      "bridge": "ignite0",
      "isGateway": true,
      "isDefaultGateway": true,
      "promiscMode": false,
      "ipMasq": true,
      "ipam": {
        "type": "host-local",
        "subnet": "10.61.0.0/16"
      }
    },
    {
      "type": "portmap",
      "capabilities": {
        "portMappings": true
      }
    },
    {
      "type": "firewall"
    },
    {
      "type": "isolation"
    },
    {
      "name": "slowdown",
      "type": "bandwidth",
      "ingressRate": 524288000,
      "ingressBurst": 1048576000,
      "egressRate": 524288000,
      "egressBurst": 1048576000
    }
  ]
}
EOF
}

## Configures iptables rules for our ignite VMs. We don't want to allow any local
## traffic except the traffic to nameservers. This is to prevent any internal attack
## vector and talking to link-local services like the google metadata server.
function setup_iptables() {
  # Make sure the below install doesn't block.
  echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
  # Ensure iptables-persistent is installed.
  apt-get install -y iptables-persistent

  # Ensure the chain exists.
  iptables --list | grep CNI-ADMIN 1>/dev/null || iptables -N CNI-ADMIN

  # Explicitly allow DNS traffic (currently, the DNS server lives in the private
  # networks for GCP and AWS. Ideally we'd want to use an internet-only DNS server
  # to prevent leaking any network details).
  iptables -A CNI-ADMIN -p udp --dport 53 -j ACCEPT

  # Disallow any host-VM network traffic from the guests, except connections made
  # FROM the host (to ssh into the guest).
  iptables -A INPUT -d 10.61.0.0/16 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
  iptables -A INPUT -s 10.61.0.0/16 -j DROP

  # Disallow any inter-VM traffic.
  # But allow to reach the gateway for internet access.
  iptables -A CNI-ADMIN -s 10.61.0.1/32 -d 10.61.0.0/16 -j ACCEPT
  iptables -A CNI-ADMIN -d 10.61.0.0/16 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
  iptables -A CNI-ADMIN -s 10.61.0.0/16 -d 10.61.0.0/16 -j DROP

  # Disallow local networks access.
  iptables -A CNI-ADMIN -s 10.61.0.0/16 -d 10.0.0.0/8 -p tcp -j DROP
  iptables -A CNI-ADMIN -s 10.61.0.0/16 -d 192.168.0.0/16 -p tcp -j DROP
  iptables -A CNI-ADMIN -s 10.61.0.0/16 -d 172.16.0.0/12 -p tcp -j DROP
  # Disallow link-local traffic, too. This usually contains cloud provider
  # resources that we don't want to expose.
  iptables -A CNI-ADMIN -s 10.61.0.0/16 -d 169.254.0.0/16 -j DROP

  # Store the iptables config.
  mkdir -p /etc/iptables
  iptables-save >/etc/iptables/rules.v4
}

## Writes a config file with the default values we use for ignite in the executor.
## This makes it easier to stand up a debugging VM with the same parameters,
## without having to find the three image versions involved here.
function configure_ignite() {
  mkdir -p /etc/ignite
  cat <<EOF >/etc/ignite/config.yaml
apiVersion: ignite.weave.works/v1alpha4
kind: Configuration
metadata:
  name: sourcegraph-executors-default
spec:
  runtime: docker
  networkPlugin: cni
  vmDefaults:
    image:
      oci: "${EXECUTOR_FIRECRACKER_IMAGE}"
    sandbox:
      oci: "${RUNTIME_IMAGE}"
    kernel:
      oci: "${KERNEL_IMAGE}"
      # Explanation of arguments passed here:
      # console: Default
      # reboot: Default
      # panic: Default
      # pci: Default
      # ip: Default
      # random.trust_cpu: Found in https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/random-for-clones.md, this makes RNG initialization much faster (saves ~1s on startup).
      # i8042.X: Makes boot faster, doesn't poll on the i8042 device on boot. See https://github.com/firecracker-microvm/firecracker/blob/main/docs/api_requests/actions.md#intel-and-amd-only-sendctrlaltdel.
      cmdLine: "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp random.trust_cpu=on i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd"
EOF
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
elif [ "${PLATFORM_TYPE}" == "aws" ]; then
  install_cloudwatch_agent
fi
install_docker
install_git
install_src_cli
install_ignite
install_cni
configure_cni
setup_iptables

# Services
install_executor
install_node_exporter

# Service prep and cleanup
generate_ignite_base_image
preheat_kernel_image
configure_ignite
cleanup
