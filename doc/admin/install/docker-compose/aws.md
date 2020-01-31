# Install Sourcegraph with Docker Compose on AWS

This tutorial shows you how to deploy Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) to a single EC2 instance on AWS.

When running Sourcegraph in production, deploying Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) is the default installation method that we recommend. However:

* If you're just starting out, we recommend [running Sourcegraph locally](../docker/index.md). It takes only a few minutes and lets you try out all of the features.
* If you need scalability and high-availability beyond what a single-node [Docker Compose](https://docs.docker.com/compose/) can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), instead.

---

## Deploy to EC2

- Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home).
- Select the Amazon Linux 2 AMI (HVM), SSD Volume Type.
- Select an appropriate instance size (we recommend `t2.2xlarge` or larger, depending on team size and number of repositories/languages enabled), then **Next: Configure Instance Details**
- Ensure the **Auto-assign Public IP** option is "Enable". This ensures your instance is accessible to the Internet.
- Add the following user data (as text) in the **Advanced Details** section:

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

- Select **Next: ...** until you get to the **Configure Security Group** page. Then add the following rules:
  - Default **HTTP** rule: port range `80`, source `0.0.0.0/0, ::/0`
  - Default **HTTPS** rule: port range `443`, source `0.0.0.0/0, ::/0`<br>(NOTE: additional work will be required later on to [configure NGINX to support SSL](../../../admin/nginx.md#nginx-ssl-https-configuration))
- Launch your instance, then navigate to its public IP in your browser. (This can be found by navigating to the instance page on EC2 and looking in the "Description" panel for the "IPv4 Public IP" value.) 1. You may have to wait a minute or two for the instance to finish initializing before Sourcegraph becomes accessible. You can monitor the status by SSHing into the instance and viewing the logs:

  - Following the status of the user data script that you provided earlier:
    
    ```bash
    tail -f /var/log/cloud-init-output.log
    ```

  - (Once the user data script completes) monitoring the health of the `sourcegraph-frontend` container:

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

## Using an external database for persistence

The Docker container has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the container, you can [use external databases](../../external_database.md) for persistence, such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) and [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/).

> NOTE: Use of external databases requires [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing).
