#!/bin/bash

# This script is uploaded to the AMI by packer to provision the
# environment. It will install all base dependencies for src in production.

set -ex
export DEBIAN_FRONTEND=noninteractive

packages=(
    # src deps
    git
    mercurial

    # infra deps
    nfs-common

    # debug tools
    atop
    htop
    iotop
    linux-tools-common
    tcpdump
    mg
    postgresql-client
)
external_packages=(
    https://s3-us-west-2.amazonaws.com/sourcegraph-deb/docker-gc_0.0.3_all.deb
    https://s3-us-west-2.amazonaws.com/sourcegraph-deb/node_exporter-0.11.0.deb
    https://github.com/papertrail/remote_syslog2/releases/download/v0.14-beta-pkgs/remote-syslog2_0.14_amd64.deb
)
declare -A files=(
    [atop]=/etc/default
    [log_files.yml]=/etc
    [papertrail-dynamic-join-group]=/usr/bin
    [sourcegraph-internal-ca.crt]=/usr/local/share/ca-certificates
)

# This is recommended by the packer docs, I guess to let the system settle
# before mutating it
sleep 30

# We want to use the latest stable git
sudo add-apt-repository -y ppa:git-core/ppa

# Update all deps
sudo apt-get update -qq
sudo unattended-upgrade

# Debs provided by ubuntu
sudo apt-get install -qy ${packages[@]}

# Get latest version of docker
wget -qO- https://get.docker.com/ | sudo sh

# Get srclib toolchains (in parallel because docker pull is slow)
#
# Download any old version of src so we can run "src toolchain get".
wget -qO- https://sourcegraph-release.s3.amazonaws.com/src/0.5.105/linux-amd64/src.gz | gunzip > /tmp/old-src && chmod +x /tmp/old-src
# Now pull the Docker images and install the toolchains from source
# (so that their dirs in SRCLIBPATH=/opt/srclib exist).
sudo mkdir -p /opt/srclib
sudo chown -R `whoami` /opt/srclib
echo -n python ruby javascript go java | xargs -d ' ' -n 1 -P 10 -I LANG bash -c 'sudo docker pull sqs1/srclib-LANG:latest && sudo docker tag -f sqs1/srclib-LANG:latest sourcegraph.com-sourcegraph-srclib-LANG && SRCLIBPATH=/opt/srclib /tmp/old-src toolchain get sourcegraph.com/sourcegraph/srclib-LANG'

# Other debs which are not provided by ubuntu
wget ${external_packages[@]}
sudo dpkg -i $(basename -a ${external_packages[@]})
rm -f $(basename -a ${external_packages[@]})

# Put uploaded files in place
for file in "${!files[@]}"; do
    dest=${files[$file]}
    sudo mv $file $dest
    sudo chown root:root $dest/$file
done

# Trust Sourcegraph internal CA (sourcegraph-internal-ca.crt was
# previously uploaded).
sudo update-ca-certificates

# Speed up initial ssh logins by skipping computation of system stats.
sudo apt-get remove --purge -y landscape-common
