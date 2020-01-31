# Install Sourcegraph with Docker on Google Cloud

This tutorial shows you how to deploy Sourcegraph to a single node running on Google Cloud.

* If you're just starting out, we recommend [running Sourcegraph locally](index.md). It takes only a few minutes and lets you try out all of the features.
* If you need scalability and high-availability beyond what a docker-compose deployment can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), instead.

---

## Deploy to Google Cloud VM

- [Open your Google Cloud console](https://console.cloud.google.com/compute/instances) to create a new VM instance and click **Create Instance**
- Choose an appropriate machine type (we recommend at least the `n1-standard-8` with 8 vCPUs and 30 GB RAM, more depending on team size and number of repositories/languages enabled)
- Choose Ubuntu 18.04 LTS as your boot disk
- Check the boxes for **Allow HTTP traffic** and **Allow HTTPS traffic** in the **Firewall** section
- Open the **Management, disks, networking, and SSH keys** dropdown section and add the following in the **Startup script** field:

  ```bash
  #!/usr/bin/env bash
  
  # Install docker
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
  sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  sudo apt-get update
  apt-cache policy docker-ce
  apt-get install -y docker-ce docker-ce-cli containerd.io

  # Install docker-compose
  curl -L "https://github.com/docker/compose/releases/download/1.25.3/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  chmod +x /usr/local/bin/docker-compose
  curl -L https://raw.githubusercontent.com/docker/compose/1.25.3/contrib/completion/bash/docker-compose -o /etc/bash_completion.d/docker-compose

  # Install git
  sudo apt-get update
  sudo apt-get install -y git

  # Clone docker-compose definition
  git clone https://github.com/sourcegraph/deploy-sourcegraph-docker.git /root/deploy-sourcegraph-docker
  cd /root/deploy-sourcegraph-docker/docker-compose
  git checkout "v3.12.3"

  # Run Sourcegraph. Restart the containers upon reboots.
  docker-compose up -d
  ```

- Create your VM, then navigate to its public IP address.

- If you have configured a DNS entry for the IP, configure `externalURL` to reflect that.

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

## Using an external database for persistence

The Docker container has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the container, you can [use external databases](../../external_database.md) for persistence, such as Google Cloud's [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/) and [Cloud Memorystore](https://cloud.google.com/memorystore/).

> NOTE: Use of external databases requires [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing).
