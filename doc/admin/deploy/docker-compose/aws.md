# Install Sourcegraph on Amazon Web Services (AWS)

This guide will take you through how to deploy Sourcegraph with [Docker Compose](https://docs.docker.com/compose/) to a single EC2 instance on Amazon Web Services (AWS).

<span class="badge badge-note">RECOMMENDED</span> Deploy a Sourcegraph instance with an [AWS AMI](../machine-images/aws-ami.md) or [AWS One-Click](aws-oneclick.md).

---

## Configure

Click **Launch Instance** from the [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home), then fill in the following values for each section:

#### Name and tags

1. Name your instance

#### Application and OS Images

1. Select **Amazon Linux** in the *Quick Start* tab

2. Select **Amazon Linux 2 AMI (HVM), SSD Volume Type** under *Amazon Machine Image (AMI)*

#### Instance type

1. Select an appropriate instance type using our [resource estimator](../resource_estimator.md) as reference

#### Key pair (login)

1. Create a new key pair for your instance, or choose an existing key pair from the drop down list

#### Network settings

1. Click `Edit` in the header to enable **Auto-assign Public IP** 

2. Under **Firewall (security group)** , create or select existing security group with the following settings:

  * Allow SSH traffic from Anywhere
  * Allow HTTPs traffic from the internet
  * Allow HTTP traffic from the internet

> NOTE: If possible, replace the IP address ranges specified with the IPs from which you actually want to allow access.

#### Configure storage

1. Click **Add New Volume** to add an *additional* EBS volume for storing data

2. Click **Advanced** in the header to update the following settings for the new Custom Volume:
  * `Storage Type`: EBS
  * `Device name`: `/dev/sdb`
  * `Volume Type`: `gp3` (General Purpose SSD)
  * `Size (GiB)`: `250GB minimum`
      * Sourcegraph needs at least as much space as all your repositories combined take up
      * Allocating as much disk space as you can upfront minimize the need for [resizing your volume](https://aws.amazon.com/premiumsupport/knowledge-center/expand-root-ebs-linux/) in the future
  * `Delete on Termination`: `No`

#### Advanced details > User Data

Copy and paste the *startup script* below into the **User Data** textbox:

```bash
curl -sfL https://raw.githubusercontent.com/sourcegraph/deploy-sourcegraph-docker/master/docker-compose/scripts/install-aws.sh | bash -s - v4.1.3
```

> NOTE: If you're deploying a production instance, we recommend [forking the deployment configuration repository](./index.md#step-1-fork-the-deployment-repository) to track any customizations you make to the deployment config. If you do so, please add the git-clone URL and revision of your fork to the end of the curl command above in the following format:
> `curl -sfL https://example.sh | bash -s - $FORK_REVISION $FORK_CLONE_URL`
> 
> - `$FORK_CLONE_URL`: The Git clone URL of your deployment repository. If it is a private repository, please check with your code host on how to generate a URL for cloning private repository
> - `$FORK_REVISION`: The revision (branch) in your fork containing the customizations, typically "release"

---

## Deploy

1. Click **Launch Instance** in the *Summary Section* on the right to launch the EC2 node running Sourcegraph.

2. In your web browser, navigate to the public IP address assigned to the EC2 node. (Look for the **IPv4 Public IP** value in your EC2 instance page under the *Description* panel.) It may take a few minutes for the instance to finish initializing before Sourcegraph becomes accessible. 

You can monitor the setup process by SSHing into the instance to run the following diagnostic commands:

```bash
# Follow the status of the startup script
tail -f /var/log/cloud-init-output.log
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

There are two, non-mutually-exclusive ways to back up data:

* [Snapshot the entire `/mnt/docker-data` EBS volume](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html) on an [automatic, scheduled basis](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html).

* <span class="badge badge-note">RECOMMENDED</span> Use [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) instead of the Dockerized PostgreSQL instance included by default. All data from Sourcegraph is derivable from the data stored in this database. Note, however, that it may take awhile to reclone repositories and rebuild indices afresh. If you require a faster restoration process, we recommend also snapshotting the EBS volume.

---

## Other resources

[HTTP and HTTPS/SSL configuration](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2)
[Site Administration Quickstart](../../../admin/how-to/site-admin-quickstart.md)
