+++
title = "cloud-init troubleshooting"
linktitle = "cloud-init"
+++

[cloud-init](https://cloudinit.readthedocs.org/en/latest/) is an
initialization system used by many cloud providers to run scripts upon
launching new cloud VMs.

# Check the logs

Check the following log files to see if errors occurred while running
the user data installation script:

* `/var/log/cloud-init.log`
* `/var/log/cloud-init-output.log`

Include the contents of these log files when reporting any
installation issues.

