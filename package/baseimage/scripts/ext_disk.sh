#!/bin/bash

# from https://groups.google.com/forum/#!topic/packer-tool/74D7csYVmKE

sudo fdisk /dev/xvda <<EEOF
d
n
p
1
1

w
EEOF
exit 0
