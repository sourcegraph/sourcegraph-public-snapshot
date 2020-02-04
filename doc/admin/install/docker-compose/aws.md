# Install Sourcegraph with Docker Compose on AWS

This tutorial shows you how to deploy Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) to a single EC2 instance on AWS.

When running Sourcegraph in production, deploying Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) is the default installation method that we recommend. However:

* If you're just starting out, we recommend [running Sourcegraph locally](../docker/index.md). It takes only a few minutes and lets you try out all of the features.
* If you need scalability and high-availability beyond what a single-node [Docker Compose](https://docs.docker.com/compose/) can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), instead.

---

## Deploy to EC2

1. Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home).
1. Select the **Amazon Linux 2 AMI (HVM), SSD Volume Type**.
1. Select an appropriate instance size (we recommend `t2.2xlarge` or larger, depending on team size and number of repositories/languages enabled), then **Next: Configure Instance Details**
1. Ensure the **Auto-assign Public IP** option is "Enable". This ensures your instance is accessible to the Internet.
1. Add the following user data (as text) in the **Advanced Details** section:

   ```yaml
   #cloud-config
   repo_update: true
   repo_upgrade: all

   runcmd:
   # Install, configure, and enable Docker
   - yum update -y
   - amazon-linux-extras install docker
   - systemctl enable --now --no-block docker
   - sed -i -e 's/1024/10240/g' /etc/sysconfig/docker
   - sed -i -e 's/4096/40960/g' /etc/sysconfig/docker
   - usermod -a -G docker ec2-user

   # Install Docker Compose
   - curl -L "https://github.com/docker/compose/releases/download/1.25.3/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   - chmod +x /usr/local/bin/docker-compose
   - curl -L https://raw.githubusercontent.com/docker/compose/1.25.3/contrib/completion/bash/docker-compose -o /etc/bash_completion.d/docker-compose
   
   # Install git, clone Docker Compose definition
   - yum install git -y
   - git clone https://github.com/sourcegraph/deploy-sourcegraph-docker.git /home/ec2-user/deploy-sourcegraph-docker
   - cd /home/ec2-user/deploy-sourcegraph-docker/docker-compose
   - git checkout "v3.12.5"

   # Run Sourcegraph. Restart the containers upon reboot.
   - docker-compose up -d
   ```

1. Select **Next: Add Storage**
1. Select the following settings for the Root volume:

    * **Size (GiB)**: `200` GB minimum *(As a rule of thumb, Sourcegraph needs at least as much space as all your repositories combined take up. Allocating as much disk space as you can upfront helps you avoid [resizing your root volume](https://aws.amazon.com/premiumsupport/knowledge-center/expand-root-ebs-linux/) later on.)*
    * **Volume Type**: General Purpose SSD (gp2)

1. Select **Next: ...** until you get to the **Configure Security Group** page. Then add the following rules:

    * Default **HTTP** rule: port range `80`, source `0.0.0.0/0, ::/0`
    * Default **HTTPS** rule: port range `443`, source `0.0.0.0/0, ::/0`<br>(NOTE: additional work will be required later on to [configure NGINX to support SSL](../../../admin/nginx.md#nginx-ssl-https-configuration))

1. Launch your instance, then navigate to its public IP in your browser. (This can be found by navigating to the instance page on EC2 and looking in the "Description" panel for the "IPv4 Public IP" value.) You may have to wait a minute or two for the instance to finish initializing before Sourcegraph becomes accessible. You can monitor the status by SSHing into the instance and viewing the logs:

    * Following the status of the user data script that you provided earlier:

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
cd /home/ec2-user/deploy-sourcerph-docker/docker-compose
git pull
git checkout vX.Y.Z
docker-compose up -d
```

---

## Storage and Backups

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. These volumes are stored at `/var/lib/docker/volumes` by [default on Linux](https://docs.docker.com/storage/#choose-the-right-type-of-mount). There are a few different back ways to backup this data:

* (**default, recommended**) The most straightfoward method to backup this data is to [snapshot the entire disk that the EC2 instance is using](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html) on an [automatic, scheduled basis](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html).

* Using an external Postgres instance (see below) lets a service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) take care of backing up all of Sourcegraph's user data for you. If the EC2 instance running Sourcegraph ever dies or is destroyed, creating a fresh instance that's connected to that external Postgres will leave Sourcegraph in the same state that it was before.

---

## Using an external database for persistence

The Docker Compose configuration has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the containers, you can [use external databases](../../external_database.md) for persistence, such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) and [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/).

> NOTE: Use of external databases requires [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing).
