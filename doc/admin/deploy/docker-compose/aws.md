# Install Sourcegraph with Docker Compose on AWS

This guide shows you how to deploy Sourcegraph via [Docker Compose](https://docs.docker.com/compose/) to a single EC2 instance on AWS.

> NOTE: Trying to decide how to deploy Sourcegraph? See [our recommendations](../index.md) for how to choose a deployment type that suits your needs.

## Determine server and service requirements 

Use the [resource estimator](../resource_estimator.md) to determine the resource requirements for your environment. You will use this information to set up the instance and configure the docker-compose YAML file. 

## Prepare a fork 

We strongly recommend that you create and run Sourcegraph from your own fork of the reference repository. You will make changes to the default configuration, for example to the docker-compose YAML file, in your fork. The fork will also enable you to keep track of your customizations when upgrading your fork from the reference repo. Refer to the following steps for preparing a clone, which use GitHub as an example, then return to this page:

1. [Fork the reference repo](index.md#step-1-fork-the-sourcegraph-docker-compose-deployment-repository)
2. [Clone your fork](index.md#step-2-clone-the-forked-repository-locally)
3. [Configure the release branch](index.md#step-3-configure-the-release-branch)
4. [Configure the YAML file](index.md#step-4-configure-the-yaml-file)
5. [Publish changes to your branch](index.md#step-5-update-your-release-branch)

## Deploy to EC2

* Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home).
* Select the **Amazon Linux 2 AMI (HVM), SSD Volume Type**.
* Select an appropriate instance size (use the [resource estimator](../resource_estimator.md) to find a good starting point for your deployment), then click **Next: Configure Instance Details.**
* Ensure the **Auto-assign Public IP** option is set to "Enable". This ensures your instance is accessible to the Internet.
* Place the following script in the **User Data** text box at the bottom of the **Configure Instance Details** page

![Screen Shot 2021-12-28 at 1 05 07 PM](https://user-images.githubusercontent.com/13024338/147607360-5b76e122-479d-44aa-9e71-0b282cbc243a.png)

> WARNING: If working from a fork of the reference repository, update the following variables in the script:
> 
> * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL`: Your fork's git clone URL
> * `DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION`: The git revision containing your fork's customizations to the base Sourcegraph Docker Compose YAML. In the [example](index.md#configure-a-release-branch) the revision is the `release` branch. 

```bash
#!/usr/bin/env bash

set -euxo pipefail

EBS_VOLUME_DEVICE_NAME='/dev/sdb'
DOCKER_DATA_ROOT='/mnt/docker-data'

DOCKER_COMPOSE_VERSION='1.29.2'
DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT='/home/ec2-user/deploy-sourcegraph-docker'

# 🚨 Update these variables with the correct values from your fork!
DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL='https://github.com/sourcegraph/deploy-sourcegraph-docker.git'
DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v3.43.1'

# Install git
yum update -y
yum install git -y

# Clone Docker Compose definition
git clone "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_CLONE_URL}" "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"
git checkout "${DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION}"

# Format (if necessary) and mount EBS volume
device_fs=$(lsblk "${EBS_VOLUME_DEVICE_NAME}" --noheadings --output fsType)
if [ "${device_fs}" == "" ] ## only format the volume if it isn't already formatted
then
  mkfs -t xfs "${EBS_VOLUME_DEVICE_NAME}"
fi
mkdir -p "${DOCKER_DATA_ROOT}"
mount "${EBS_VOLUME_DEVICE_NAME}" "${DOCKER_DATA_ROOT}"

# Mount EBS volume on reboots
EBS_UUID=$(blkid -s UUID -o value "${EBS_VOLUME_DEVICE_NAME}")
echo "UUID=${EBS_UUID}  ${DOCKER_DATA_ROOT}  xfs  defaults,nofail  0  2" >> '/etc/fstab'
umount "${DOCKER_DATA_ROOT}"
mount -a

# Install, configure, and enable Docker
yum update -y
amazon-linux-extras install docker
systemctl enable --now docker
sed -i -e 's/1024/262144/g' /etc/sysconfig/docker
sed -i -e 's/4096/262144/g' /etc/sysconfig/docker
usermod -a -G docker ec2-user

# Install jq for scripting
yum install -y jq

# Edit Docker storage directory to mounted volume
DOCKER_DAEMON_CONFIG_FILE='/etc/docker/daemon.json'

## initialize the config file with empty json if it doesn't exist
if [ ! -f "${DOCKER_DAEMON_CONFIG_FILE}" ]
then
  mkdir -p $(dirname "${DOCKER_DAEMON_CONFIG_FILE}")
  echo '{}' > "${DOCKER_DAEMON_CONFIG_FILE}"
fi

## update Docker's 'data-root' to point to our mounted disk
tmp_config=$(mktemp)
trap "rm -f ${tmp_config}" EXIT
cat "${DOCKER_DAEMON_CONFIG_FILE}" | jq --arg DATA_ROOT "${DOCKER_DATA_ROOT}" '.["data-root"]=$DATA_ROOT' > "${tmp_config}"
cat "${tmp_config}" > "${DOCKER_DAEMON_CONFIG_FILE}"

## finally, restart Docker daemon to pick up our changes
systemctl restart --now docker

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
curl -L "https://raw.githubusercontent.com/docker/compose/${DOCKER_COMPOSE_VERSION}/contrib/completion/bash/docker-compose" -o /etc/bash_completion.d/docker-compose

# Run Sourcegraph. Restart the containers upon reboot.
cd "${DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT}"/docker-compose
docker-compose up -d
```

* Select **Next: Add Storage**
* Click "Add New Volume" and add an additional volume (for storing Docker data) with the following settings:

  * **Volume Type** (left-most column): EBS
  * **IMPORTANT: Device**: `/dev/sdb`
  * **Size (GiB)**: `250` GB minimum *(As a rule of thumb, Sourcegraph needs at least as much space as all your repositories combined take up. Allocating as much disk space as you can upfront helps you avoid [resizing your volume](https://aws.amazon.com/premiumsupport/knowledge-center/expand-root-ebs-linux/) later on.)*
  * **Volume Type**: General Purpose SSD (gp2)
  * **Delete on Termination**: Leave this setting unchecked

* Select **Next: ...** until you get to the **Configure Security Group** page. Then add the following rules:
  * Default **HTTP** rule: port range `80`, source `0.0.0.0/0, ::/0`
  * Default **HTTPS** rule: port range `443`, source `0.0.0.0/0, ::/0`<br>(NOTE: additional work will be required later on to [configure SSL in the Docker Compose deployment](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2))

> WARNING: While this port configuration will work, it provides open access of the ports specified. If possible, replace the IP address ranges specified with the IPs from which you actually want to allow access.

* Launch your instance, then navigate to its public IP in your browser. (This can be found by navigating to the instance page on EC2 and looking in the "Description" panel for the "IPv4 Public IP" value.) You may have to wait a minute or two for the instance to finish initializing before Sourcegraph becomes accessible. You can monitor the status by SSHing into the instance and using the diagnostic commands:

```bash
# Follow the status of the user data script you provided earlier
tail -f /var/log/cloud-init-output.log

# (Once the user data script completes) monitor the health of the "sourcegraph-frontend" container
docker ps --filter="name=sourcegraph-frontend-0"
```

---

## Update your Sourcegraph version

Refer to the [Docker Compose upgrade docs](upgrade.md).

## Storage and Backups

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. The previous script [configures Docker](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file) to store all Docker data on the additional EBS volume that was attached to the instance (mounted at `/mnt/docker-data` - the volumes themselves are stored under `/mnt/docker-data/volumes`) There are a few different ways to backup this data:

* (**recommended**) The most straightforward method to backup this data is to [snapshot the entire `/mnt/docker-data` EBS disk](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html) on an [automatic, scheduled basis](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html).

* Using an external Postgres instance lets a service such as [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/) take care of backing up all of Sourcegraph's user data for you. If the EC2 instance running Sourcegraph ever dies or is destroyed, creating a fresh instance that's connected to that external Postgres will leave Sourcegraph in the same state that it was before.
