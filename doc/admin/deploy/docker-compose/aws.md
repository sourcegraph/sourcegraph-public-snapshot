# Install Sourcegraph on Amazon Web Services (AWS)

This guide will take you through how to deploy Sourcegraph with [Docker Compose](https://docs.docker.com/compose/) to a single EC2 instance on Amazon Web Services (AWS).

## Prerequisites

- Determine the instance type and resource requirements for your Sourcegraph instance referring to the [resource estimator](../resource_estimator.md)
- **[RECOMMENDED]** Follow Step 1 to 5 in our [Docker Compose installation guide](https://docs.sourcegraph.com/admin/deploy/docker-compose#installation) to prepare a fork of the Sourcegraph Docker Compose deployment repository with `release branch` set up

---

## Configuration

Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home), then configure the instance following the instructions below for each section:

#### Name and tags
1. Name your instance

#### Application and OS Images
1. Select **Amazon Linux** in the *Quick Start* tab
2. Select **Amazon Linux 2 AMI (HVM), SSD Volume Type** under *Amazon Machine Image (AMI)*

#### Instance type

1. Select an appropriate instance type
   * Refer to the [resource estimator](../resource_estimator.md) to find a good starting point for your deployment

#### Key pair (login)

1. Create a new key pair for your instance, or choose an existing key pair from the drop down list.

#### Network settings

1. Click `Edit` in the header to enable **Auto-assign Public IP** . This ensures your instance is accessible to the Internet.
2. Under **Firewall (security group)** , create or select existing security group with the following settings:
  * Allow SSH traffic from Anywhere
  * Allow HTTPs traffic from the internet
      * Default **HTTPS** rule: port range `443`, source `0.0.0.0/0, ::/0`
  * Allow HTTP traffic from the internet
      * Default **HTTP** rule: port range `80`, source `0.0.0.0/0, ::/0`
3. Additional work will be required later on to [configure SSL in the Docker Compose deployment](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2))

> WARNING: While this port configuration will work, it provides open access of the ports specified. If possible, replace the IP address ranges specified with the IPs from which you actually want to allow access.

#### Configure storage

