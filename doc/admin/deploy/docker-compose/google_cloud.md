# Install Sourcegraph on Google Cloud

This guide will take you through how to deploy Sourcegraph with [Docker Compose](https://docs.docker.com/compose/) to a single node running on Google Cloud.

## Prerequisites

- Determine the instance type and resource requirements for your Sourcegraph instance referring to the [resource estimator](../resource_estimator.md)
- **[RECOMMENDED]** Follow Step 1 to 5 in our [Docker Compose installation guide](https://docs.sourcegraph.com/admin/deploy/docker-compose#installation) to prepare a fork of the Sourcegraph Docker Compose deployment repository with `release branch` set up

---

## Configuration

Click **Create Instance** in your [Google Cloud Compute Engine Console](https://console.cloud.google.com/compute/instances) to create a new VM instance, then configure the instance following the instructions below for each section:

#### Machine configuration
1. Choose an appropriate **machine type**
    * Refer to the [resource estimator](../resource_estimator.md) to find a good starting point for your deployment
  
#### Boot disk
1. Click **CHANGE** to update the boot disk:
   * `Operating System`: Ubuntu
   * `Version`: Ubuntu 18.04 LTS (x86/64, amd64 bionic image)
   * `Boot disk type`: SSD persistent disk
   * `Size (GB)`: Use default

#### Firewall
1. Check box to **Allow HTTP traffic**
1. Check box to **Allow HTTPS traffic**

#### Advanced options > Disks

Expand the **Advanced options** section and the **Disks** section within to add an additional disk to store data from the Sourcegraph Docker instance.

1. Click **+ ADD NEW DISK** to setup the new disk with the following settings:
  * `Name`: "sourcegraph-docker-disk" (or something similarly descriptive)
  * `Description`: "Disk for storing Docker data for Sourcegraph" (or something similarly descriptive)
  * `Disk source type`: Blank disk
  * `Disk type`: SSD persistent disk
  * `Size`: `250GB` minimum
      * Sourcegraph needs at least as much space as all your repositories combined take up
      * Allocating as much disk space as you can upfront minimize the need for [resizing this disk](https://cloud.google.com/compute/docs/disks/add-persistent-disk#resize_pd) later on
  * `(optional, recommended) Snapshot schedule`: The most straightfoward way of automatically backing Sourcegraph's data is to set up a [snapshot schedule](https://cloud.google.com/compute/docs/disks/scheduled-snapshots) for this disk. We strongly recommend that you take the time to do so here.
  * `Attachment settings - Mode`: Read/write
  * `Attachment settings - Deletion rule`: Keep disk

#### Advanced options > Management
Expand the **Advanced options** section and the **Management** section within:

![image](https://user-images.githubusercontent.com/68532117/188502993-05ef0eb2-ceb5-47bb-a6b6-177839149dd5.png)

1. Copy and paste the *Startup script* in the **Automation** field
2. **[RECOMMENDED]** Update the *startup script* with the information of your fork and release branch if deploying from a fork of the reference repository
    * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL`: The git clone URL of your fork
    * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION`: The git revision containing your fork's customizations to the base Sourcegraph Docker Compose YAML. In the [example](index.md#step-3-configure-the-release-branch) the revision is the `release` branch

```bash
#!/usr/bin/env bash

set -euxo pipefail

# 🚨 Update these variables with the correct values from your fork
DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL='https://github.com/sourcegraph/deploy-sourcegraph-docker.git'
DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v3.43.1'

# IMPORTANT: DO NOT MAKE ANY CHANGES FROM THIS POINT ONWARD
PERSISTENT_DISK_DEVICE_NAME='/dev/sdb'
DOCKER_DATA_ROOT='/mnt/docker-data'

DOCKER_COMPOSE_VERSION='1.29.2'
DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT='/root/deploy-sourcegraph-docker'

# Install git
sudo apt-get update -y
sudo apt-get install -y git

# Clone Docker Compose definition
git clone "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL}" "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
git checkout "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION}"

# Format (if necessary) and mount GCP persistent disk
device_fs=$(sudo lsblk "${PERSISTENT_DISK_DEVICE_NAME}" --noheadings --output fsType)
if [ "${device_fs}" == "" ] ## only format the volume if it isn't already formatted
then
    sudo mkfs.ext4 -m 0 -E lazy_itable_init=0,lazy_journal_init=0,discard "${PERSISTENT_DISK_DEVICE_NAME}"
fi
sudo mkdir -p "${DOCKER_DATA_ROOT}"
sudo mount -o discard,defaults "${PERSISTENT_DISK_DEVICE_NAME}" "${DOCKER_DATA_ROOT}"

# Mount GCP disk on reboots
DISK_UUID=$(sudo blkid -s UUID -o value "${PERSISTENT_DISK_DEVICE_NAME}")
sudo echo "UUID=${DISK_UUID}  ${DOCKER_DATA_ROOT}  ext4  discard,defaults,nofail  0  2" >> '/etc/fstab'
umount "${DOCKER_DATA_ROOT}"
mount -a

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo apt-get update -y
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update -y
apt-cache policy docker-ce
apt-get install -y docker-ce docker-ce-cli containerd.io

# Install jq for scripting
sudo apt-get update -y
sudo apt-get install -y jq

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
sudo cat "${DOCKER_DAEMON_CONFIG_FILE}" | sudo jq --arg DATA_ROOT "${DOCKER_DATA_ROOT}" '.["data-root"]=$DATA_ROOT' > "${tmp_config}"
sudo cat "${tmp_config}" > "${DOCKER_DAEMON_CONFIG_FILE}"

## finally, restart Docker daemon to pick up our changes
sudo systemctl restart --now docker

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
curl -L "https://raw.githubusercontent.com/docker/compose/${DOCKER_COMPOSE_VERSION}/contrib/completion/bash/docker-compose" -o /etc/bash_completion.d/docker-compose

# Run Sourcegraph. Restart the containers upon reboot.
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"/docker-compose
docker-compose up -d
```

## Deploy

1. Click **CREATE** to create your VM with Sourcegraph installed
2. Navigate to the public IP address assigned to your instance to visit your newly created Sourcegraph instance

It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

You can monitor the status by SSHing into the instance and running the following diagnostic commands:

```bash
# Follow the status of the user data script you provided earlier
tail -c +0 -f /var/log/syslog | grep startup-script

# (Once the user data script completes) monitor the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

---

## Upgrade

Please refer to the [Docker Compose upgrade docs](upgrade.md) for detailed instructions on updating your Sourcegraph instance.

## Storage and Backups

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. The script above [configures Docker](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all Docker data on the additional persistent disk that was attached to the instance (mounted at `/mnt/docker-data` - the volumes themselves are stored under `/mnt/docker-data/volumes`) There are a few different ways to backup this data:

* (**recommended**) The most straightfoward method to backup this data is to [snapshot the entire `/mnt/docker-data` persistsent disk](https://cloud.google.com/compute/docs/disks/create-snapshots) on an [automatic, scheduled basis](https://cloud.google.com/compute/docs/disks/scheduled-snapshots). The directions above tell you how to set up this schedule when the instance is first created, but you can also [create the schedule afterwards](https://cloud.google.com/compute/docs/disks/scheduled-snapshots).

* Using an external Postgres instance (see below) lets a service such as [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/) take care of backing up all of Sourcegraph's user data for you. If the VM instance running Sourcegraph ever dies or is destroyed, creating a fresh instance that's connected to that external Postgres will leave Sourcegraph in the same state that it was before.
