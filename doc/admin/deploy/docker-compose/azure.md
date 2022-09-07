# Install Sourcegraph on Azure

This guide will take you through how to set up a Sourcegraph instance on an Azure virtual machine with [Docker Compose](https://docs.docker.com/compose/).

## Prerequisites

- Determine the instance type and resource requirements for your Sourcegraph instance referring to the [resource estimator](../resource_estimator.md)
- <span class="badge badge-critical">RECOMMENDED</span> Follow our [Docker Compose installation guide](https://docs.sourcegraph.com/admin/deploy/docker-compose#installation) to create your own customized copy of the Sourcegraph Docker Compose deployment repository with `release branch` set up

---

## Configuration

In the [Azure Quickstart Center](https://portal.azure.com/?quickstart=true#view/Microsoft_Azure_Resources/QuickstartCenterBlade), click `Deploy a virtual machine` to `Create a virtual machine`, then configure the instance as suggested below for each section:

> NOTE: Please use the default values for items that are not covered below.

#### Basics

* `Virtual machine name:` Give your virtual machine a name
* `Availability options:` No infrastructure redundancy required
* `Image:` Ubuntu Server 18.04 LTS - Gen2
* `VM architecture:` x64
* `Size:` Select an appropriate instance type using our [resource estimator](../resource_estimator.md) as reference
* `Authentication type:` Select one that works best for you. SSH Key is recommended.
* `Inbound port rules:` Allowed selected ports
* `Select inbound ports:` HTTP (80), HTTPS (443), SSH (22)

#### Disks

* `OS disk type:` SSD is required --Premium SSD (Recommended) or Standard SSD
* `Delete with VM:` Unchecked

#### Disks > Data disks

Click `Create and attach a new disk` to create two disks:

* **Disk 1** - storage for root
   - `Source type:` None (empty disk)
   - `Size:` Minimum `16GB`
   - `Performance tier:` 5000 IOS (Recommended)
   - `Enable shared disk:` No
   - `Delete disk with VM:` Checked
   - `Host caching:` Read/write
   - `LUN:` 0
* **Disk 2** - storage for the Sourcegraph instance
   - `Source type:` None (empty disk)
   - `Size:` Minimum `256GB`
         * Sourcegraph needs at least as much space as all your repositories combined take up
         * Allocating as much disk space as you can upfront minimize the need for [expanding your volume](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/expand-disks) in the future
   - `Performance tier:` 5000 IOS (Recommended)
   - `Enable shared disk:` No
   - `Delete disk with VM:` Unchecked
   - `Host caching:` Read/write
   - `LUN:` 1

> NOTE: Typically, the device name for the `LUN: 0` disk is `dev/sda` while the device name for the `LUN: 1` disk is `dev/sdb` --this is important to note as our startup script mounts the disks based on device names (`PERSISTENT_DISK_DEVICE_NAME`).

#### Networking

* `Inbound port rules:` Allowed selected ports
* `Select inbound ports:` HTTP (80), HTTPS (443), SSH (22)

>NOTE: Additional work will be required later on to [configure SSL in the Docker Compose deployment](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2))


#### Management

* Endable backup - Recommended

#### Advanced

* Enable `user data`
* Copy and paste the *Startup script* below in the **Custom data** and **User Data** text boxes
  * <span class="badge badge-warning">IMPORTANT</span> Update the *startup script* with the information of your fork and release branch **if deploying from a fork of the reference repository**
    * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL`: The git clone URL of your fork
    * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION`: The git revision containing your fork's customizations to the base Sourcegraph Docker Compose YAML. In the [example](index.md#step-3-configure-the-release-branch) the revision is the `release` branch

```bash
#!/usr/bin/env bash

set -euxo pipefail

# Update these variables with the correct values from your fork!
DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL='https://github.com/sourcegraph/deploy-sourcegraph-docker.git'
DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v3.43.1'

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
docker-compose up -d --remove-orphans
```

## Deploy

1. Click **Review + create** to review the configurations and make sure the validation has passed before creating the instance 
1. Navigate to the public IP address assigned to your instance to visit your newly created Sourcegraph instance
  - Look for the **Public IP address** in your Virtual Machine dashboard under *Networking* in the *Properties* tab

It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

---

## Upgrade

Please refer to the [Docker Compose upgrade docs](upgrade.md) for detailed instructions on updating your Sourcegraph instance.

## Storage and Backups

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. The startup script [configures Docker](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all Docker data on the disk that is attached to the instance (mounted at `/mnt/docker-data` - the volumes themselves are stored under `/mnt/docker-data/volumes`).


* (**recommended**) The most straightforward method to [backup the data](https://docs.microsoft.com/en-us/azure/virtual-machines/backup-and-disaster-recovery-for-azure-iaas-disks) or [enable incremental snapshot](https://docs.microsoft.com/en-us/azure/virtual-machines/disks-incremental-snapshots?tabs=azure-cli).

* Using an external Postgres instance lets a service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) take care of backing up all of Sourcegraph's user data for you. If the EC2 instance running Sourcegraph ever dies or is destroyed, creating a fresh instance that's connected to that external Postgres will leave Sourcegraph in the same state that it was before.
