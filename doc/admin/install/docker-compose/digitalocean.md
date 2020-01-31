# Install Sourcegraph with Docker Compose on DigitalOcean

This tutorial shows you how to deploy Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) to a single Droplet running on DigitalOcean.

When running Sourcegraph in production, deploying Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) is the default installation method that we recommend. However:

* If you're just starting out, we recommend [running Sourcegraph locally](../docker/index.md). It takes only a few minutes and lets you try out all of the features.
* If you need scalability and high-availability beyond what a single-node [Docker Compose](https://docs.docker.com/compose/) can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), instead.

---

## Run Sourcegraph on a Digital Ocean Droplet

1. [Create a new Digital Ocean Droplet](https://cloud.digitalocean.com/droplets/new). Set the
   operating system to be Ubuntu 18.04. For droplet size, we recommend at least 8 CPU and 32 GB RAM ,
   but you may need more depending on team size and number of repositories. We recommend you set up
   SSH access (Authentication > SSH keys) for convenient access to the droplet.
1. In the "Select additional options" section of the Droplet creation page, select the "User Data" and "Monitoring" boxes,
   and paste the following script in the "`Enter user data here...`" text box:

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

1. You may have to wait a minute or two for the instance to finish initializing before Sourcegraph becomes accessible. You can monitor the status by SSHing into the Droplet and viewing the logs:

  - Following the status of the user data script that you provided earlier:
    
    ```bash
    tail -f /var/log/cloud-init-output.log
    ```

  - (Once the user data script completes) monitoring the health of the `sourcegraph-frontend` container:

     ```bash
     docker ps --filter="name=sourcegraph-frontend-0"
     ```

1. Navigate to the droplet's IP address to finish initializing Sourcegraph. If you have configured a
   DNS entry for the IP, configure `externalURL` to reflect that.

### After initialization

After initial setup, we recommend you do the following:

* Restrict the accessibility of ports other than `80` and `443` via [Cloud
  Firewalls](https://www.digitalocean.com/docs/networking/firewalls/quickstart/).
* Set up [TLS/SSL](../../nginx.md#nginx-ssl-https-configuration) in the NGINX configuration.

---

## Update your Sourcegraph version

To update to the most recent version of Sourcegraph (X.Y.Z), SSH into your instance and run the following:

```bash
cd /root/deploy-sourcerph-docker/docker-compose
git pull
git checkout vX.Y.Z
docker-compose up -d
```
