# Install Sourcegraph on Google Cloud

This guide will take you through how to deploy Sourcegraph with [Docker Compose](https://docs.docker.com/compose/) to a single node running on Google Cloud.

---

## Configure

Click **Create Instance** in your [Google Cloud Compute Engine Console](https://console.cloud.google.com/compute/instances) to create a new VM instance, then configure the instance following the instructions below for each section:

#### Machine configuration
1. Select an appropriate machine type using our [resource estimator](../resource_estimator.md) as reference
  
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

1. Expand the **Advanced options** section and the **Disks** section within to add an additional disk to store data from the Sourcegraph Docker instance.

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

1. Expand the **Advanced options** section and the **Management** section within

2. Copy and paste the *Startup script* below into the **Automation** field

##### Startup script

```bash
curl -sfL https://raw.githubusercontent.com/sourcegraph/deploy-sourcegraph-docker/master/docker-compose/scripts/install-gcp.sh | bash -s - v4.1.3
```

> NOTE: If you're deploying a production instance, we recommend [forking the deployment configuration repository](./index.md#step-1-fork-the-deployment-repository) to track any customizations you make to the deployment config. If you do so, please add the git-clone URL and revision of your fork to the end of the curl command above in the following format:
> `curl -sfL https://example.sh | bash -s - $FORK_REVISION $FORK_CLONE_URL`
> 
> - `$FORK_CLONE_URL`: The Git clone URL of your deployment repository. If it is a private repository, please check with your code host on how to generate a URL for cloning private repository
> - `$FORK_REVISION`: The revision (branch) in your fork containing the customizations, typically "release"

---

## Deploy

1. Click **CREATE** to create your VM with Sourcegraph installed
2. Navigate to the public IP address assigned to your instance to visit your newly created Sourcegraph instance

It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

You can monitor the setup process by SSHing into the instance to run the following diagnostic commands:

```bash
# Follow the status of the startup script
tail -c +0 -f /var/log/syslog | grep startup-script
# Once installation is completed, check the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

---

## Upgrade

Please refer to the [Docker Compose upgrade docs](upgrade.md) for detailed instructions on updating your Sourcegraph instance.

---

## Storage and Backups

Data is persisted within a [Docker volume](https://docs.docker.com/storage/volumes/) as defined in the [deployment repository](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). The startup script configures Docker using a [daemon configuration file](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all the data on the attached data volume, which is mounted at `/mnt/docker-data`, where volumes are stored within `/mnt/docker-data/volumes`.

The most straightforward method to backup the data is to [snapshot the entire `/mnt/docker-data` volume](https://cloud.google.com/compute/docs/disks/create-snapshots) automatically using a [snapshot schedule](https://cloud.google.com/compute/docs/disks/scheduled-snapshots). You can also [set up a snapshot snapshot schedule](https://cloud.google.com/compute/docs/disks/scheduled-snapshots) after your instance has been created.

<span class="badge badge-note">RECOMMENDED</span> Using an external Postgres service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) takes care of backing up all the user data for you. If the Sourcegraph instance ever dies or gets destroyed, creating a fresh new instance connected to the old external Postgres service will get Sourcegraph back to its previous state.

---

## Other resources

[HTTP and HTTPS/SSL configuration](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2)
[Site Administration Quickstart](../../../admin/how-to/site-admin-quickstart.md)
