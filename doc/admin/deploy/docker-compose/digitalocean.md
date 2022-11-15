---
title: Install Sourcegraph on DigitalOcean
---

# Install Sourcegraph on DigitalOcean

This guide will take you through how to deploy a Sourcegraph instance to a single DigitalOcean Droplet with [Docker Compose](https://docs.docker.com/compose/).

---

## Configure

[Create a new DigitalOcean Droplet](https://cloud.digitalocean.com/droplets/new) first, then configure the droplet following the instructions below for each section:

#### Choose an image

1. Select **Ubuntu 18.04** under *Distributions*

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/188502550-bbab20a0-df2d-4e45-a484-628e022505c9.png"/>

#### Choose a plan

1. Select an appropriate droplet size using our [resource estimator](../resource_estimator.md) as reference

#### Add block storage

1. Click on **Add Volume** to add a new block storage

2. Select size for the block storage --Minimum 250GB
      * Sourcegraph needs at least as much space as all your repositories combined take up
      * Allocating as much disk space as you can upfront minimize the need for switching to a droplet with a larger root disk later on

3. Under **Choose configuration options**, select "Manually Format and Mount"

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/188502606-66bc1301-edbb-493d-b64d-f4a3e6dd0487.png"/>

#### Authentication

1. <span class="badge badge-note">RECOMMENDED</span> Select **SSH keys** to create a **New SSH Key** for convenient access to the droplet

<img class="screenshot w-100" src="https://user-images.githubusercontent.com/68532117/188502682-37667333-75d0-4bd9-8fa8-3b92924c9755.png"/>

#### Authentication > Enable backups

1. <span class="badge badge-note">RECOMMENDED</span> Select **Enable backups** checkbox under *Select additional options* to enable weekly backups of all your data

#### Authentication > User data

1. Copy and paste the *Startup script* below into the **User data** text box:

##### Startup script

```bash
curl -sfL https://raw.githubusercontent.com/sourcegraph/deploy-sourcegraph-docker/master/docker-compose/scripts/install-digitalocean.sh | bash -s - v4.1.3
```

> NOTE: If you're deploying a production instance, we recommend [forking the deployment configuration repository](./index.md#step-1-fork-the-deployment-repository) to track any customizations you make to the deployment config. If you do so, please add the git-clone URL and revision of your fork to the end of the curl command above in the following format:
> `curl -sfL https://example.sh | bash -s - $FORK_REVISION $FORK_CLONE_URL`
> 
> - `$FORK_CLONE_URL`: The Git clone URL of your deployment repository. If it is a private repository, please check with your code host on how to generate a URL for cloning private repository
> - `$FORK_REVISION`: The revision (branch) in your fork containing the customizations, typically "release"

---

## Deploy

1. Click **Create Droplet** to create your droplet with Sourcegraph installed
   - Please ensure the configurations align with the instructions above before creating the instance 

2. Navigate to the droplet's IP address to complete initializing Sourcegraph

>NOTE: It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

You can monitor the setup process by SSHing into the instance to run the following diagnostic commands:

```bash
# Follow the status of the startup script
tail -f /var/log/cloud-init-output.log
# Once installation is completed, check the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

---

## Next

After the initial deployment has been completed, it is strongly recommended to set up the following:

* Restrict the accessibility of ports other than `80` and `443` via [Cloud
  Firewalls](https://www.digitalocean.com/docs/networking/firewalls/quickstart/).
* Set up [TLS/SSL](../../http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2) in the Docker Compose deployment

> NOTE: If you have configured a DNS entry for the IP, please ensure to update `externalURL` in your Sourcegraph instance's Site Configuration to reflect that

---

## Upgrade

Please refer to the [Docker Compose upgrade docs](upgrade.md) for detailed instructions on updating your Sourcegraph instance.

---

## Storage and Backups

Data is persisted within a [Docker volume](https://docs.docker.com/storage/volumes/) as defined in the [deployment repository](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). The startup script configures Docker using a [daemon configuration file](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all the data on the attached data volume, which is mounted at `/mnt/docker-data`, where volumes are stored within `/mnt/docker-data/volumes`.

The most straightforward method to backup this data is to [snapshot the entire `/mnt/docker-data` block storage volume on an automatic scheduled basis](https://www.digitalocean.com/docs/images/snapshots/).

<span class="badge badge-note">RECOMMENDED</span> Using an external Postgres service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) takes care of backing up all the user data for you. If the Sourcegraph instance ever dies or gets destroyed, creating a fresh new instance connected to the old external Postgres service will get Sourcegraph back to its previous state.

---

## Other resources

[HTTP and HTTPS/SSL configuration](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2)
[Site Administration Quickstart](../../../admin/how-to/site-admin-quickstart.md)
