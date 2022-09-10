# Install Sourcegraph on Amazon Web Services (AWS)

This guide will take you through how to deploy Sourcegraph with [Docker Compose](https://docs.docker.com/compose/) to a single EC2 instance on Amazon Web Services (AWS).

## Prerequisites

- An AWS account
- Determine your instance size and resource requirements using the [resource estimator](../resource_estimator.md)
- <span class="badge badge-note">RECOMMENDED</span> [Create your own customized copy of the deployment repository](../index.md#installation)

---

## Configure

Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home), then configure the instance following the instructions below for each section:

#### Name and tags

1. Name your instance

#### Application and OS Images

1. Select **Amazon Linux** in the *Quick Start* tab

2. Select **Amazon Linux 2 AMI (HVM), SSD Volume Type** under *Amazon Machine Image (AMI)*

#### Instance type

1. Select an appropriate instance type using our [resource estimator](../resource_estimator.md) as reference

#### Key pair (login)

1. Create a new key pair for your instance, or choose an existing key pair from the drop down list

#### Network settings

1. Click `Edit` in the header to enable **Auto-assign Public IP** 

2. Under **Firewall (security group)** , create or select existing security group with the following settings:

  * Allow SSH traffic from Anywhere
  * Allow HTTPs traffic from the internet
  * Allow HTTP traffic from the internet

> NOTE: If possible, replace the IP address ranges specified with the IPs from which you actually want to allow access.

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

<span class="badge badge-warning">IMPORTANT</span> **Required for users deploying with a customized copy of the deployment repository:**

- Update the *startup script* with the information of your deployment repository:
  - `DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL`: The git clone URL of your deployment repository
  - `DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION`: The git revision (branch) containing your customizations

```bash
#!/usr/bin/env bash
set -euxo pipefail
###############################################################################
# ACTION REQUIRED: REPLACE THE URL AND REVISION WITH YOUR DEPLOYMENT REPO INFO
###############################################################################
DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL='https://github.com/sourcegraph/deploy-sourcegraph-docker.git'
DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v3.43.2'
##################### NO CHANGES REQUIRED BELOW THIS LINE #####################
DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT='/home/ec2-user/deploy-sourcegraph-docker'
DOCKER_COMPOSE_VERSION='1.29.2'
DOCKER_DAEMON_CONFIG_FILE='/etc/docker/daemon.json'
DOCKER_DATA_ROOT='/mnt/docker-data'
EBS_VOLUME_DEVICE_NAME='/dev/sdb'
EBS_VOLUME_LABEL='sourcegraph'
# Install git
yum update -y
yum install git -y
# Clone the deployment repository
git clone "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL}" "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
git checkout "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION}"
# Format (if unformatted) and then mount the attached volume
device_fs=$(lsblk "${EBS_VOLUME_DEVICE_NAME}" --noheadings --output fsType)
if [ "${device_fs}" == "" ]
then
  mkfs -t xfs "${EBS_VOLUME_DEVICE_NAME}"
  xfs_admin -L "${EBS_VOLUME_LABEL} ${EBS_VOLUME_DEVICE_NAME}"
fi
mkdir -p "${DOCKER_DATA_ROOT}"
mount "${EBS_VOLUME_DEVICE_NAME}" "${DOCKER_DATA_ROOT}"
# Mount file system by label on reboot
echo "LABEL=${EBS_VOLUME_LABEL}  ${DOCKER_DATA_ROOT}  xfs  defaults,nofail  0  2" >> '/etc/fstab'
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
## Initialize the config file with empty json if it doesn't exist
if [ ! -f "${DOCKER_DAEMON_CONFIG_FILE}" ]
then
  mkdir -p $(dirname "${DOCKER_DAEMON_CONFIG_FILE}")
  echo '{}' > "${DOCKER_DAEMON_CONFIG_FILE}"
fi
## Point Docker storage to mounted volume
tmp_config=$(mktemp)
trap "rm -f ${tmp_config}" EXIT
cat "${DOCKER_DAEMON_CONFIG_FILE}" | jq --arg DATA_ROOT "${DOCKER_DATA_ROOT}" '.["data-root"]=$DATA_ROOT' > "${tmp_config}"
cat "${tmp_config}" > "${DOCKER_DAEMON_CONFIG_FILE}"
# Restart Docker daemon to pick up new changes
systemctl restart --now docker
# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
curl -L "https://raw.githubusercontent.com/docker/compose/${DOCKER_COMPOSE_VERSION}/contrib/completion/bash/docker-compose" -o /etc/bash_completion.d/docker-compose
# Start Sourcegraph with Docker Compose
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"/docker-compose
docker-compose up -d --remove-orphans
```

## Deploy

1. Click **Launch Instance** in the *Summary Section* on the right to create your Sourcegraph instance
   - Please ensure the configurations align with the instructions above before creating the instance 

2. Navigate to the public IP address assigned to your instance to visit your newly created Sourcegraph instance
      - Look for the **IPv4 Public IP** value in your EC2 instance page under the *Description* panel

It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

You can monitor the setup process by SSHing into the instance to run the following diagnostic commands:

```bash
# Follow the status of the startup script
tail -f /var/log/cloud-init-output.log
# Once installation is completed, check the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

## Next

After the initial deployment has been completed, it is strongly recommended to set up the following:

* Restrict the accessibility of ports other than `80` and `443` via [Cloud
  Firewalls](https://www.digitalocean.com/docs/networking/firewalls/quickstart/).
* Set up [TLS/SSL](../../http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2) in the Docker Compose deployment

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

## Upgrade

Please refer to the [Docker Compose upgrade docs](upgrade.md) for detailed instructions.

## Storage and Backups

Data is persisted within a [Docker volume](https://docs.docker.com/storage/volumes/) as defined in the [deployment repository](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). The startup script configures Docker using a [daemon configuration file](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all the data on the attached data volume, which is mounted at `/mnt/docker-data`, where volumes are stored within `/mnt/docker-data/volumes`.

The most straightforward method to backup this data is to [snapshot the entire `/mnt/docker-data` EBS disk](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html) on an [automatic, scheduled basis](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html).

<span class="badge badge-note">RECOMMENDED</span> Using an external Postgres service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) takes care of backing up all the user data for you. If the Sourcegraph instance ever dies or gets destroyed, creating a fresh new instance connected to the old external Postgres service will get Sourcegraph back to its previous state.
