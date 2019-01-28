# Install Sourcegraph with Docker on AWS

<style>
div.alert-info {
    background-color: rgb(221, 241, 255);
    border-radius: 0.5em;
    padding: 0.25em 1em 0.25em 1em;
}
</style>

This tutorial shows you how to deploy Sourcegraph to a single node running on AWS.

If you're just starting out, we recommend [installing Sourcegraph locally](index.md). It takes only a few minutes and lets you try out all of the features. If you need scalability and high-availability beyond what a single-server deployment can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph).

---

## Deploy to EC2

### Option A: use the AWS wizard

- Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home).
- Select any Amazon Linux (the first option, `ami-824c4ee2`, works fine).
- Select an appropriate instance size (we recommend t2.medium/large, depending on team size and number of repositories/languages enabled), then **Next: Configure Instance Details**
- Add the following user data (as text) in the **Advanced Details** section:

  ```
  #cloud-config
  repo_update: true
  repo_upgrade: all

  packages:
  - docker

  runcmd:
  - mkdir -p /home/ec2-user/.sourcegraph/config
  - mkdir -p /home/ec2-user/.sourcegraph/data
  - sed -i -e 's/1024/10240/g' /etc/sysconfig/docker
  - sed -i -e 's/4096/40960/g' /etc/sysconfig/docker
  - service docker start
  - usermod -a -G docker ec2-user
  - [ sh, -c, 'docker run -d --publish 80:7080 --publish 443:7443 --publish 2633:2633 --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:3.0.0-beta.3' ]
  ```

- Select **Next: ...** until you get to the **Configure Security Group** page, then add the default **HTTP** rule (port range "80", source "0.0.0.0/0, ::/0")
- Launch your instance, then navigate to the its public URL.
- If you have configured a DNS entry for the IP, configure `externalURL` to reflect that. (Note: `externalURL` was called `appURL` in Sourcegraph 2.13 and earlier.)

### Option B: use the CLI

Use the [`aws` CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html) to boot an EC2 instance running Sourcegraph.

First, create a `cloud-init.txt` file with user data contents as shown above or below. Then run:

`aws ec2 run-instances --image-id ami-824c4ee2 --count 1 --instance-type t2.small --key-name id_rsa --security-groups default --user-data file://cloud-init.txt`

Substitute the path to your `cloud-init.txt` file, the name of your key pair, and an appropriate security group. To start you probably want a security group which exposes port 80 and 22 (for SSH) to the public internet.

---

## Update your Sourcegraph version

To update to the most recent version of Sourcegraph (X.Y.Z), SSH into your instance and run the following:

```
docker ps # get the $CONTAINER_ID of the running sourcegraph/server container
docker rm -f $CONTAINER_ID
docker run -d ... sourcegraph/server:X.Y.Z
```

---

## Using an external database for persistence

The Docker container has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the container, you can [use external databases](../../external_database.md) for persistence, such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) and [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/).

The [site configuration JSON](../../site_config/index.md) is not yet stored in the database, so you must manually back it up. This will no longer be necessary in [Sourcegraph 3.0](https://github.com/sourcegraph/about/pull/36). <!-- TODO: remove this when https://github.com/sourcegraph/about/pull/36 is merged -->

> NOTE: Use of external databases requires [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing).
