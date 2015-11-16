+++
title = "Installing on Ubuntu Linux"
navtitle = "on Ubuntu Linux"
+++

Note: If you are installing a team server, consider following one of the
cloud provider installation instructions instead.

Sourcegraph is supported on Ubuntu 12.04 and 14.04. Install using one
of the following methods:

* `wget -O - https://sourcegraph.com/.download/install.sh | bash`
* Download [src.deb (64-bit)](https://sourcegraph.com/.download/latest/linux-amd64/src.deb) and install with `sudo dpkg -i src.deb`

Next, edit `/etc/sourcegraph/config.ini`. When done, run `sudo restart
src` and visit [http://localhost:3080](http://localhost:3080) (or an
alternate address, if you modified the configuration).

# Next steps

* [Add language support]({{< relref "config/toolchains.md" >}})
* [Getting started with Sourcegraph for your team]({{< relref "getting-started/index.md" >}})
