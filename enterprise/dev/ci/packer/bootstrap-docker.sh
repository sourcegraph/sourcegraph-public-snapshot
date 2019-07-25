#!/usr/bin/env bash

set -ex

sudo yum update -y

sudo amazon-linux-extras install docker

sudo sed -i -e 's/1024/10240/g' /etc/sysconfig/docker
sudo sed -i -e 's/4096/40960/g' /etc/sysconfig/docker
sudo usermod -a -G docker ec2-user

sudo systemctl enable --now --no-block docker
