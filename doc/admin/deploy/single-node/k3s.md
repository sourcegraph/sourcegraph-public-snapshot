---
title: Install Sourcegraph locally with K3s
---

# Install Sourcegraph with K3s

This guide will take you through how to set up a Sourcegraph instance locally with [K3s](https://k3s.io/), a tool that lets you run a single-node Kubernetes cluster on your local machine, where we will deploy our Sourcegraph instance to using Sourcegraph Helm Charts.

## Prerequisites

Following are the prerequisites for running Sourcegraph with [K3s](https://k3s.io/) on your Linux machine.

- Ubuntu 18.04 or above
- Minimum of **8 CPU** and **32GB memory** available

The scripts below will install the following on your machine:

- [K3s](https://k3s.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [sourcegraph/deploy repository](https://github.com/sourcegraph/deploy)

## Deploy

Run the following scripts:

##### Start up script

```bash
#!/usr/bin/env bash
set -exuo pipefail

###############################################################################
# ACTION REQUIRED IF RUNNING THIS SCRIPT MANUALLY
# IMPORTANT: Keep this commented when building with the packer pipeline
###############################################################################
INSTANCE_VERSION="4.1.2"  # e.g. 4.0.0
INSTANCE_SIZE="XS"        # e.g. XS / S / M / L / XL

##################### NO CHANGES REQUIRED BELOW THIS LINE #####################
# Variables
###############################################################################
SOURCEGRAPH_VERSION=$INSTANCE_VERSION
SOURCEGRAPH_SIZE=$INSTANCE_SIZE
INSTANCE_USERNAME=$(whoami)
VOLUME_DEVICE_NAME='/dev/nvme1n1'
SOURCEGRAPH_DEPLOY_REPO_URL='https://github.com/sourcegraph/deploy.git'
DEPLOY_PATH="/home/$INSTANCE_USERNAME/deploy/install"
KUBECONFIG_FILE='/etc/rancher/k3s/k3s.yaml'

###############################################################################
# Prepare the system
###############################################################################
# Install git
sudo yum update -y
sudo yum install git -y

# Clone the deployment repository
cd /home/$INSTANCE_USERNAME/
git clone $SOURCEGRAPH_DEPLOY_REPO_URL
cd $DEPLOY_PATH
cp override."$SOURCEGRAPH_SIZE".yaml override.yaml

###############################################################################
# Configure data volume
###############################################################################
# Format (if necessary) and mount the volume
device_fs=$(lsblk $VOLUME_DEVICE_NAME --noheadings --output fsType)
if [ "$device_fs" == "" ]; then
    sudo mkfs -t xfs $VOLUME_DEVICE_NAME
    sudo xfs_admin -L /mnt/data $VOLUME_DEVICE_NAME
fi
sudo mkdir -p /mnt/data
sudo mount $VOLUME_DEVICE_NAME /mnt/data

# Mount data disk on reboots by linking disk label to data root path
sudo sh -c 'echo "LABEL=/mnt/data  /mnt/data  xfs  defaults,nofail  0  2" >> /etc/fstab'
sudo umount /mnt/data
sudo mount -a

###############################################################################
# Kernel parameters required by Sourcegraph
###############################################################################
# These must be set in order for Zoekt (Sourcegraph's search indexing backend)
# to perform at scale without running into limitations.
sudo sh -c "echo 'fs.inotify.max_user_watches=128000' >> /etc/sysctl.conf"
sudo sh -c "echo 'vm.max_map_count=300000' >> /etc/sysctl.conf"
sudo sysctl --system # Reload configuration (no restart required.)

sudo sh -c "echo '* soft nproc 8192' >> /etc/security/limits.conf"
sudo sh -c "echo '* hard nproc 16384' >> /etc/security/limits.conf"
sudo sh -c "echo '* soft nofile 262144' >> /etc/security/limits.conf"
sudo sh -c "echo '* hard nofile 262144' >> /etc/security/limits.conf"

###############################################################################
# Configure network and volumes for k3s
###############################################################################
# Ensure k3s cluster networking/DNS is allowed in local firewall.
# For details see: https://github.com/k3s-io/k3s/issues/24#issuecomment-469759329
sudo yum install iptables-services -y
sudo systemctl enable iptables
sudo systemctl start iptables
sudo iptables -I INPUT 1 -i cni0 -s 10.42.0.0/16 -j ACCEPT
sudo service iptables save

# Put ephemeral kubelet/pod storage in our data disk (since it is the only large disk we have.)
sudo mkdir -p /mnt/data/kubelet /var/lib/kubelet
sudo sh -c 'echo "/mnt/data/kubelet    /var/lib/kubelet    none    bind" >> /etc/fstab'
sudo mount -a

# Put persistent volume pod storage in our data disk, and k3s's embedded database there too (it
# must be kept around in order for k3s to keep PVs attached to the right folder on disk if a node
# is lost (i.e. during an upgrade of Sourcegraph), see https://github.com/rancher/local-path-provisioner/issues/26
sudo mkdir -p /mnt/data/db
sudo mkdir -p /var/lib/rancher/k3s/server
sudo ln -s /mnt/data/db /var/lib/rancher/k3s/server/db
sudo mkdir -p /mnt/data/storage
sudo mkdir -p /var/lib/rancher/k3s
sudo ln -s /mnt/data/storage /var/lib/rancher/k3s/storage

###############################################################################
# Install k3s (Kubernetes single-machine deployment)
###############################################################################
curl -sfL https://get.k3s.io | K3S_TOKEN=none sh -s - \
    --node-name sourcegraph-0 \
    --write-kubeconfig-mode 644 \
    --cluster-cidr 10.10.0.0/16 \
    --kubelet-arg containerd=/run/k3s/containerd/containerd.sock \
    --etcd-expose-metrics true

# Confirm k3s and kubectl are up and running
sleep 5 && k3s kubectl get node

# Correct permissions of k3s config file
sudo chown $INSTANCE_USERNAME /etc/rancher/k3s/k3s.yaml
sudo chmod go-r /etc/rancher/k3s/k3s.yaml

# Set KUBECONFIG to point to k3s for 'kubectl' commands to work
export KUBECONFIG='/etc/rancher/k3s/k3s.yaml'
cp /etc/rancher/k3s/k3s.yaml /home/$INSTANCE_USERNAME/.kube/config

# Add standard bash aliases
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' | tee -a /home/$INSTANCE_USERNAME/.bash_profile
###############################################################################
# Set up Sourcegraph using Helm
###############################################################################
# Install Helm
curl -sSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
helm version --short

# Create override configMap for prometheus before startup Sourcegraph
helm --kubeconfig $KUBECONFIG_FILE repo add sourcegraph https://helm.sourcegraph.com/release
kubectl --kubeconfig $KUBECONFIG_FILE apply -f ./prometheus-override.ConfigMap.yaml
helm --kubeconfig $KUBECONFIG_FILE upgrade -i -f ./override.yaml --version "$SOURCEGRAPH_VERSION" sourcegraph sourcegraph/sourcegraph
kubectl --kubeconfig $KUBECONFIG_FILE create -f $DEPLOY_PATH/ingress.yaml
```

## Upgrade

Please refer to the [upgrade docs for all Sourcegraph Helm instances](../kubernetes/operations.md#upgrade).

## Downgrade

See instructions for upgrades.

## Uninstall

See the [official K3s docs](https://docs.k3s.io/installation/uninstall) for detailed instructions on uninstalling K3s.
