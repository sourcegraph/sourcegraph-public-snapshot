# Install Sourcegraph on Azure

This guide will take you through how to set up a Sourcegraph instance on an Azure virtual machine with [Docker Compose](https://docs.docker.com/compose/).

---

## Configure

In the [Azure Quickstart Center](https://portal.azure.com/?quickstart=true#view/Microsoft_Azure_Resources/QuickstartCenterBlade), click `Deploy a virtual machine` to `Create a virtual machine`, then configure the instance following the instructions below for each section:

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

Click `Create and attach a new disk` to create **two** disks:

* **Disk 1** - storage for root
   - `Source type:` None (empty disk)
   - `Size:` 50GB
   - `Performance tier:` 5000 IOS (Recommended)
   - `Enable shared disk:` No
   - `Delete disk with VM:` Checked
   - `Host caching:` Read/write
   - `LUN:` 0
* **Disk 2** - storage for the Sourcegraph instance
   - `Source type:` None (empty disk)
   - `Size:` Minimum 250GB
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

> NOTE: If possible, replace the IP address ranges specified with the IPs from which you actually want to allow access.

#### Management

* <span class="badge badge-note">RECOMMENDED</span> Endable backup

#### Advanced

* Enable `user data`
* In the **Custom data** and **User Data** text boxes, copy and paste the [startup script](#startup-script) from below 

##### Startup script

```bash
curl -sfL https://raw.githubusercontent.com/sourcegraph/deploy-sourcegraph-docker/master/docker-compose/scripts/install-azure.sh | bash -s - v4.1.3
```

> NOTE: If you're deploying a production instance, we recommend [forking the deployment configuration repository](./index.md#step-1-fork-the-deployment-repository) to track any customizations you make to the deployment config. If you do so, please add the git-clone URL and revision of your fork to the end of the curl command above in the following format:
> `curl -sfL https://example.sh | bash -s - $FORK_REVISION $FORK_CLONE_URL`
> 
> - `$FORK_CLONE_URL`: The Git clone URL of your deployment repository. If it is a private repository, please check with your code host on how to generate a URL for cloning private repository
> - `$FORK_REVISION`: The revision (branch) in your fork containing the customizations, typically "release"

---

## Deploy

1. Click **Review + create** to create the instance
  - Please review the configurations and make sure the validation has passed before creating the instance 


2. Navigate to the `public IP address` assigned to your instance to visit your newly created instance
  - Look for the `Public IP address` in your Virtual Machine dashboard under *Networking* in the *Properties* tab

>NOTE: It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible.

You can monitor the setup process by SSHing into the instance to run the following diagnostic commands:

```bash
# Follow the status of the startup script
tail -c +0 -f /var/log/syslog | grep cloud-init
# Once installation is completed, check the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

---

## Upgrade

See the [Docker Compose upgrade docs](upgrade.md).

---

## Storage and Backups

Data is persisted within a [Docker volume](https://docs.docker.com/storage/volumes/) as defined in the [deployment repository](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). The startup script configures Docker using a [daemon configuration file](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all the data on the attached data volume, which is mounted at `/mnt/docker-data`, where volumes are stored within `/mnt/docker-data/volumes`.

The most straightforward method to [backup the data](https://docs.microsoft.com/en-us/azure/virtual-machines/backup-and-disaster-recovery-for-azure-iaas-disks) is to [enable incremental snapshot](https://docs.microsoft.com/en-us/azure/virtual-machines/disks-incremental-snapshots?tabs=azure-cli)

<span class="badge badge-note">RECOMMENDED</span> Using an external Postgres service such as [Azure Database for PostgreSQL](https://learn.microsoft.com/en-us/azure/postgresql/) takes care of backing up all the user data for you. If the Sourcegraph instance ever dies or gets destroyed, creating a fresh new instance connected to the old external Postgres service will get Sourcegraph back to its previous state.

---

## Other resources

[HTTP and HTTPS/SSL configuration](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2)
[Site Administration Quickstart](../../../admin/how-to/site-admin-quickstart.md)
