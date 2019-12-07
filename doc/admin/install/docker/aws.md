# Install Sourcegraph with Docker on AWS

This tutorial shows you how to deploy Sourcegraph to a single node running on AWS.

If you're just starting out, we recommend [installing Sourcegraph locally](index.md). It takes only a few minutes and lets you try out all of the features. If you need scalability and high-availability beyond what a single-server deployment can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph).

---

## Deploy to EC2

> NOTE: While these instructions recommend opening up port `443`, additional work is required to [configure NGINX to support SSL](../../../admin/nginx.md#nginx-ssl-https-configuration).

### Option A: use the AWS wizard

- Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home).
- Select the Amazon Linux 2 AMI (`ami-032509850cf9ee54e` at this time of writing).
- Select an appropriate instance size (we recommend t2.medium/large, depending on team size and number of repositories/languages enabled), then **Next: Configure Instance Details**
- Add the following user data (as text) in the **Advanced Details** section:

  ```yaml
  #cloud-config
  repo_update: true
  repo_upgrade: all

  runcmd:
  # Create the directory structure for Sourcegraph data
  - mkdir -p /home/ec2-user/.sourcegraph/config
  - mkdir -p /home/ec2-user/.sourcegraph/data

  # Install, configure, and enable Docker
  - yum update -y
  - amazon-linux-extras install docker
  - systemctl enable --now --no-block docker
  - sed -i -e 's/1024/10240/g' /etc/sysconfig/docker
  - sed -i -e 's/4096/40960/g' /etc/sysconfig/docker
  - usermod -a -G docker ec2-user

  # Install and run Sourcegraph. Restart the container upon subsequent reboots
  - [ sh, -c, 'docker run -d --publish 80:7080 --publish 443:7080 --publish 2633:2633 --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:3.10.2' ]
  ```

- Select **Next: ...** until you get to the **Configure Security Group** page, then add the default **HTTP** rule (port range "80", source "0.0.0.0/0, ::/0")
- Launch your instance, then navigate to the its public URL.
- If you have configured a DNS entry for the IP, configure `externalURL` to reflect that. (Note: `externalURL` was called `appURL` in Sourcegraph 2.13 and earlier.)

### Option B: use the CLI

Use the [`aws` CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html) to boot an EC2 instance running Sourcegraph.

First, create a `cloud-init.txt` file with user data contents as shown above or below. Then run:

`aws ec2 run-instances --image-id ami-032509850cf9ee54e --count 1 --instance-type t2.medium --key-name id_rsa --security-groups default --user-data file://cloud-init.txt`

Substitute the path to your `cloud-init.txt` file, the name of your key pair, and an appropriate security group. To start you probably want a security group which exposes port 80, 443, 2633 (for the management console), and 22 (for SSH) to the public internet.

---

## Update your Sourcegraph version

To update to the most recent version of Sourcegraph (X.Y.Z), SSH into your instance and run the following:

```bash
docker ps # get the $CONTAINER_ID of the running sourcegraph/server container
docker rm -f $CONTAINER_ID
docker run docker run -d --publish 80:7080 --publish 443:7080 --publish 2633:2633 --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:X.Y.Z
```

---

## Using an external database for persistence

The Docker container has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the container, you can [use external databases](../../external_database.md) for persistence, such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) and [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/).

> NOTE: Use of external databases requires [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing).