1. Click **Add New Volume** to add an *additional* EBS volume for storing data
2. Click **Advanced** in the header to update the following settings for the new Custom Volume:
  * `Storage Type`: EBS
  * `Device name`: `/dev/sdb`
  * `Volume Type`: `gp3` (General Purpose SSD)
  * `Size (GiB)`: `250GB minimum`
      * Sourcegraph needs at least as much space as all your repositories combined take up
      * Allocating as much disk space as you can upfront minimize the need for [resizing your volume](https://aws.amazon.com/premiumsupport/knowledge-center/expand-root-ebs-linux/) in the future
  * `Delete on Termination`: `No`

#### Advanced details > User Data

1. Copy and paste the *Startup script* below in the **User Data** text box at the bottom
2. **[RECOMMENDED]** Update the *startup script* with the information of your fork and release branch if deploying from a fork of the reference repository
    * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL`: The git clone URL of your fork
    * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION`: The git revision containing your fork's customizations to the base Sourcegraph Docker Compose YAML. In the [example](index.md#step-3-configure-the-release-branch) the revision is the `release` branch

```bash
#!/usr/bin/env bash

set -euxo pipefail

# 🚨 Update these variables with the correct values from your fork!
DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL='https://github.com/sourcegraph/deploy-sourcegraph-docker.git'
DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v3.43.1'

# IMPORTANT: DO NOT MAKE ANY CHANGES FROM THIS POINT ONWARD
EBS_VOLUME_DEVICE_NAME='/dev/sdb'
DOCKER_DATA_ROOT='/mnt/docker-data'

DOCKER_COMPOSE_VERSION='1.29.2'
DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT='/home/ec2-user/deploy-sourcegraph-docker'

# Install git
yum update -y
yum install git -y

# Clone Docker Compose definition
git clone "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL}" "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
git checkout "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION}"

# Format (if necessary) and mount EBS volume
device_fs=$(lsblk "${EBS_VOLUME_DEVICE_NAME}" --noheadings --output fsType)
if [ "${device_fs}" == "" ] ## only format the volume if it isn't already formatted
then
  mkfs -t xfs "${EBS_VOLUME_DEVICE_NAME}"
fi
mkdir -p "${DOCKER_DATA_ROOT}"
mount "${EBS_VOLUME_DEVICE_NAME}" "${DOCKER_DATA_ROOT}"

# Mount EBS volume on reboots
EBS_UUID=$(blkid -s UUID -o value "${EBS_VOLUME_DEVICE_NAME}")
echo "UUID=${EBS_UUID}  ${DOCKER_DATA_ROOT}  xfs  defaults,nofail  0  2" >> '/etc/fstab'
umount "${DOCKER_DATA_ROOT}"
mount -a

# Install, configure, and enable Docker
yum update -y
amazon-linux-extras install docker
systemctl enable --now docker
sed -i -e 's/1024/262144/g' /etc/sysconfig/docker
sed -i -e 's/4096/262144/g' /etc/sysconfig/docker
usermod -a -G docker ec2-user

# Install jq for scripting
yum install -y jq

# Edit Docker storage directory to mounted volume
DOCKER_DAEMON_CONFIG_FILE='/etc/docker/daemon.json'

## initialize the config file with empty json if it doesn't exist
if [ ! -f "${DOCKER_DAEMON_CONFIG_FILE}" ]
then
  mkdir -p $(dirname "${DOCKER_DAEMON_CONFIG_FILE}")
  echo '{}' > "${DOCKER_DAEMON_CONFIG_FILE}"
fi

## update Docker's 'data-root' to point to our mounted disk
tmp_config=$(mktemp)
trap "rm -f ${tmp_config}" EXIT
cat "${DOCKER_DAEMON_CONFIG_FILE}" | jq --arg DATA_ROOT "${DOCKER_DATA_ROOT}" '.["data-root"]=$DATA_ROOT' > "${tmp_config}"
cat "${tmp_config}" > "${DOCKER_DAEMON_CONFIG_FILE}"

## finally, restart Docker daemon to pick up our changes
systemctl restart --now docker

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
curl -L "https://raw.githubusercontent.com/docker/compose/${DOCKER_COMPOSE_VERSION}/contrib/completion/bash/docker-compose" -o /etc/bash_completion.d/docker-compose

# Run Sourcegraph. Restart the containers upon reboot.
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"/docker-compose
docker-compose up -d
```

![image](https://user-images.githubusercontent.com/68532117/188503178-6072319e-74c8-439c-b270-cb32ef8ff7df.png)

## Deploy

1. Click **Launch Instance** in the *Summary Section* on the right to create your Sourcegraph instance
2. Navigate to the public IP address assigned to your instance to visit your newly created Sourcegraph instance
      - Look for the **IPv4 Public IP** value in your EC2 instance page under the *Description* panel

It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

You can monitor the status by SSHing into the instance and running the following diagnostic commands:

```bash
# Follow the status of the user data script you provided earlier
tail -f /var/log/cloud-init-output.log

# (Once the user data script completes) monitor the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

---

## Upgrade

Please refer to the [Docker Compose upgrade docs](upgrade.md) for detailed instructions on updating your Sourcegraph instance.

## Storage and Backups

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. The previous script [configures Docker](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all Docker data on the additional EBS volume that was attached to the instance (mounted at `/mnt/docker-data` - the volumes themselves are stored under `/mnt/docker-data/volumes`) There are a few different ways to backup this data:

* (**recommended**) The most straightforward method to backup this data is to [snapshot the entire `/mnt/docker-data` EBS disk](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html) on an [automatic, scheduled basis](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html).

* Using an external Postgres instance lets a service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) take care of backing up all of Sourcegraph's user data for you. If the EC2 instance running Sourcegraph ever dies or is destroyed, creating a fresh instance that's connected to that external Postgres will leave Sourcegraph in the same state that it was before.
