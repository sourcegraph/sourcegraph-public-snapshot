# Install Sourcegraph with Docker Compose on Google Cloud

This tutorial shows you how to deploy Sourcegraph via [Docker Compose](https://docs.docker.com/compose/)to a single node running on Google Cloud.

* If you're just starting out, we recommend [running Sourcegraph locally](index.md). It takes only a few minutes and lets you try out all of the features.
* If you need scalability and high-availability beyond what a docker-compose deployment can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), instead.

---

## Deploy to Google Cloud VM

1. [Open your Google Cloud console](https://console.cloud.google.com/compute/instances) to create a new VM instance and click **Create Instance**
1. Choose an appropriate machine type (we recommend at least the `n1-standard-8` with `8` vCPUs and `30` GB RAM, more depending on team size and number of repositories/languages enabled)
1. Under the "Boot Disk" options, select the following:

    * **Operating System**: Ubuntu
    * **Version**: Ubuntu 18.04 LTS
    * **Boot disk type**: SSD persistent disk
    * **Size**: `200` GB minimum *(As a rule of thumb, Sourcegraph needs at least as much space as all your repositories combined take up. Allocating as much disk space as you can upfront helps you avoid [resizing your boot disk](https://cloud.google.com/compute/docs/disks/add-persistent-disk#resize_pd) later on.)*

1. Check the boxes for **Allow HTTP traffic** and **Allow HTTPS traffic** in the **Firewall** section
1. Open the **Management, disks, networking, and SSH keys** dropdown section and add the following in the **Startup script** field:

    ```bash
    #!/usr/bin/env bash

    set -euxo pipefail

    DOCKER_COMPOSE_VERSION='1.25.3'
    SOURCEGRAPH_VERSION='v3.12.5'
    DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT='/root/deploy-sourcegraph-docker'

    # Install Docker
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    sudo apt-get update
    apt-cache policy docker-ce
    apt-get install -y docker-ce docker-ce-cli containerd.io

    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    curl -L https://raw.githubusercontent.com/docker/compose/${DOCKER_COMPOSE_VERSION}/contrib/completion/bash/docker-compose -o /etc/bash_completion.d/docker-compose

    # Install git
    sudo apt-get update
    sudo apt-get install -y git

    # Clone Docker Compose definition
    git clone https://github.com/sourcegraph/deploy-sourcegraph-docker.git ${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}
    cd ${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}/docker-compose
    git checkout ${SOURCEGRAPH_VERSION}

    # Run Sourcegraph. Restart the containers upon reboot.
    docker-compose up -d
    ```

1. Create your VM, then navigate to its public IP address.
1. If you have configured a DNS entry for the IP, configure `externalURL` to reflect that.
1. You may have to wait a minute or two for the instance to finish initializing before Sourcegraph becomes accessible. You can monitor the status by SSHing into the instance and:
    * Following the status of the startup script that you provided earlier:

      ```bash
      tail -f /var/log/cloud-init-output.log
      ```

    * (Once the user data script completes) monitoring the health of the `sourcegraph-frontend` container:

      ```bash
      docker ps --filter="name=sourcegraph-frontend-0"
      ```

---

## Update your Sourcegraph version

To update to the most recent version of Sourcegraph (X.Y.Z), SSH into your instance and run the following:

```bash
cd /root/deploy-sourcerph-docker/docker-compose
git pull
git checkout vX.Y.Z
docker-compose up -d
```

---

## Storage and Backups

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. These volumes are stored at `/var/lib/docker/volumes` by [default on Linux](https://docs.docker.com/storage/#choose-the-right-type-of-mount). There are a few different back ways to backup this data:

* (**default, recommended**) The most straightfoward method backup to backup this data is to [snapshot the entire disk that the GCP instance is using](https://cloud.google.com/compute/docs/disks/create-snapshots) on an [automatic, scheduled basis](https://cloud.google.com/compute/docs/disks/scheduled-snapshots).

* Using an external Postgres instance (see below) lets a service such as [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/) take care of backing up all of Sourcegraph's user data for you. If the VM instance running Sourcegraph ever dies or is destroyed, creating a fresh instance that's connected to that external Postgres will leave Sourcegraph in the same state that it was before.

---

## Using an external database for persistence

The Docker Compose configuration has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the containers, you can [use external databases](../../external_database.md) for persistence, such as Google Cloud's [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/) and [Cloud Memorystore](https://cloud.google.com/memorystore/).

> NOTE: Use of external databases requires [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing).
